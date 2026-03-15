package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"genai-gallery-backend/internal/auth"
	"genai-gallery-backend/internal/config"
)

// NetworkAuthMiddleware conditionally requires a Bearer token or Basic Auth.
// If the request originates from a Loopback address (localhost) or a Private LAN block,
// it is allowed immediately, UNLESS config.RequireAuth is true.
// If it's originating from an external/public network or forced by config,
// it checks the Authorization header for either the generated Bearer token or the Basic Auth password.
func NetworkAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		parsedIP := net.ParseIP(clientIP)

		// Bypass auth if not explicitly required AND IP is local/private
		if !config.RequireAuth && parsedIP != nil {
			if parsedIP.IsLoopback() || parsedIP.IsPrivate() {
				c.Next()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")

		// If Authorization header is present and starts with Bearer, check the Bearer token
		if authHeader != "" && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 {
				token := parts[1]
				if token == auth.GlobalBearerToken {
					c.Next()
					return
				}
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid Bearer token"})
			return
		}

		// Fallback to Basic Auth (for browsers)
		_, password, hasBasicAuth := c.Request.BasicAuth()
		if hasBasicAuth && password == auth.GlobalBasicAuthPassword {
			c.Next()
			return
		}

		// If no valid auth was provided, issue a WWW-Authenticate challenge for Basic Auth
		c.Header("WWW-Authenticate", `Basic realm="GenAI Gallery"`)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
	}
}
