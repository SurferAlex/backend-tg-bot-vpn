package handlers

import (
	"connect-bot/internal/happ"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// VPNKeyHandler sends Happ iOS connect flow.
type VPNKeyHandler struct {
	bot              *tgbotapi.BotAPI
	iosAppStoreURL   string
	openRedirectBase string
	defaultVless     string
}

func NewVPNKeyHandler(bot *tgbotapi.BotAPI, iosAppStoreURL, openRedirectBase, defaultVless string) *VPNKeyHandler {
	return &VPNKeyHandler{
		bot:              bot,
		iosAppStoreURL:   iosAppStoreURL,
		openRedirectBase: openRedirectBase,
		defaultVless:     defaultVless,
	}
}

// SendOpenHapp shows instructions and inline buttons.
func (h *VPNKeyHandler) SendOpenHapp(chatID int64, extraCaption, vlessURL string) error {
	if vlessURL == "" {
		vlessURL = h.defaultVless
	}
	return happ.DeliverIOS(h.bot, chatID, happ.DeliveryOptions{
		IOSAppStoreURL:   h.iosAppStoreURL,
		OpenRedirectBase: h.openRedirectBase,
		VlessURL:         vlessURL,
		ExtraCaption:     extraCaption,
	})
}
