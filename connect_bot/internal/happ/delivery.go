package happ

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	btnDownloadHapp  = "🔗 Скачать Happ"
	btnOpenHapp      = "📲 Открыть Happ"
	btnHappInstalled = "✅ Happ уже установлен"
	callbackHappHelp = "happ:import_help"
)

// DeliveryOptions configures Happ iOS onboarding.
type DeliveryOptions struct {
	ExtraCaption     string
	IOSAppStoreURL   string
	OpenRedirectBase string
	RoutingB64       string // base64 JSON routing profile
}

// BuildMessageText returns instructions.
func BuildMessageText(opts DeliveryOptions) string {
	var b strings.Builder
	if s := strings.TrimSpace(opts.ExtraCaption); s != "" {
		b.WriteString(s)
		b.WriteString("\n\n")
	}
	b.WriteString("1️⃣ «Скачать Happ» — App Store\n")
	b.WriteString("2️⃣ «Открыть Happ» — профиль маршрутизации (routing)\n\n")
	b.WriteString("Отправьте ссылку с https://routing.happ.su, JSON профиля или happ://routing/…")
	return b.String()
}

func InlineDownloadURL(opts DeliveryOptions) string {
	return AppStoreURL(opts.IOSAppStoreURL)
}

// InlineOpenURL → https redirect → happ://routing/onadd/{base64}.
func InlineOpenURL(opts DeliveryOptions) string {
	b64 := ResolveRoutingB64(opts.RoutingB64, "")
	base := ResolveOpenRedirectBase(opts.OpenRedirectBase)
	if u := OpenRedirectURL(base, b64); u != "" {
		return u
	}
	return AppStoreURL(opts.IOSAppStoreURL)
}

func BuildInlineKeyboard(opts DeliveryOptions) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(btnDownloadHapp, InlineDownloadURL(opts)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(btnOpenHapp, InlineOpenURL(opts)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnHappInstalled, callbackHappHelp),
		),
	)
}

func ImportHelpText() string {
	return "Создайте профиль на https://routing.happ.su → скопируйте ссылку happ://routing/… и отправьте боту."
}

func DeliverIOS(bot *tgbotapi.BotAPI, chatID int64, opts DeliveryOptions) error {
	msg := tgbotapi.NewMessage(chatID, BuildMessageText(opts))
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = BuildInlineKeyboard(opts)
	_, err := bot.Send(msg)
	return err
}
