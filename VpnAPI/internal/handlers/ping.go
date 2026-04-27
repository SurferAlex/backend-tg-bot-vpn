package handlers

import "github.com/gin-gonic/gin"

func (h *Handlers) Ping(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong"})
}
