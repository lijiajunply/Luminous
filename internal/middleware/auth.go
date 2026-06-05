package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"luminous/internal/config"
	"luminous/internal/response"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Cfg.Auth.AdminToken == "" {
			response.Error(c, http.StatusServiceUnavailable, "admin token not configured")
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, "invalid authorization format")
			c.Abort()
			return
		}

		token := parts[1]
		if subtle.ConstantTimeCompare([]byte(token), []byte(config.Cfg.Auth.AdminToken)) != 1 {
			response.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Next()
	}
}
