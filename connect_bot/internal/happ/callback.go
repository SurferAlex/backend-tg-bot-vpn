package happ

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCallback answers inline callbacks for Happ flow. Returns true if handled.
func HandleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) bool {
	if cb == nil || cb.Data != callbackHappHelp {
		return false
	}
	_, _ = bot.Request(tgbotapi.NewCallback(cb.ID, ""))
	if cb.Message != nil {
		msg := tgbotapi.NewMessage(cb.Message.Chat.ID, ImportHelpText())
		_, _ = bot.Send(msg)
	}
	return true
}
