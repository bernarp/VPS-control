package middleware

import (
	"strings"

	"DiscordBotControl/internal/apierror"
	"DiscordBotControl/internal/auth"
	"DiscordBotControl/internal/database/sqlite3_local"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware(
	jwtService *auth.AuthJwtService,
	cookieService *auth.AuthCookieService,
	tokenRepo *sqlite3_local.TokenRepository,
	logger *zap.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := cookieService.GetAuthCookie(c)

		if err != nil || token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				apierror.Abort(c, apierror.Errors.PERMISSION_DENIED)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				apierror.Abort(c, apierror.Errors.INVALID_REQUEST)
				return
			}
			token = parts[1]
		}

		if token == "" || strings.Count(token, ".") != 2 {
			apierror.Abort(c, apierror.Errors.INVALID_REQUEST)
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			apierror.Abort(c, apierror.Errors.TOKEN_EXPIRED.Wrap(err))
			return
		}

		if claims.Username == "" || claims.Issuer != jwtService.GetIssuer() {
			apierror.Abort(c, apierror.Errors.INVALID_CREDENTIALS)
			return
		}

		if err := tokenRepo.ValidateToken(claims.JTI); err != nil {
			apierror.Abort(c, apierror.Errors.TOKEN_EXPIRED)
			return
		}

		c.Set(auth.CtxUsername, claims.Username)
		c.Set(auth.CtxUserID, claims.UserID)
		c.Set(auth.CtxJTI, claims.JTI)
		c.Set(auth.CtxRoles, claims.Roles)
		c.Set(auth.CtxPermissions, claims.Permissions)
		c.Set(auth.CtxClaims, claims)

		c.Next()
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := auth.GetClaims(c)
		if !ok {
			apierror.Abort(c, apierror.Errors.PERMISSION_DENIED)
			return
		}

		if !claims.HasPermission(permission) {
			apierror.Abort(c, apierror.Errors.ACTION_NOT_ALLOWED)
			return
		}

		c.Next()
	}
}

// RequireAnyPermission возвращена в файл
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := auth.GetClaims(c)
		if !ok {
			apierror.Abort(c, apierror.Errors.PERMISSION_DENIED)
			return
		}

		if !claims.HasAnyPermission(permissions...) {
			apierror.Abort(c, apierror.Errors.ACTION_NOT_ALLOWED)
			return
		}

		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := auth.GetClaims(c)
		if !ok {
			apierror.Abort(c, apierror.Errors.PERMISSION_DENIED)
			return
		}

		if !claims.HasRole(role) {
			apierror.Abort(c, apierror.Errors.ACTION_NOT_ALLOWED)
			return
		}

		c.Next()
	}
}
