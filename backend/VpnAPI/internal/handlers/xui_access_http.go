package handlers

import (
	"api-vpn/internal/model"
	"api-vpn/internal/usecase"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

type AccessResponse struct {
	ClientUUID     string `json:"clientUuid"`
	InboundID      int64  `json:"inboundId"`
	XUIClientEmail string `json:"xuiClientEmail"`
	VLESSURI       string `json:"vlessUri"`
}

func toAccessResponse(a model.XUIAccess) AccessResponse {
	return AccessResponse{
		ClientUUID:     a.ClientUUID.String(),
		InboundID:      a.InboundID,
		XUIClientEmail: a.XUIClientEmail,
		VLESSURI:       a.VLESSURI,
	}
}

func (h *Handlers) ProvisionAccess(c *gin.Context) {
	if h.XUIAccess == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "xui access not configured"})
		return
	}
	idStr := c.Param("uuid")
	id, err := uuid.FromString(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}
	a, err := h.XUIAccess.Provision(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case errors.Is(err, usecase.ErrInactive):
			c.JSON(http.StatusForbidden, gin.H{"error": "inactive"})
		case errors.Is(err, usecase.ErrExpired):
			c.JSON(http.StatusForbidden, gin.H{"error": "expired"})
		default:
			log.Printf("provision access failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	c.JSON(http.StatusOK, toAccessResponse(a))
}

func (h *Handlers) GetAccess(c *gin.Context) {
	if h.XUIAccess == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "xui access not configured"})
		return
	}
	idStr := c.Param("uuid")
	id, err := uuid.FromString(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}
	a, err := h.XUIAccess.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		log.Printf("get access failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, toAccessResponse(a))
}

func (h *Handlers) RevokeAccess(c *gin.Context) {
	if h.XUIAccess == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "xui access not configured"})
		return
	}
	idStr := c.Param("uuid")
	id, err := uuid.FromString(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}
	if err := h.XUIAccess.Revoke(c.Request.Context(), id); err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		log.Printf("revoke access failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusNoContent)
}
