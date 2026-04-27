package api

import (
	"tg-bot/internal/dedup"
	"tg-bot/internal/handlers"
	"tg-bot/internal/middleware"
	"tg-bot/internal/telegram"
	"tg-bot/vpnapi"

	"github.com/gin-gonic/gin"
)

func SetupServer(
	notifier *telegram.Notifier,
	internalToken string,
	deduper *dedup.Deduper,
	vpnAPI *vpnapi.API) *gin.Engine {
	r := gin.Default()

	h := &handlers.Handlers{
		Notifier: notifier,
		Deduper:  deduper,
		VPNAPI:   vpnAPI,
	}

	RegisterRoutes(r, h, internalToken)

	return r
}

func RegisterRoutes(r *gin.Engine, h *handlers.Handlers, internalToken string) {
	internal := r.Group("/internal", middleware.InternalToken(internalToken))
	{
		internal.POST("/alert", h.Alert)
	}
}
