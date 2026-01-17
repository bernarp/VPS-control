// middleware/ratelimit_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimiter_Basic(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	ip := "192.168.1.1"

	for i := 0; i < 3; i++ {
		rl.mu.Lock()
		entry, exists := rl.requests[ip]
		if !exists {
			rl.requests[ip] = &rateLimitEntry{
				count:     1,
				resetTime: time.Now().Add(rl.window),
			}
		} else {
			entry.count++
		}
		rl.mu.Unlock()
	}

	rl.mu.RLock()
	count := rl.requests[ip].count
	rl.mu.RUnlock()

	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	middleware := RateLimitMiddleware(2, time.Minute)

	t.Run(
		"allows requests under limit", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.RemoteAddr = "10.0.0.1:12345"

			middleware(c)

			if c.IsAborted() {
				t.Error("request should not be aborted")
			}
		},
	)

	t.Run(
		"blocks after limit exceeded", func(t *testing.T) {
			mw := RateLimitMiddleware(1, time.Minute)
			w1 := httptest.NewRecorder()
			c1, _ := gin.CreateTestContext(w1)
			c1.Request = httptest.NewRequest("GET", "/", nil)
			c1.Request.RemoteAddr = "10.0.0.2:12345"
			mw(c1)

			w2 := httptest.NewRecorder()
			c2, _ := gin.CreateTestContext(w2)
			c2.Request = httptest.NewRequest("GET", "/", nil)
			c2.Request.RemoteAddr = "10.0.0.2:12345"
			mw(c2)

			if !c2.IsAborted() {
				t.Error("second request should be aborted")
			}

			if w2.Code != http.StatusTooManyRequests {
				t.Errorf("status = %d, want %d", w2.Code, http.StatusTooManyRequests)
			}
		},
	)
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	mw := RateLimitMiddleware(1, time.Minute)

	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest("GET", "/", nil)
	c1.Request.RemoteAddr = "1.1.1.1:1234"
	mw(c1)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/", nil)
	c2.Request.RemoteAddr = "2.2.2.2:1234"
	mw(c2)

	if c2.IsAborted() {
		t.Error("different IP should not be rate limited")
	}
}
