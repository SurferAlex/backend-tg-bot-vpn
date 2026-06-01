package botapp

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	BtnStart    = "Старт"
	BtnOpenHapp = "📲 Подключить Happ"
)

func MainMenuMarkup() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnStart),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnOpenHapp),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func WelcomeText(botUsername string) string {
	_ = botUsername
	return "Добро пожаловать!\n\n" +
		"1. Вставьте ключ vless:// в чат или получите у админа.\n" +
		"2. «Подключить Happ» → «Скачать Happ» (App Store) или «Открыть Happ» (добавить ключ)."
}

func HelpText() string {
	return "Порядок:\n" +
		"1. Отправьте vless:// в чат (бот запомнит ключ).\n" +
		"2. Меню «Подключить Happ» или /connect.\n" +
		"3. «Скачать Happ» — App Store; «Открыть Happ» — happ://add/ваш ключ.\n\n" +
		"Команды: /start, /help, /connect"
}
