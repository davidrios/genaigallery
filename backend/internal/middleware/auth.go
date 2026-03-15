package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"genai-gallery-backend/internal/auth"
)

// NetworkAuthMiddleware conditionally requires a Bearer token.
// If the request originates from a Loopback address (localhost) or a Private LAN block,
// it is allowed immediately. If it's originating from an external/public network,
// it checks the Authorization header for the generated Bearer token.
func NetworkAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		parsedIP := net.ParseIP(clientIP)

		// If we can't parse the IP (should be rare with gin), fallback to secure mode (deny)
		if parsedIP != nil {
			if parsedIP.IsLoopback() {
				c.Next()
				return
			}
			if parsedIP.IsPrivate() {
				c.Next()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required for external access"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		token := parts[1]
		if token != auth.GlobalBearerToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
			return
		}

		c.Next()
	}
}
