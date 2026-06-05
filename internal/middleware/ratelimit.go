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
	rate     float64
	burst    int
	stop     chan struct{}
}

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

func newRateLimiter(rate, burst int) *rateLimiter {
	if rate <= 0 {
		rate = 10
	}
	if burst <= 0 {
		burst = 30
	}
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     float64(rate),
		burst:    burst,
		stop:     make(chan struct{}),
	}
	go rl.cleanup(5 * time.Minute)
	return rl
}

func (rl *rateLimiter) cleanup(interval time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			go rl.cleanup(interval)
		}
	}()
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

const maxVisitors = 10000

func (rl *rateLimiter) allow(ip string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists {
		if len(rl.visitors) >= maxVisitors {
			return false, 0
		}
		v = &visitor{tokens: float64(rl.burst), lastSeen: now}
		rl.visitors[ip] = v
	} else {
		elapsed := now.Sub(v.lastSeen).Seconds()
		v.lastSeen = now
		v.tokens += elapsed * rl.rate
		if v.tokens > float64(rl.burst) {
			v.tokens = float64(rl.burst)
		}
	}

	if v.tokens >= 1 {
		v.tokens--
		return true, int(v.tokens)
	}
	return false, 0
}

var defaultLimiter *rateLimiter
var defaultMu sync.Mutex

func StopRateLimiter() {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultLimiter != nil {
		close(defaultLimiter.stop)
		defaultLimiter = nil
	}
}

func RateLimitMiddleware(rate, burst int) gin.HandlerFunc {
	defaultMu.Lock()
	if defaultLimiter == nil {
		defaultLimiter = newRateLimiter(rate, burst)
	}
	limiter := defaultLimiter
	defaultMu.Unlock()

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
