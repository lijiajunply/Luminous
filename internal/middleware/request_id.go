package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
)

func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func sanitizeRequestID(rid string) string {
	var b strings.Builder
	for _, r := range rid {
		if unicode.IsPrint(r) && !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
		if b.Len() >= 64 {
			break
		}
	}
	return b.String()
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid != "" {
			rid = sanitizeRequestID(rid)
		}
		if rid == "" {
			rid = generateID()
		}
		c.Header("X-Request-ID", rid)
		c.Set("request_id", rid)
		c.Next()
	}
}
