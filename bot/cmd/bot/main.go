package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"tg-bot/internal/api"
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
		cbNewPrefix   = "new:"
		defaultMaxIPs = 2
	)

	type pendingNew struct {
		days      int
		createdAt time.Time
		chatID    int64
	}
	pending := map[int64]pendingNew{}

	newKeyMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать на 30 дней", cbNewPrefix+"30"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать на 60 дней", cbNewPrefix+"60"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать на 90 дней", cbNewPrefix+"90"),
		),
	)

	for update := range updates {
		if update.CallbackQuery != nil {
			cb := update.CallbackQuery
			from := cb.From
			if from == nil || !isAdmin(from.ID) {
				_, _ = bot.Request(tgbotapi.NewCallback(cb.ID, "not allowed"))
				continue
			}

			data := cb.Data
			if strings.HasPrefix(data, cbNewPrefix) {
				daysStr := strings.TrimPrefix(data, cbNewPrefix)
				days, err := strconv.Atoi(daysStr)
				if err != nil || (days != 30 && days != 60 && days != 90) {
					_, _ = bot.Request(tgbotapi.NewCallback(cb.ID, "bad days"))
					continue
				}

				pending[from.ID] = pendingNew{days: days, createdAt: time.Now(), chatID: cb.Message.Chat.ID}
				_, _ = bot.Request(tgbotapi.NewCallback(cb.ID, "ok"))

				msg := tgbotapi.NewMessage(cb.Message.Chat.ID, fmt.Sprintf("Ок. Введите имя клиента для ключа на %d дней (или /cancel).", days))
				_, _ = bot.Send(msg)
				continue
			}

			_, _ = bot.Request(tgbotapi.NewCallback(cb.ID, "unknown action"))
			continue
		}

		if update.Message == nil {
			continue
		}

		from := update.Message.From
		if from == nil || !isAdmin(from.ID) {
			continue
		}

		if !update.Message.IsCommand() {
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
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ок, я на связи. Это приватный бот.")
				_, _ = bot.Send(msg)
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Команды: /start, /help, /new, /cancel, /get <uuid>, /provision <uuid>, /revoke <uuid>")
				_, _ = bot.Send(msg)
			case "new":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери срок:")
				msg.ReplyMarkup = newKeyMarkup
				_, _ = bot.Send(msg)
			case "cancel":
				if _, ok := pending[from.ID]; ok {
					delete(pending, from.ID)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ок, отменено."))
				} else {
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
