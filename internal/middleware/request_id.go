package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			b := make([]byte, 8)
			if _, err := rand.Read(b); err != nil {
				rid = fmt.Sprintf("%x", time.Now().UnixNano())
			} else {
				rid = hex.EncodeToString(b)
			}
		}
		c.Header("X-Request-ID", rid)
		c.Set("request_id", rid)
		c.Next()
	}
}
