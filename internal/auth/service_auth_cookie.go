package auth

import (
	"net/http"

	"VPS-control/internal/config"

	"github.com/gin-gonic/gin"
)

var _ SetAuthCookie = (*AuthCookieService)(nil)

type AuthCookieService struct {
	name     string
	maxAge   int
	secure   bool
	httpOnly bool
	sameSite http.SameSite
}

func NewAuthCookieService(cfg *config.Config) *AuthCookieService {
	sameSite := http.SameSiteStrictMode
	switch cfg.Cookie.SameSite {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}

	return &AuthCookieService{
		name:     cfg.Cookie.Name,
		maxAge:   int(cfg.JWT.TTL.Seconds()),
		secure:   cfg.Cookie.Secure,
		httpOnly: cfg.Cookie.HttpOnly,
		sameSite: sameSite,
	}
}

func (s *AuthCookieService) SetAuthCookie(
	c *gin.Context,
	token string,
) {
	c.SetSameSite(s.sameSite)
	c.SetCookie(
		s.name,
		token,
		s.maxAge,
		"/",
		"",
		s.secure,
		s.httpOnly,
	)
}

func (s *AuthCookieService) GetAuthCookie(c *gin.Context) (string, error) {
	return c.Cookie(s.name)
}

func (s *AuthCookieService) ClearAuthCookie(c *gin.Context) {
	c.SetSameSite(s.sameSite)
	c.SetCookie(
		s.name,
		"",
		-1,
		"/",
		"",
		s.secure,
		s.httpOnly,
	)
}
