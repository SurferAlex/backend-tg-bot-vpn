package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"tg-bot/internal/api"
	"tg-bot/internal/botapp"
	"tg-bot/internal/config"
	"tg-bot/internal/dedup"
	"tg-bot/internal/telegram"
	"tg-bot/vpnapi"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var bot *tgbotapi.BotAPI
	for {
		bot, err = tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
		if err == nil {
			break
		}
		log.Printf("Error creating bot (will retry): %v", err)

		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
		}
	}

	log.Printf("authorized as @%s", bot.Self.UserName)

	notifier := &telegram.Notifier{Bot: bot, AdminIDs: cfg.Telegram.AdminIDs}

	d := dedup.New(10 * time.Minute)
	vpn := vpnapi.New(cfg.VPNAPI.BaseURL, cfg.VPNAPI.InternalToken)
	r := api.SetupServer(notifier, cfg.Alert.InternalToken, d, vpn)

	srv := &http.Server{
		Addr:              cfg.Alert.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		bot.StopReceivingUpdates()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown: %v", err)
		}
	}()

	go func() {
		log.Printf("alert http server started on %s", cfg.Alert.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http server error: %v", err)
		}
	}()

	isAdmin := func(id int64) bool {
		for _, a := range cfg.Telegram.AdminIDs {
			if id == a {
				return true
			}
		}
		return false
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	const (
		defaultMaxIPs = 2
	)

	type pendingNew struct {
		days      int
		createdAt time.Time
		chatID    int64
	}
	pending := map[int64]pendingNew{}

	type pendingActionKind string

	const (
		actionGet       pendingActionKind = "get"
		actionProvision pendingActionKind = "provision"
		actionRevoke    pendingActionKind = "revoke"
	)

	type pendingAction struct {
		kind      pendingActionKind
		createdAt time.Time
		chatID    int64
	}

	pendingActionByUser := map[int64]pendingAction{}

	newKeyMarkup := botapp.NewKeyMarkup()
	mainMenuMarkup := botapp.MainMenuMarkup()
	cancelMarkup := botapp.CancelMarkup()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		from := update.Message.From
		if from == nil || !isAdmin(from.ID) {
			continue
		}

		sendMenu := func(chatID int64) {
			msg := tgbotapi.NewMessage(chatID, "Меню:")
			msg.ReplyMarkup = mainMenuMarkup
			_, _ = bot.Send(msg)
		}

		// Global navigation buttons
		text := strings.TrimSpace(update.Message.Text)
		if text == botapp.BtnCancel {
			delete(pendingActionByUser, from.ID)
			delete(pending, from.ID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ок, отменено.")
			msg.ReplyMarkup = mainMenuMarkup
			_, _ = bot.Send(msg)
			continue
		}
		if text == botapp.BtnBack {
			sendMenu(update.Message.Chat.ID)
			continue
		}

		if !update.Message.IsCommand() {
			if pa, ok := pendingActionByUser[from.ID]; ok {
				if time.Since(pa.createdAt) > 5*time.Minute {
					delete(pendingActionByUser, from.ID)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Сессия истекла. Откройте меню ещё раз.")
					msg.ReplyMarkup = mainMenuMarkup
					_, _ = bot.Send(msg)
					continue
				}

				uuidArg := strings.TrimSpace(update.Message.Text)
				if uuidArg == "" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "UUID пустой. Введите client_uuid или нажмите «Отмена».")
					msg.ReplyMarkup = cancelMarkup
					_, _ = bot.Send(msg)
					continue
				}

				delete(pendingActionByUser, from.ID)

				switch pa.kind {
				case actionGet:
					a, err := vpn.GetAccess(ctx, uuidArg)
					if err != nil {
						log.Printf("get access failed (clientUuid=%s): %v", uuidArg, err)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить доступ. Проверьте UUID и попробуйте ещё раз.")
						msg.ReplyMarkup = mainMenuMarkup
						_, _ = bot.Send(msg)
						continue
					}
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, a.VLESSURI))

				case actionProvision:
					a, err := vpn.Provision(ctx, uuidArg)
					if err != nil {
						log.Printf("provision failed (clientUuid=%s): %v", uuidArg, err)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось выдать доступ. Проверьте UUID и попробуйте ещё раз.")
						msg.ReplyMarkup = mainMenuMarkup
						_, _ = bot.Send(msg)
						continue
					}
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, a.VLESSURI))

				case actionRevoke:
					if err := vpn.Revoke(ctx, uuidArg); err != nil {
						log.Printf("revoke failed (clientUuid=%s): %v", uuidArg, err)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось отозвать ключ. Проверьте UUID и попробуйте ещё раз.")
						msg.ReplyMarkup = mainMenuMarkup
						_, _ = bot.Send(msg)
						continue
					}
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ключ успешно отозван."))
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню:")
				msg.ReplyMarkup = mainMenuMarkup
				_, _ = bot.Send(msg)
				continue
			}

			// Create key: choose days
			switch text {
			case botapp.BtnDays30, botapp.BtnDays60, botapp.BtnDays90:
				var days int
				switch text {
				case botapp.BtnDays30:
					days = 30
				case botapp.BtnDays60:
					days = 60
				case botapp.BtnDays90:
					days = 90
				}
				pending[from.ID] = pendingNew{days: days, createdAt: time.Now(), chatID: update.Message.Chat.ID}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ок. Введите имя клиента для ключа на %d дней (или нажмите «Отмена»).", days))
				msg.ReplyMarkup = cancelMarkup
				_, _ = bot.Send(msg)
				continue
			}

			if p, ok := pending[from.ID]; ok {
				if time.Since(p.createdAt) > 5*time.Minute {
					delete(pending, from.ID)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сессия создания истекла. Нажмите /new ещё раз."))
					continue
				}

				name := strings.TrimSpace(update.Message.Text)
				if name == "" {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Имя пустое. Введите имя клиента или /cancel."))
					continue
				}

				if len([]rune(name)) > 64 {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Имя слишком длинное (макс 64 символа)."))
					continue
				}

				delete(pending, from.ID)

				ttlSeconds := int64(p.days) * 24 * 60 * 60
				note := name

				client, err := vpn.CreateClient(ctx, vpnapi.CreateClientRequest{
					TelegramUserID: nil,
					MaxIPs:         defaultMaxIPs,
					TTLSeconds:     ttlSeconds,
					Note:           &note,
				})
				if err != nil {
					log.Printf("create client failed (name=%q, days=%d): %v", name, p.days, err)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось создать клиента. Попробуйте ещё раз позже."))
					continue
				}
				a, err := vpn.Provision(ctx, client.ClientUUID)
				if err != nil {
					log.Printf("provision failed (clientUuid=%s): %v", client.ClientUUID, err)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось выдать доступ. Попробуйте ещё раз позже."))
					continue
				}

				info := fmt.Sprintf("Имя: %s\nСрок: %d дней\nВнутренний UUID (для /revoke): %s", name, p.days, client.ClientUUID)
				if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, info)); err != nil {
					log.Printf("send info message failed: %v", err)
				}
				time.Sleep(300 * time.Millisecond)

				if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, a.VLESSURI)); err != nil {
					log.Printf("send vless uri failed: %v", err)
				}
				time.Sleep(300 * time.Millisecond)

				if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, client.ClientUUID)); err != nil {
					log.Printf("send uuid failed: %v", err)
				}
				continue
			}

			// Main menu actions
			switch text {
			case botapp.BtnCreate:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери срок:")
				msg.ReplyMarkup = newKeyMarkup
				_, _ = bot.Send(msg)
				continue
			case botapp.BtnAccess:
				pendingActionByUser[from.ID] = pendingAction{kind: actionGet, createdAt: time.Now(), chatID: update.Message.Chat.ID}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите client_uuid (или нажмите «Отмена»):")
				msg.ReplyMarkup = cancelMarkup
				_, _ = bot.Send(msg)
				continue
			case botapp.BtnProvision:
				pendingActionByUser[from.ID] = pendingAction{kind: actionProvision, createdAt: time.Now(), chatID: update.Message.Chat.ID}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите client_uuid для provision (или нажмите «Отмена»):")
				msg.ReplyMarkup = cancelMarkup
				_, _ = bot.Send(msg)
				continue
			case botapp.BtnRevoke:
				pendingActionByUser[from.ID] = pendingAction{kind: actionRevoke, createdAt: time.Now(), chatID: update.Message.Chat.ID}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите client_uuid для revoke (или нажмите «Отмена»):")
				msg.ReplyMarkup = cancelMarkup
				_, _ = bot.Send(msg)
				continue
			}
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				sendMenu(update.Message.Chat.ID)
			case "menu":
				sendMenu(update.Message.Chat.ID)
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Команды: /start, /help, /new, /cancel, /get, /provision, /revoke")
				_, _ = bot.Send(msg)
			case "new":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери срок:")
				msg.ReplyMarkup = newKeyMarkup
				_, _ = bot.Send(msg)
			case "cancel":
				if _, ok := pending[from.ID]; ok {
					delete(pending, from.ID)
					delete(pendingActionByUser, from.ID)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отменено.")
					msg.ReplyMarkup = mainMenuMarkup
					_, _ = bot.Send(msg)
				} else {
					if _, ok := pendingActionByUser[from.ID]; ok {
						delete(pendingActionByUser, from.ID)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отменено.")
						msg.ReplyMarkup = mainMenuMarkup
						_, _ = bot.Send(msg)
						continue
					}
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Нечего отменять."))
				}
			case "get":
				uuidArg := strings.TrimSpace(update.Message.CommandArguments())
				if uuidArg == "" {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Использование: /get <client_uuid>"))
					continue
				}
				a, err := vpn.GetAccess(ctx, uuidArg)
				if err != nil {
					log.Printf("get access failed (clientUuid=%s): %v", uuidArg, err)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить доступ. Проверьте UUID и попробуйте ещё раз."))
					continue
				}
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, a.VLESSURI))
			case "provision":
				uuidArg := strings.TrimSpace(update.Message.CommandArguments())
				if uuidArg == "" {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Использование: /provision <client_uuid>"))
					continue
				}
				a, err := vpn.Provision(ctx, uuidArg)
				if err != nil {
					log.Printf("provision failed (clientUuid=%s): %v", uuidArg, err)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось выдать доступ. Проверьте UUID и попробуйте ещё раз."))
					continue
				}
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, a.VLESSURI))
			case "revoke":
				uuidArg := strings.TrimSpace(update.Message.CommandArguments())
				if uuidArg == "" {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Использование: /revoke <client_uuid>"))
					continue
				}
				if err := vpn.Revoke(ctx, uuidArg); err != nil {
					log.Printf("revoke failed (clientUuid=%s): %v", uuidArg, err)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось отозвать ключ. Проверьте UUID и попробуйте ещё раз."))
					continue
				}
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ключ успешно отозван."))
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда.")
				_, _ = bot.Send(msg)
			}
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я получил ваше сообщение.")
		_, _ = bot.Send(msg)
	}
}
