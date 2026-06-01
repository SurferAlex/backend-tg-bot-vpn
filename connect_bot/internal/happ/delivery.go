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
	OpenRedirectBase string // https://host/happ/open — for inline «Открыть Happ»
	VlessURL         string // vless://… → happ://add/vless://…
}

// BuildMessageText returns instructions (key may be sent separately or pasted earlier).
func BuildMessageText(opts DeliveryOptions) string {
	var b strings.Builder
	if s := strings.TrimSpace(opts.ExtraCaption); s != "" {
		b.WriteString(s)
		b.WriteString("\n\n")
	}
	b.WriteString("1️⃣ «Скачать Happ» — App Store, если приложения ещё нет\n")
	b.WriteString("2️⃣ «Открыть Happ» — добавит ваш ключ в Happ\n")
	b.WriteString("\n")
	b.WriteString("Сначала отправьте ключ vless:// в этот чат.")
	return b.String()
}

// InlineDownloadURL always opens App Store.
func InlineDownloadURL(opts DeliveryOptions) string {
	return AppStoreURL(opts.IOSAppStoreURL)
}

// InlineOpenURL is https → happ://add/{vless} (Telegram rejects happ:// in buttons).
func InlineOpenURL(opts DeliveryOptions) string {
	vless := strings.TrimSpace(opts.VlessURL)
	if vless == "" {
		return AppStoreURL(opts.IOSAppStoreURL)
	}
	base := ResolveOpenRedirectBase(opts.OpenRedirectBase)
	if u := OpenRedirectURL(base, vless); u != "" {
		return u
	}
	return AppStoreURL(opts.IOSAppStoreURL)
}

// BuildInlineKeyboard uses https only (Telegram allows http(s)/tg, not happ://).
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

// ImportHelpText is sent when user taps «Happ уже установлен».
func ImportHelpText() string {
	return "Откройте Happ → «+» → «Импорт из буфера» или снова нажмите «Открыть Happ» после отправки vless:// в чат."
}

// DeliverIOS sends Happ onboarding (instructions + inline buttons).
func DeliverIOS(bot *tgbotapi.BotAPI, chatID int64, opts DeliveryOptions) error {
	msg := tgbotapi.NewMessage(chatID, BuildMessageText(opts))
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = BuildInlineKeyboard(opts)
	_, err := bot.Send(msg)
	return err
}
