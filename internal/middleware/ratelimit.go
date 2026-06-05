package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	burst    int
	stop     chan struct{}
}

type visitor struct {
	tokens   int
	lastSeen time.Time
}

func newRateLimiter(rate, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
		stop:     make(chan struct{}),
	}
	go rl.cleanup(5 * time.Minute)
	return rl
}

func (rl *rateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-interval)
			for ip, v := range rl.visitors {
				if v.lastSeen.Before(cutoff) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stop:
			return
		}
	}
}

func (rl *rateLimiter) allow(ip string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{tokens: rl.burst - 1, lastSeen: now}
		return true, rl.burst - 1
	}

	elapsed := now.Sub(v.lastSeen)
	v.lastSeen = now
	v.tokens += int(elapsed.Seconds() * float64(rl.rate))
	if v.tokens > rl.burst {
		v.tokens = rl.burst
	}

	if v.tokens > 0 {
		v.tokens--
		return true, v.tokens
	}
	return false, 0
}

var defaultLimiter *rateLimiter

func init() {
	defaultLimiter = newRateLimiter(10, 30)
}

func StopRateLimiter() {
	close(defaultLimiter.stop)
}

func RateLimitMiddleware(rate, burst int) gin.HandlerFunc {
	limiter := defaultLimiter
	if rate != 10 || burst != 30 {
		limiter = newRateLimiter(rate, burst)
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		allowed, remaining := limiter.allow(ip)
		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.burst))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if !allowed {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "rate limit exceeded",
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
