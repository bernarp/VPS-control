// auth/cookie_test.go
package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"VPS-control/internal/config"

	"github.com/gin-gonic/gin"
)

func newTestCookieService() *AuthCookieService {
	cfg := &config.Config{
		Cookie: config.CookieConfig{
			Name:     "test_token",
			Secure:   false,
			HttpOnly: true,
			SameSite: "strict",
		},
		JWT: config.JWTConfig{
			TTL: time.Hour,
		},
	}
	return NewAuthCookieService(cfg)
}

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	return c, w
}

func TestSetAuthCookie(t *testing.T) {
	svc := newTestCookieService()
	c, w := setupTestContext()

	svc.SetAuthCookie(c, "test-token-value")

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies set")
	}

	var found *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "test_token" {
			found = cookie
			break
		}
	}

	if found == nil {
		t.Fatal("cookie 'test_token' not found")
	}

	if found.Value != "test-token-value" {
		t.Errorf("cookie value = %q, want %q", found.Value, "test-token-value")
	}

	if !found.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}

	if found.MaxAge != 3600 {
		t.Errorf("cookie MaxAge = %d, want %d", found.MaxAge, 3600)
	}
}

func TestGetAuthCookie(t *testing.T) {
	svc := newTestCookieService()

	t.Run(
		"cookie exists", func(t *testing.T) {
			c, _ := setupTestContext()
			c.Request.AddCookie(
				&http.Cookie{
					Name:  "test_token",
					Value: "my-token",
				},
			)

			token, err := svc.GetAuthCookie(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token != "my-token" {
				t.Errorf("token = %q, want %q", token, "my-token")
			}
		},
	)

	t.Run(
		"cookie missing", func(t *testing.T) {
			c, _ := setupTestContext()

			_, err := svc.GetAuthCookie(c)
			if err == nil {
				t.Error("expected error for missing cookie")
			}
		},
	)
}

func TestClearAuthCookie(t *testing.T) {
	svc := newTestCookieService()
	c, w := setupTestContext()

	svc.ClearAuthCookie(c)

	cookies := w.Result().Cookies()
	var found *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "test_token" {
			found = cookie
			break
		}
	}

	if found == nil {
		t.Fatal("cookie not found in response")
	}

	if found.MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", found.MaxAge)
	}

	if found.Value != "" {
		t.Errorf("cookie value = %q, want empty", found.Value)
	}
}

func TestSameSiteModes(t *testing.T) {
	tests := []struct {
		configValue string
		expected    http.SameSite
	}{
		{"strict", http.SameSiteStrictMode},
		{"lax", http.SameSiteLaxMode},
		{"none", http.SameSiteNoneMode},
		{"invalid", http.SameSiteStrictMode},
		{"", http.SameSiteStrictMode},
	}

	for _, tt := range tests {
		t.Run(
			tt.configValue, func(t *testing.T) {
				cfg := &config.Config{
					Cookie: config.CookieConfig{
						Name:     "test",
						SameSite: tt.configValue,
					},
					JWT: config.JWTConfig{TTL: time.Hour},
				}

				svc := NewAuthCookieService(cfg)
				if svc.sameSite != tt.expected {
					t.Errorf("sameSite = %v, want %v", svc.sameSite, tt.expected)
				}
			},
		)
	}
}

func TestCookieServiceFields(t *testing.T) {
	cfg := &config.Config{
		Cookie: config.CookieConfig{
			Name:     "custom_cookie",
			Secure:   true,
			HttpOnly: false,
			SameSite: "lax",
		},
		JWT: config.JWTConfig{TTL: 30 * time.Minute},
	}

	svc := NewAuthCookieService(cfg)

	if svc.name != "custom_cookie" {
		t.Errorf("name = %q, want %q", svc.name, "custom_cookie")
	}

	if svc.maxAge != 1800 {
		t.Errorf("maxAge = %d, want %d", svc.maxAge, 1800)
	}

	if !svc.secure {
		t.Error("secure should be true")
	}

	if svc.httpOnly {
		t.Error("httpOnly should be false")
	}

	if svc.sameSite != http.SameSiteLaxMode {
		t.Errorf("sameSite = %v, want %v", svc.sameSite, http.SameSiteLaxMode)
	}
}
