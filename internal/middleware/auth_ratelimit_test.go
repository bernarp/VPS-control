package middleware

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestAuthRateLimiter_IsBlocked(t *testing.T) {
	arl := NewAuthRateLimiter(3, 30*time.Second, zap.NewNop())
	ip := "192.168.1.100"

	t.Run(
		"not blocked initially", func(t *testing.T) {
			blocked, _ := arl.IsBlocked(ip)
			if blocked {
				t.Error("should not be blocked initially")
			}
		},
	)

	t.Run(
		"blocked after max attempts", func(t *testing.T) {
			for i := 0; i < 3; i++ {
				arl.RecordFailure(ip)
			}

			blocked, retryAfter := arl.IsBlocked(ip)
			if !blocked {
				t.Error("should be blocked after 3 failures")
			}
			if retryAfter <= 0 {
				t.Error("retryAfter should be positive")
			}
		},
	)
}

func TestAuthRateLimiter_ResetFailures(t *testing.T) {
	arl := NewAuthRateLimiter(3, 30*time.Second, zap.NewNop())
	ip := "192.168.1.101"

	arl.RecordFailure(ip)
	arl.RecordFailure(ip)

	arl.ResetFailures(ip)

	arl.mu.RLock()
	_, exists := arl.attempts[ip]
	arl.mu.RUnlock()

	if exists {
		t.Error("failures should be reset")
	}
}

func TestAuthRateLimiter_RecordFailure(t *testing.T) {
	arl := NewAuthRateLimiter(5, time.Minute, zap.NewNop())
	ip := "192.168.1.102"

	arl.RecordFailure(ip)

	arl.mu.RLock()
	attempt := arl.attempts[ip]
	arl.mu.RUnlock()

	if attempt.failures != 1 {
		t.Errorf("failures = %d, want 1", attempt.failures)
	}

	arl.RecordFailure(ip)
	arl.RecordFailure(ip)

	arl.mu.RLock()
	attempt = arl.attempts[ip]
	arl.mu.RUnlock()

	if attempt.failures != 3 {
		t.Errorf("failures = %d, want 3", attempt.failures)
	}
}

func TestAuthRateLimiter_DifferentIPs(t *testing.T) {
	arl := NewAuthRateLimiter(2, time.Minute, zap.NewNop())

	arl.RecordFailure("1.1.1.1")
	arl.RecordFailure("1.1.1.1")

	blocked1, _ := arl.IsBlocked("1.1.1.1")
	blocked2, _ := arl.IsBlocked("2.2.2.2")

	if !blocked1 {
		t.Error("IP 1.1.1.1 should be blocked")
	}
	if blocked2 {
		t.Error("IP 2.2.2.2 should not be blocked")
	}
}
