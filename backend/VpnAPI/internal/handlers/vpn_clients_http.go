package handlers

import (
	"api-vpn/internal/model"
	"api-vpn/internal/usecase"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

type CreateClientRequest struct {
	ServerID       string  `json:"serverId"`
	TelegramUserID *int64  `json:"telegramUserId"`
	MaxIPs         int     `json:"maxIps"`
	TTLSeconds     int64   `json:"ttlSeconds"`
	Note           *string `json:"note"`
}
type ClientResponse struct {
	ID             int64     `json:"id"`
	ClientUUID     string    `json:"clientUuid"`
	ServerID       string    `json:"serverId"`
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
		ServerID:       c.ServerID,
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
	if _, err := h.Servers.GetActiveByID(c.Request.Context(), req.ServerID); err != nil {
		if errors.Is(err, usecase.ErrNotFound) || errors.Is(err, usecase.ErrInvalidServer) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid serverId"})
			return
		}
		log.Printf("validate server failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	client, err := h.Clients.Create(c.Request.Context(), model.CreateVPNClientParams{
		ClientUUID:     u,
		ServerID:       req.ServerID,
		TelegramUserID: req.TelegramUserID,
		MaxIPs:         req.MaxIPs,
		KeyExpiresAt:   time.Now().Add(time.Duration(ttl) * time.Second),
		Note:           req.Note,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidServer) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid serverId"})
			return
		}
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

type ExtendClientRequest struct {
	AddDays int `json:"addDays"`
}

type UpdateMaxIPsRequest struct {
	MaxIPs int `json:"maxIps"`
}

func (h *Handlers) ResolveClient(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}
	client, err := h.Clients.ResolveRef(c.Request.Context(), q)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case errors.Is(err, usecase.ErrAmbiguousRef):
			c.JSON(http.StatusConflict, gin.H{"error": "ambiguous name, use client uuid"})
		default:
			log.Printf("resolve client failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	c.JSON(http.StatusOK, toClientResponse(client))
}

func (h *Handlers) ExtendClient(c *gin.Context) {
	id, ok := parseClientUUID(c)
	if !ok {
		return
	}
	var req ExtendClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	client, err := h.Clients.Extend(c.Request.Context(), id, req.AddDays)
	if err != nil {
		writeClientMutationError(c, err)
		return
	}
	if err := h.provisionClient(c, id); err != nil {
		log.Printf("extend client: provision failed (uuid=%s): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provision failed"})
		return
	}
	c.JSON(http.StatusOK, toClientResponse(client))
}

func (h *Handlers) UpdateClientMaxIPs(c *gin.Context) {
	id, ok := parseClientUUID(c)
	if !ok {
		return
	}
	var req UpdateMaxIPsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	client, err := h.Clients.UpdateMaxIPs(c.Request.Context(), id, req.MaxIPs)
	if err != nil {
		writeClientMutationError(c, err)
		return
	}
	if err := h.provisionClient(c, id); err != nil {
		log.Printf("update max ips: provision failed (uuid=%s): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provision failed"})
		return
	}
	c.JSON(http.StatusOK, toClientResponse(client))
}

func parseClientUUID(c *gin.Context) (uuid.UUID, bool) {
	idStr := c.Param("uuid")
	id, err := uuid.FromString(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return uuid.Nil, false
	}
	return id, true
}

func writeClientMutationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, usecase.ErrInactive):
		c.JSON(http.StatusBadRequest, gin.H{"error": "inactive"})
	case errors.Is(err, usecase.ErrInvalidExtend), errors.Is(err, usecase.ErrInvalidMaxIPs), errors.Is(err, usecase.ErrAmbiguousRef):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		log.Printf("client mutation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func (h *Handlers) provisionClient(c *gin.Context, id uuid.UUID) error {
	if h.XUIAccess == nil {
		return nil
	}
	_, err := h.XUIAccess.Provision(c.Request.Context(), id)
	return err
}
