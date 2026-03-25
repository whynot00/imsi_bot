package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ApiKeyAuth validates the X-API-Key header against the API_KEY env variable.
func ApiKeyAuth() gin.HandlerFunc {
	key := os.Getenv("API_KEY")
	if key == "" {
		panic("API_KEY environment variable is not set")
	}
	return func(c *gin.Context) {
		if c.GetHeader("X-API-Key") != key {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		c.Next()
	}
}
