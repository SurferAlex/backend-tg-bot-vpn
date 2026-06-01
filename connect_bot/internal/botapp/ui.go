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
		"1. Создайте профиль на https://routing.happ.su\n" +
		"2. Отправьте боту ссылку happ://routing/… или JSON\n" +
		"3. «Открыть Happ» — добавит маршрутизацию в приложение"
}

func HelpText() string {
	return "Порядок:\n" +
		"1. Профиль на https://routing.happ.su → скопировать ссылку\n" +
		"2. Вставить в чат боту\n" +
		"3. «Подключить Happ» → «Открыть Happ»\n\n" +
		"Команды: /start, /help, /connect"
}
