package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"connect-bot/internal/botapp"
	"connect-bot/internal/config"
	"connect-bot/internal/handlers"
	"connect-bot/internal/happ"
	"connect-bot/internal/redirect"
	"connect-bot/internal/userrouting"
	"connect-bot/internal/uservless"

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

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("telegram bot: %v", err)
	}

	log.Printf("connect_bot: authorized as @%s", bot.Self.UserName)

	openRedirect := happ.ResolveOpenRedirectBase(cfg.HappRedirectPublicURL)
	if cfg.HappRedirectPublicURL == "" {
		log.Printf("connect_bot: HAPP_REDIRECT_PUBLIC_URL not set, using default %s", openRedirect)
	} else {
		log.Printf("connect_bot: happ open redirect %s", openRedirect)
	}

	redirectSrv := redirect.NewServer(cfg.HappRedirectListenAddr, cfg.HappRoutingOnAdd, cfg.HappDefaultRoutingB64)
	go func() {
		if err := redirectSrv.Run(ctx); err != nil {
			log.Printf("happ redirect server stopped: %v", err)
		}
	}()

	vpnKeys := handlers.NewVPNKeyHandler(bot, cfg.HappIOSAppStore, cfg.HappRedirectPublicURL, cfg.HappDefaultRoutingB64)
	vlessKeys := &uservless.Store{}
	routingProfiles := &userrouting.Store{}
	menu := botapp.MainMenuMarkup()

	sendWelcome := func(chatID int64) {
		msg := tgbotapi.NewMessage(chatID, botapp.WelcomeText(bot.Self.UserName))
		msg.ReplyMarkup = menu
		if _, err := bot.Send(msg); err != nil {
			log.Printf("send welcome failed (chatId=%d): %v", chatID, err)
		}
	}

	sendHappConnect := func(chatID int64) {
		if err := vpnKeys.SendOpenHapp(chatID, "", vlessKeys.Get(chatID), routingProfiles.Get(chatID)); err != nil {
			log.Printf("send happ connect failed (chatId=%d): %v", chatID, err)
			_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Не удалось отправить инструкцию. Попробуйте позже."))
		}
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			bot.StopReceivingUpdates()
			return
		case update, ok := <-updates:
			if !ok {
				return
			}

			if update.CallbackQuery != nil {
				if happ.HandleCallback(bot, update.CallbackQuery) {
					continue
				}
			}

			if update.Message == nil {
				continue
			}

			text := strings.TrimSpace(update.Message.Text)
			chatID := update.Message.Chat.ID

			if strings.HasPrefix(strings.ToLower(text), "vless://") {
				vlessKeys.Set(chatID, text)
				if err := vpnKeys.SendOpenHapp(chatID, "Ключ получен. Нажмите «Открыть Happ»:", text, ""); err != nil {
					log.Printf("send happ after vless failed (chatId=%d): %v", chatID, err)
					_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Ключ сохранён, но кнопки не отправились."))
				}
				continue
			}

			if isRoutingMessage(text) {
				b64, err := happ.ParseRoutingInput(text)
				if err != nil {
					_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Не распознан профиль routing."))
					continue
				}
				routingProfiles.Set(chatID, b64)
				if err := vpnKeys.SendOpenHapp(chatID, "Профиль routing сохранён:", "", b64); err != nil {
					log.Printf("send happ after routing failed (chatId=%d): %v", chatID, err)
				}
				continue
			}

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					sendWelcome(chatID)
					continue
				case "help":
					msg := tgbotapi.NewMessage(chatID, botapp.HelpText())
					msg.ReplyMarkup = menu
					_, _ = bot.Send(msg)
					continue
				case "connect":
					sendHappConnect(chatID)
					continue
				}
			}

			switch text {
			case botapp.BtnOpenHapp:
				sendHappConnect(chatID)
				continue
			case botapp.BtnStart:
				sendWelcome(chatID)
				continue
			}
		}
	}
}

func isRoutingMessage(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	lower := strings.ToLower(text)
	return strings.Contains(lower, "happ://routing/") || strings.HasPrefix(text, "{")
}
