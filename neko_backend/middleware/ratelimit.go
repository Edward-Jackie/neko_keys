package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// bucket is a single per-IP token bucket.
type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// rateLimiter is a per-client-IP token-bucket limiter implemented with the
// standard library only (no external dependency). Tokens refill continuously
// at ratePerSec up to a maximum of burst.
type rateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*bucket
	ratePerSec float64
	burst      float64
}

func newRateLimiter(perMinute, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors:   make(map[string]*bucket),
		ratePerSec: float64(perMinute) / 60.0,
		burst:      float64(burst),
	}
	go rl.cleanupLoop()
	return rl
}

// allow consumes one token for ip, returning false when the bucket is empty.
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.visitors[ip]
	if !ok {
		rl.visitors[ip] = &bucket{tokens: rl.burst - 1, lastSeen: now}
		return true
	}

	// Refill based on time elapsed since the last request.
	b.tokens += now.Sub(b.lastSeen).Seconds() * rl.ratePerSec
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// cleanupLoop evicts buckets that have been idle long enough to be at full
// capacity, keeping the map from growing unbounded.
func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, b := range rl.visitors {
			if time.Since(b.lastSeen) > 30*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// limiterClientIP mirrors handlers.getClientIP so the limiter keys on the real
// client address behind a reverse proxy.
func limiterClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.SplitN(ip, ",", 2)[0])
	}
	if ip := c.GetHeader("X-Real-Ip"); ip != "" {
		return ip
	}
	return c.ClientIP()
}

// RateLimit returns a Gin middleware that limits each client IP to perMinute
// requests, allowing short bursts up to burst. Exceeding the limit returns 429.
func RateLimit(perMinute, burst int) gin.HandlerFunc {
	rl := newRateLimiter(perMinute, burst)
	return func(c *gin.Context) {
		if !rl.allow(limiterClientIP(c)) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}
