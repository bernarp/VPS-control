package middleware

import (
	"DiscordBotControl/internal/apierror"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type loginAttempt struct {
	failures  int
	blockedAt time.Time
}

type AuthRateLimiter struct {
	attempts    map[string]*loginAttempt
	mu          sync.RWMutex
	maxAttempts int
	blockTime   time.Duration
	logger      *zap.Logger
}

func NewAuthRateLimiter(
	maxAttempts int,
	blockTime time.Duration,
	logger *zap.Logger,
) *AuthRateLimiter {
	arl := &AuthRateLimiter{
		attempts:    make(map[string]*loginAttempt),
		maxAttempts: maxAttempts,
		blockTime:   blockTime,
		logger:      logger.Named("auth_rate_limit"),
	}
	go arl.cleanup()
	return arl
}

func (arl *AuthRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		arl.mu.Lock()
		now := time.Now()
		for ip, attempt := range arl.attempts {
			if now.After(attempt.blockedAt.Add(arl.blockTime)) {
				delete(arl.attempts, ip)
			}
		}
		arl.mu.Unlock()
	}
}

func (arl *AuthRateLimiter) IsBlocked(ip string) (bool, int) {
	arl.mu.RLock()
	defer arl.mu.RUnlock()

	attempt, exists := arl.attempts[ip]
	if !exists {
		return false, 0
	}

	if attempt.failures >= arl.maxAttempts {
		remaining := int(arl.blockTime.Seconds() - time.Since(attempt.blockedAt).Seconds())
		if remaining > 0 {
			return true, remaining
		}
	}
	return false, 0
}

func (arl *AuthRateLimiter) RecordFailure(ip string) {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	attempt, exists := arl.attempts[ip]
	if !exists {
		arl.attempts[ip] = &loginAttempt{
			failures:  1,
			blockedAt: time.Now(),
		}
		return
	}

	attempt.failures++
	if attempt.failures >= arl.maxAttempts {
		attempt.blockedAt = time.Now()
		arl.logger.Warn(
			"IP blocked due to too many login attempts",
			zap.String("ip", ip),
			zap.Int("attempts", attempt.failures),
		)
	}
}

func (arl *AuthRateLimiter) ResetFailures(ip string) {
	arl.mu.Lock()
	defer arl.mu.Unlock()
	delete(arl.attempts, ip)
}

func AuthRateLimitMiddleware(
	maxAttempts int,
	blockTime time.Duration,
	logger *zap.Logger,
) gin.HandlerFunc {
	limiter := NewAuthRateLimiter(maxAttempts, blockTime, logger)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if blocked, retryAfter := limiter.IsBlocked(ip); blocked {
			apierror.Abort(
				c, apierror.Errors.RATE_LIMIT_EXCEEDED.WithMeta(
					gin.H{
						"retry_after": retryAfter,
					},
				),
			)
			return
		}

		c.Next()

		status := c.Writer.Status()
		switch status {
		case http.StatusUnauthorized, http.StatusForbidden:
			limiter.RecordFailure(ip)
		case http.StatusOK:
			limiter.ResetFailures(ip)
		}
	}
}
