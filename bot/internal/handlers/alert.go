package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"tg-bot/internal/model"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) Alert(c *gin.Context) {
	var req model.AlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("alert: invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if h.Deduper != nil {
		if ok := h.Deduper.Allow(req.ClientUUID, time.Now()); !ok {
			c.Status(http.StatusNoContent)
			return
		}
	}

	text := req.Text
	if text == "" {
		text = fmt.Sprintf(
			"ALERT: превышение IP\nClientUUID: %s\nИспользуется IP/Ограничение IP: %d/%d",
			req.ClientUUID,
			req.IPCount,
			req.MaxIPs,
		)
	}

	if h.VPNAPI != nil && req.ClientUUID != "" {
		client, err := h.VPNAPI.GetClient(c.Request.Context(), req.ClientUUID)
		if err != nil {
			log.Printf("alert: vpnapi get client failed (clientUuid=%s): %v", req.ClientUUID, err)
			text += "\nVpnAPI: unavailable"
		} else {

			text += "\n---"
			text += "\nmaxIps: " + fmt.Sprintf("%d", client.MaxIPs)
			text += "\nkeyExpiresAt: " + client.KeyExpiresAt.Format(time.RFC3339)
			text += "\nisActive: " + fmt.Sprintf("%v", client.IsActive)
			if client.Note != nil && strings.TrimSpace(*client.Note) != "" {
				text += "\nnote: " + strings.TrimSpace(*client.Note)
			}
		}
	}

	h.Notifier.NotifyAdmins(text)
	c.Status(http.StatusNoContent)
}
