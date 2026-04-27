package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InternalToken(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Internal-Token") != expected {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
