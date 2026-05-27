package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type monitorTargetResponse struct {
	ClientUUID     string `json:"clientUuid"`
	MaxIPs         int    `json:"maxIps"`
	XUIClientEmail string `json:"xuiClientEmail"`
}

func (h *Handlers) ListMonitorTargets(c *gin.Context) {
	targets, err := h.Clients.ListMonitorTargets(c.Request.Context())
	if err != nil {
		log.Printf("list monitor targets failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	out := make([]monitorTargetResponse, 0, len(targets))
	for _, t := range targets {
		out = append(out, monitorTargetResponse{
			ClientUUID:     t.ClientUUID.String(),
			MaxIPs:         t.MaxIPs,
			XUIClientEmail: t.XUIClientEmail,
		})
	}
	c.JSON(http.StatusOK, gin.H{"targets": out})
}
