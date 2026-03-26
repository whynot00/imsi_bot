package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// IPWhitelist allows requests only from IPs listed in the ALLOWED_IPS env variable.
// Multiple IPs can be separated by commas: "1.2.3.4,5.6.7.8".
// Also always allows localhost for local development.
func IPWhitelist() gin.HandlerFunc {
	raw := os.Getenv("ALLOWED_IPS")
	if raw == "" {
		panic("ALLOWED_IPS environment variable is not set")
	}

	allowed := make(map[string]struct{})
	for _, ip := range strings.Split(raw, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			allowed[ip] = struct{}{}
		}
	}
	// always allow localhost
	for _, lo := range []string{"127.0.0.1", "::1"} {
		allowed[lo] = struct{}{}
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		// strip port if present
		host, _, err := net.SplitHostPort(clientIP)
		if err == nil {
			clientIP = host
		}

		if _, ok := allowed[clientIP]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.Next()
	}
}
