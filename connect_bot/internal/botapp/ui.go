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
		"1. Отправьте ключ vless:// в чат\n" +
		"2. «Открыть Happ» — откроется Happ с ключом"
}

func HelpText() string {
	return "Порядок:\n" +
		"1. vless:// в чат\n" +
		"2. «Подключить Happ» → «Открыть Happ»\n" +
		"3. На странице нажмите «Открыть Happ» (в Safari надёжнее)\n\n" +
		"Команды: /start, /help, /connect"
}
