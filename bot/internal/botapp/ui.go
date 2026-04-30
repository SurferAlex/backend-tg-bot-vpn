package botapp

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	BtnCreate    = "➕ Создать ключ"
	BtnAccess    = "🔑 Доступ"
	BtnProvision = "⚡ Provision"
	BtnRevoke    = "⛔ Revoke"

	BtnBack   = "⬅️ Назад"
	BtnCancel = "Отмена"

	BtnDays30 = "30 дней"
	BtnDays60 = "60 дней"
	BtnDays90 = "90 дней"
)

func MainMenuMarkup() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnCreate),
			tgbotapi.NewKeyboardButton(BtnAccess),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnProvision),
			tgbotapi.NewKeyboardButton(BtnRevoke),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func NewKeyMarkup() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnDays30),
			tgbotapi.NewKeyboardButton(BtnDays60),
			tgbotapi.NewKeyboardButton(BtnDays90),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnBack),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func BackMarkup() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnBack),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func CancelMarkup() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnCancel),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}
