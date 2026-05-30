package botapp

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	BtnCreate    = "➕ Создать ключ"
	BtnAccess    = "🔑 Доступ"
	BtnProvision = "⚡ Provision"
	BtnRevoke    = "⛔ Revoke"
	BtnExtend    = "📅 Продлить ключ"
	BtnMaxIPs    = "Limit IP"

	BtnBack   = "⬅️ Назад"
	BtnCancel = "Отмена"

	BtnDays30 = "30 дней"
	BtnDays60 = "60 дней"
	BtnDays90 = "90 дней"

	BtnIPs1 = "1 IP"
	BtnIPs2 = "2 IP"
	BtnIPs3 = "3 IP"
	BtnIPs4 = "4 IP"
	BtnIPs5 = "5 IP"
	BtnIPs6 = "6 IP"

	minMaxIPs = 1
	maxMaxIPs = 6
)

// MaxIPsButtonLabel returns the reply-keyboard label for the given limit.
func MaxIPsButtonLabel(n int) string {
	return fmt.Sprintf("%d IP", n)
}

// ParseMaxIPsButton parses a max-IPs button label. Returns (0, false) if not matched.
func ParseMaxIPsButton(text string) (int, bool) {
	for n := minMaxIPs; n <= maxMaxIPs; n++ {
		if text == MaxIPsButtonLabel(n) {
			return n, true
		}
	}
	return 0, false
}

// IsMainMenuButton reports fixed reply-keyboard labels from the main menu.
func IsMainMenuButton(text string) bool {
	switch text {
	case BtnCreate, BtnAccess, BtnProvision, BtnRevoke, BtnExtend, BtnMaxIPs:
		return true
	default:
		return false
	}
}

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
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnExtend),
			tgbotapi.NewKeyboardButton(BtnMaxIPs),
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

// ServerLabel is a VPN node shown in the create-key flow.
type ServerLabel struct {
	ID   string
	Name string
}

const serverBtnPrefix = "🖥 "

func ServerButtonLabel(name string) string {
	return serverBtnPrefix + name
}

// ParseServerButton matches a server keyboard label to server id.
func ParseServerButton(text string, servers []ServerLabel) (string, bool) {
	for _, s := range servers {
		if text == ServerButtonLabel(s.Name) {
			return s.ID, true
		}
	}
	return "", false
}

func ServersMarkup(servers []ServerLabel) tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard()
	var row []tgbotapi.KeyboardButton
	for i, s := range servers {
		row = append(row, tgbotapi.NewKeyboardButton(ServerButtonLabel(s.Name)))
		if len(row) == 2 || i == len(servers)-1 {
			kb.Keyboard = append(kb.Keyboard, row)
			row = nil
		}
	}
	kb.Keyboard = append(kb.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(BtnBack)))
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func MaxIPsMarkup() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnIPs1),
			tgbotapi.NewKeyboardButton(BtnIPs2),
			tgbotapi.NewKeyboardButton(BtnIPs3),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnIPs4),
			tgbotapi.NewKeyboardButton(BtnIPs5),
			tgbotapi.NewKeyboardButton(BtnIPs6),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnBack),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}
