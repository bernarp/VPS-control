package internal

import (
	"DiscordBotControl/internal/auth"
	"DiscordBotControl/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(
	rg *gin.RouterGroup,
	h *auth.Handler,
	authMW gin.HandlerFunc,
) {
	rg.POST("/login", h.Login)
	rg.POST("/verify", authMW, h.Verify)
	rg.POST("/logout", authMW, h.Logout)

	sessions := rg.Group("/sessions")
	sessions.Use(authMW)
	{
		sessions.GET("", middleware.RequirePermission(auth.PermUserView), h.GetSessions)
		sessions.POST("/revoke", middleware.RequirePermission(auth.PermUserEdit), h.RevokeSession)
	}
}
