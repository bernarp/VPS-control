package middleware

import (
	"VPS-control/internal/apierror"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

type RateLimiter struct {
	requests map[string]*rateLimitEntry
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(
	limit int,
	window time.Duration,
) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*rateLimitEntry),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.requests {
			if now.After(entry.resetTime) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func RateLimitMiddleware(
	limit int,
	window time.Duration,
) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		limiter.mu.Lock()
		entry, exists := limiter.requests[ip]

		if !exists || now.After(entry.resetTime) {
			limiter.requests[ip] = &rateLimitEntry{
				count:     1,
				resetTime: now.Add(limiter.window),
			}
			limiter.mu.Unlock()
			c.Next()
			return
		}

		if entry.count >= limiter.limit {
			limiter.mu.Unlock()
			retryAfter := int(entry.resetTime.Sub(now).Seconds())
			apierror.Abort(
				c, apierror.Errors.RATE_LIMIT_EXCEEDED.WithMeta(
					gin.H{
						"retry_after": retryAfter,
					},
				),
			)
			return
		}

		entry.count++
		limiter.mu.Unlock()
		c.Next()
	}
}
