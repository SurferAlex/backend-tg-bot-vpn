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
	VlessURL         string // vless:// → happ://add/vless://…
	RoutingB64       string // base64 routing profile
}

func BuildMessageText(opts DeliveryOptions) string {
	var b strings.Builder
	if s := strings.TrimSpace(opts.ExtraCaption); s != "" {
		b.WriteString(s)
		b.WriteString("\n\n")
	}
	b.WriteString("1️⃣ «Скачать Happ» — App Store\n")
	b.WriteString("2️⃣ «Открыть Happ» — откроет Happ с вашим ключом\n\n")
	b.WriteString("Отправьте vless:// в чат или профиль с https://routing.happ.su")
	return b.String()
}

func InlineDownloadURL(opts DeliveryOptions) string {
	return AppStoreURL(opts.IOSAppStoreURL)
}

// InlineOpenURL prefers vless redirect, else routing.
func InlineOpenURL(opts DeliveryOptions) string {
	base := ResolveOpenRedirectBase(opts.OpenRedirectBase)
	if u := OpenRedirectURLVless(base, opts.VlessURL); u != "" {
		return u
	}
	if u := OpenRedirectURL(base, opts.RoutingB64); u != "" {
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
	return "Отправьте vless:// в чат. Если Happ не открылся — «Открыть в Safari» и нажмите кнопку на странице."
}

func DeliverIOS(bot *tgbotapi.BotAPI, chatID int64, opts DeliveryOptions) error {
	msg := tgbotapi.NewMessage(chatID, BuildMessageText(opts))
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = BuildInlineKeyboard(opts)
	_, err := bot.Send(msg)
	return err
}
