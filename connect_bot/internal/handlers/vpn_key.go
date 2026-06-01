package handlers

import (
	"connect-bot/internal/happ"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// VPNKeyHandler sends Happ iOS connect flow.
type VPNKeyHandler struct {
	bot               *tgbotapi.BotAPI
	iosAppStoreURL    string
	openRedirectBase  string
	defaultRoutingB64 string
}

func NewVPNKeyHandler(bot *tgbotapi.BotAPI, iosAppStoreURL, openRedirectBase, defaultRoutingB64 string) *VPNKeyHandler {
	return &VPNKeyHandler{
		bot:               bot,
		iosAppStoreURL:    iosAppStoreURL,
		openRedirectBase:  openRedirectBase,
		defaultRoutingB64: defaultRoutingB64,
	}
}

// SendOpenHapp shows instructions and inline buttons.
func (h *VPNKeyHandler) SendOpenHapp(chatID int64, extraCaption, vlessURL, routingB64 string) error {
	if routingB64 == "" {
		routingB64 = h.defaultRoutingB64
	}
	return happ.DeliverIOS(h.bot, chatID, happ.DeliveryOptions{
		IOSAppStoreURL:   h.iosAppStoreURL,
		OpenRedirectBase: h.openRedirectBase,
		VlessURL:         vlessURL,
		RoutingB64:       routingB64,
		ExtraCaption:     extraCaption,
	})
}
