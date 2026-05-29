package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServerResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handlers) ListServers(c *gin.Context) {
	list, err := h.Servers.ListActive(c.Request.Context())
	if err != nil {
		log.Printf("list servers failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	out := make([]ServerResponse, 0, len(list))
	for _, s := range list {
		out = append(out, ServerResponse{ID: s.ID, Name: s.Name})
	}
	c.JSON(http.StatusOK, gin.H{"servers": out})
}
