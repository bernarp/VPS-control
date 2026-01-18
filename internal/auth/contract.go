package auth

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	Login(c *gin.Context)
	Verify(c *gin.Context)
	Logout(c *gin.Context)
	GetSessions(c *gin.Context)
	RevokeSession(c *gin.Context)
}

type JwtProvider interface {
	GenerateToken(data TokenData) (string, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
	GetIssuer() string
	GetTTL() time.Duration
}

type SetAuthCookie interface {
	SetAuthCookie(
		c *gin.Context,
		token string,
	)
	GetAuthCookie(c *gin.Context) (string, error)
	ClearAuthCookie(c *gin.Context)
}

type AuthManager interface {
	Login(
		ctx context.Context,
		username, password string,
	) (*AuthResult, error)
	GetUserPermissions(
		ctx context.Context,
		userID int,
	) ([]string, error)
	HasPermission(
		ctx context.Context,
		userID int,
		permission string,
	) (bool, error)
}
