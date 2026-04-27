package handlers

import (
	"tg-bot/internal/dedup"
	"tg-bot/internal/telegram"
	"tg-bot/vpnapi"
)

type Handlers struct {
	Notifier *telegram.Notifier
	Deduper  *dedup.Deduper
	VPNAPI   *vpnapi.API
}
