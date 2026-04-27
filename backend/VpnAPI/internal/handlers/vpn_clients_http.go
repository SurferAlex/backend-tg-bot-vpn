package handlers

import (
	"api-vpn/internal/model"
	"api-vpn/internal/usecase"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

type CreateClientRequest struct {
	TelegramUserID *int64  `json:"telegramUserId"`
	MaxIPs         int     `json:"maxIps"`
	TTLSeconds     int64   `json:"ttlSeconds"`
	Note           *string `json:"note"`
}
type ClientResponse struct {
	ID             int64     `json:"id"`
	ClientUUID     string    `json:"clientUuid"`
	TelegramUserID *int64    `json:"telegramUserId,omitempty"`
	MaxIPs         int       `json:"maxIps"`
	KeyExpiresAt   time.Time `json:"keyExpiresAt"`
	IsActive       bool      `json:"isActive"`
	Note           *string   `json:"note,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func toClientResponse(c model.VPNClient) ClientResponse {
	return ClientResponse{
		ID:             c.ID,
		ClientUUID:     c.ClientUUID.String(),
		TelegramUserID: c.TelegramUserID,
		MaxIPs:         c.MaxIPs,
		KeyExpiresAt:   c.KeyExpiresAt,
		IsActive:       c.IsActive,
		Note:           c.Note,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
func (h *Handlers) CreateClient(c *gin.Context) {
	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	u, err := uuid.NewV4()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "uuid generation failed"})
		return
	}
	ttl := req.TTLSeconds
	if ttl <= 0 {
		ttl = 24 * 60 * 60
	}
	client, err := h.Clients.Create(c.Request.Context(), model.CreateVPNClientParams{
		ClientUUID:     u,
		TelegramUserID: req.TelegramUserID,
		MaxIPs:         req.MaxIPs,
		KeyExpiresAt:   time.Now().Add(time.Duration(ttl) * time.Second),
		Note:           req.Note,
	})
	if err != nil {
		log.Printf("create client failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, toClientResponse(client))
}
func (h *Handlers) GetClient(c *gin.Context) {
	idStr := c.Param("uuid")
	id, err := uuid.FromString(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}
	client, err := h.Clients.GetByUUID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		log.Printf("get client failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, toClientResponse(client))
}
func (h *Handlers) DeactivateClient(c *gin.Context) {
	idStr := c.Param("uuid")
	id, err := uuid.FromString(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}
	if err := h.Clients.Deactivate(c.Request.Context(), id); err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		log.Printf("deactivate client failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusNoContent)
}
