package main

import (
	"context"
	"errors"
	"fmt"
	"html"
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

	type pendingNew struct {
		serverID   string
		serverName string
		days       int
		maxIPs     int
		createdAt  time.Time
		chatID     int64
	}
	pending := map[int64]pendingNew{}

	resolveClientRef := func(ref string) (string, error) {
		client, err := vpn.ResolveClient(ctx, ref)
		if err != nil {
			return "", err
		}
		return client.ClientUUID, nil
	}

	listServerLabels := func() ([]botapp.ServerLabel, error) {
		servers, err := vpn.ListServers(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]botapp.ServerLabel, 0, len(servers))
		for _, s := range servers {
			out = append(out, botapp.ServerLabel{ID: s.ID, Name: s.Name})
		}
		return out, nil
	}

	startCreateFlow := func(chatID, userID int64) {
		labels, err := listServerLabels()
		if err != nil {
			log.Printf("list servers failed: %v", err)
			_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Не удалось загрузить список серверов. Попробуйте позже."))
			return
		}
		if len(labels) == 0 {
			_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Нет доступных VPN-серверов. Добавьте сервер в VpnAPI (таблица vpn_servers)."))
			return
		}
		pending[userID] = pendingNew{createdAt: time.Now(), chatID: chatID}
		msg := tgbotapi.NewMessage(chatID, "Выбери сервер:")
		msg.ReplyMarkup = botapp.ServersMarkup(labels)
		_, _ = bot.Send(msg)
	}

	type pendingActionKind string

	const (
		actionGet       pendingActionKind = "get"
		actionProvision pendingActionKind = "provision"
		actionRevoke    pendingActionKind = "revoke"
		actionExtend    pendingActionKind = "extend"
		actionMaxIPs    pendingActionKind = "max_ips"
	)

	type pendingAction struct {
		kind       pendingActionKind
		clientUUID string
		createdAt  time.Time
		chatID     int64
	}

	pendingActionByUser := map[int64]pendingAction{}

	newKeyMarkup := botapp.NewKeyMarkup()
	maxIPsMarkup := botapp.MaxIPsMarkup()
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
			// Menu before create-flow state: stale «Создать ключ» must not swallow other buttons.
			if botapp.IsMainMenuButton(text) {
				delete(pending, from.ID)
				delete(pendingActionByUser, from.ID)
				switch text {
				case botapp.BtnCreate:
					startCreateFlow(update.Message.Chat.ID, from.ID)
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
				case botapp.BtnExtend:
					pendingActionByUser[from.ID] = pendingAction{kind: actionExtend, createdAt: time.Now(), chatID: update.Message.Chat.ID}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите имя клиента или UUID (как при создании ключа):")
					msg.ReplyMarkup = cancelMarkup
					_, _ = bot.Send(msg)
					continue
				case botapp.BtnMaxIPs:
					pendingActionByUser[from.ID] = pendingAction{kind: actionMaxIPs, createdAt: time.Now(), chatID: update.Message.Chat.ID}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите имя клиента или UUID:")
					msg.ReplyMarkup = cancelMarkup
					_, _ = bot.Send(msg)
					continue
				}
			}

			if pa, ok := pendingActionByUser[from.ID]; ok {
				if time.Since(pa.createdAt) > 5*time.Minute {
					delete(pendingActionByUser, from.ID)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Сессия истекла. Откройте меню ещё раз.")
					msg.ReplyMarkup = mainMenuMarkup
					_, _ = bot.Send(msg)
					continue
				}

				finishPendingMenu := func() {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню:")
					msg.ReplyMarkup = mainMenuMarkup
					_, _ = bot.Send(msg)
				}

				if pa.clientUUID != "" && pa.kind == actionExtend {
					var addDays int
					switch text {
					case botapp.BtnDays30:
						addDays = 30
					case botapp.BtnDays60:
						addDays = 60
					case botapp.BtnDays90:
						addDays = 90
					default:
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери срок: 30, 60 или 90 дней.")
						msg.ReplyMarkup = newKeyMarkup
						_, _ = bot.Send(msg)
						continue
					}
					delete(pendingActionByUser, from.ID)
					client, err := vpn.ExtendClient(ctx, pa.clientUUID, addDays)
					if err != nil {
						log.Printf("extend failed (clientUuid=%s, days=%d): %v", pa.clientUUID, addDays, err)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось продлить ключ. Проверьте имя/UUID и что ключ активен.")
						msg.ReplyMarkup = mainMenuMarkup
						_, _ = bot.Send(msg)
						continue
					}
					info := fmt.Sprintf(
						"Ключ продлён на %d дн.\nUUID: <code>%s</code>\nДействует до: %s\nЛимит IP: %d",
						addDays, client.ClientUUID, client.KeyExpiresAt.UTC().Format("2006-01-02 15:04 UTC"), client.MaxIPs,
					)
					infoMsg := tgbotapi.NewMessage(update.Message.Chat.ID, info)
					infoMsg.ParseMode = tgbotapi.ModeHTML
					_, _ = bot.Send(infoMsg)
					finishPendingMenu()
					continue
				}

				if pa.clientUUID != "" && pa.kind == actionMaxIPs {
					maxIPs, ok := botapp.ParseMaxIPsButton(text)
					if !ok {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери лимит IP: от 1 до 6.")
						msg.ReplyMarkup = maxIPsMarkup
						_, _ = bot.Send(msg)
						continue
					}
					delete(pendingActionByUser, from.ID)
					client, err := vpn.UpdateMaxIPs(ctx, pa.clientUUID, maxIPs)
					if err != nil {
						log.Printf("update max ips failed (clientUuid=%s, maxIps=%d): %v", pa.clientUUID, maxIPs, err)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось изменить лимит IP. Проверьте имя/UUID и что ключ активен.")
						msg.ReplyMarkup = mainMenuMarkup
						_, _ = bot.Send(msg)
						continue
					}
					info := fmt.Sprintf(
						"Лимит IP обновлён: %d\nUUID: <code>%s</code>\nДействует до: %s",
						client.MaxIPs, client.ClientUUID, client.KeyExpiresAt.UTC().Format("2006-01-02 15:04 UTC"),
					)
					infoMsg := tgbotapi.NewMessage(update.Message.Chat.ID, info)
					infoMsg.ParseMode = tgbotapi.ModeHTML
					_, _ = bot.Send(infoMsg)
					finishPendingMenu()
					continue
				}

				ref := strings.TrimSpace(update.Message.Text)
				if ref == "" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ввод пустой. Введите имя или UUID (или «Отмена»).")
					msg.ReplyMarkup = cancelMarkup
					_, _ = bot.Send(msg)
					continue
				}

				switch pa.kind {
				case actionExtend, actionMaxIPs:
					clientUUID, err := resolveClientRef(ref)
					if err != nil {
						if errors.Is(err, vpnapi.ErrAmbiguousClient) {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Найдено несколько клиентов с таким именем. Укажите UUID.")
							msg.ReplyMarkup = cancelMarkup
							_, _ = bot.Send(msg)
							continue
						}
						log.Printf("resolve client failed (ref=%q): %v", ref, err)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Клиент не найден. Проверьте имя или UUID.")
						msg.ReplyMarkup = cancelMarkup
						_, _ = bot.Send(msg)
						continue
					}
					pa.clientUUID = clientUUID
					pendingActionByUser[from.ID] = pa
					if pa.kind == actionExtend {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На сколько продлить?")
						msg.ReplyMarkup = newKeyMarkup
						_, _ = bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери новый лимит IP:")
						msg.ReplyMarkup = maxIPsMarkup
						_, _ = bot.Send(msg)
					}
					continue
				}

				uuidArg := ref

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

				finishPendingMenu()
				continue
			}

			if labels, err := listServerLabels(); err == nil {
				if serverID, ok := botapp.ParseServerButton(text, labels); ok {
					if p, exists := pending[from.ID]; exists && p.serverID == "" {
						if time.Since(p.createdAt) > 5*time.Minute {
							delete(pending, from.ID)
							_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сессия создания истекла. Нажмите /new ещё раз."))
							continue
						}
						for _, lb := range labels {
							if lb.ID == serverID {
								p.serverName = lb.Name
								break
							}
						}
						p.serverID = serverID
						pending[from.ID] = p
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Сервер: %s\nВыбери срок:", p.serverName))
						msg.ReplyMarkup = newKeyMarkup
						_, _ = bot.Send(msg)
						continue
					}
				}
			}

			// Create key: choose days
			switch text {
			case botapp.BtnDays30, botapp.BtnDays60, botapp.BtnDays90:
				p, exists := pending[from.ID]
				if !exists || p.serverID == "" {
					startCreateFlow(update.Message.Chat.ID, from.ID)
					continue
				}
				var days int
				switch text {
				case botapp.BtnDays30:
					days = 30
				case botapp.BtnDays60:
					days = 60
				case botapp.BtnDays90:
					days = 90
				}
				if time.Since(p.createdAt) > 5*time.Minute {
					delete(pending, from.ID)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сессия создания истекла. Нажмите /new ещё раз."))
					continue
				}
				p.days = days
				pending[from.ID] = p
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Срок: %d дней. Выбери лимит IP:", days))
				msg.ReplyMarkup = maxIPsMarkup
				_, _ = bot.Send(msg)
				continue
			}

			if maxIPs, ok := botapp.ParseMaxIPsButton(text); ok {
				if p, exists := pending[from.ID]; exists && p.serverID != "" && p.maxIPs == 0 {
					if time.Since(p.createdAt) > 5*time.Minute {
						delete(pending, from.ID)
						_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сессия создания истекла. Нажмите /new ещё раз."))
						continue
					}
					p.maxIPs = maxIPs
					pending[from.ID] = p
					msg := tgbotapi.NewMessage(update.Message.Chat.ID,
						fmt.Sprintf("Лимит: %d IP. Введите имя клиента (или «Отмена»).", maxIPs))
					msg.ReplyMarkup = cancelMarkup
					_, _ = bot.Send(msg)
					continue
				}
			}

			if p, ok := pending[from.ID]; ok {
				if time.Since(p.createdAt) > 5*time.Minute {
					delete(pending, from.ID)
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сессия создания истекла. Нажмите /new ещё раз."))
					continue
				}

				if p.serverID == "" {
					startCreateFlow(update.Message.Chat.ID, from.ID)
					continue
				}
				if p.days == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выбери срок:")
					msg.ReplyMarkup = newKeyMarkup
					_, _ = bot.Send(msg)
					continue
				}
				if p.maxIPs == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выбери лимит IP:")
					msg.ReplyMarkup = maxIPsMarkup
					_, _ = bot.Send(msg)
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
					ServerID:       p.serverID,
					TelegramUserID: nil,
					MaxIPs:         p.maxIPs,
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

				info := fmt.Sprintf(
					"Сервер: %s\nИмя: %s\nСрок: %d дней\nЛимит IP: %d\nВнутренний UUID (для /revoke): <code>%s</code>",
					html.EscapeString(p.serverName), html.EscapeString(name), p.days, p.maxIPs, client.ClientUUID,
				)
				infoMsg := tgbotapi.NewMessage(update.Message.Chat.ID, info)
				infoMsg.ParseMode = tgbotapi.ModeHTML
				if _, err := bot.Send(infoMsg); err != nil {
					log.Printf("send info message failed: %v", err)
				}
				time.Sleep(300 * time.Millisecond)

				if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, a.VLESSURI)); err != nil {
					log.Printf("send vless uri failed: %v", err)
				}
				time.Sleep(300 * time.Millisecond)

				uuidMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "<code>"+client.ClientUUID+"</code>")
				uuidMsg.ParseMode = tgbotapi.ModeHTML
				if _, err := bot.Send(uuidMsg); err != nil {
					log.Printf("send uuid failed: %v", err)
				}
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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Команды: /start, /help, /new, /cancel, /get, /provision, /revoke\nПродление и Limit IP — кнопки в меню (имя или UUID → срок / IP).")
				_, _ = bot.Send(msg)
			case "new":
				startCreateFlow(update.Message.Chat.ID, from.ID)
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
