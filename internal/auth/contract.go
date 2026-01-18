package auth

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type SetAuthCookie interface {
	SetAuthCookie(
		c *gin.Context,
		token string,
	)
	GetAuthCookie(c *gin.Context) (string, error)
	ClearAuthCookie(c *gin.Context)
}
type JwtProvider interface {
	GenerateToken(data TokenData) (string, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
	GetIssuer() string
	GetTTL() time.Duration
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
