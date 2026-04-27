package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Notifier struct {
	Bot      *tgbotapi.BotAPI
	AdminIDs []int64
}

func (n *Notifier) NotifyAdmins(text string) {
	for _, id := range n.AdminIDs {
		msg := tgbotapi.NewMessage(id, text)
		_, _ = n.Bot.Send(msg)
	}
}
