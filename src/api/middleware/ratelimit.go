package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientState
	limit    int
	window   time.Duration
}

type clientState struct {
	count    int
	resetAt  time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string]*clientState),
		limit:   limit,
		window:  window,
	}
	// Periodic cleanup of expired entries
	go func() {
		for {
			time.Sleep(window * 2)
			rl.mu.Lock()
			now := time.Now()
			for ip, cs := range rl.clients {
				if now.After(cs.resetAt) {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cs, exists := rl.clients[clientIP]
	if !exists || now.After(cs.resetAt) {
		rl.clients[clientIP] = &clientState{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	cs.count++
	return cs.count <= rl.limit
}

// RateLimit returns middleware that limits requests per IP.
// limit: max requests allowed within the window duration.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(limit, window)
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			return
		}
		c.Next()
	}
}
