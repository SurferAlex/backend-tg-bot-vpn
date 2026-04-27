package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InternalToken(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal token is not configured"})
			return
		}
		if c.GetHeader("X-Internal-Token") != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
