package handlers

import (
	"errors"
	"log"
	"net/http"

	"api-vpn/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

type XUIAccessResponse struct {
	ClientUUID     string `json:"clientUuid"`
	InboundID      int64  `json:"inboundId"`
	XUIClientEmail string `json:"xuiClientEmail"`
	VLESSURI       string `json:"vlessUri"`
}

func (h *Handlers) GetAccess(c *gin.Context) {
	id, err := uuid.FromString(c.Param("uuid"))
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
	c.JSON(http.StatusOK, XUIAccessResponse{
		ClientUUID:     a.ClientUUID.String(),
		InboundID:      a.InboundID,
		XUIClientEmail: a.XUIClientEmail,
		VLESSURI:       a.VLESSURI,
	})
}

func (h *Handlers) ProvisionAccess(c *gin.Context) {
	id, err := uuid.FromString(c.Param("uuid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}
	a, err := h.XUIAccess.Provision(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, usecase.ErrInactive) || errors.Is(err, usecase.ErrExpired) {
			if errors.Is(err, usecase.ErrInactive) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "client is inactive"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "client key is expired"})
			return
		}
		log.Printf("provision access failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, XUIAccessResponse{
		ClientUUID:     a.ClientUUID.String(),
		InboundID:      a.InboundID,
		XUIClientEmail: a.XUIClientEmail,
		VLESSURI:       a.VLESSURI,
	})
}

func (h *Handlers) RevokeAccess(c *gin.Context) {
	id, err := uuid.FromString(c.Param("uuid"))
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
