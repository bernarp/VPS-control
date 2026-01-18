package internal

import (
	"VPS-control/internal/auth"
	"VPS-control/internal/middleware"
	"VPS-control/internal/vps/fail2ban"

	"github.com/gin-gonic/gin"
)

func RegisterFail2BanRoutes(
	rg *gin.RouterGroup,
	h *fail2ban.Handler,
) {
	f2bGroup := rg.Group("/fail2ban")
	{
		f2bGroup.GET("/status", middleware.RequirePermission(auth.PermF2BViewStatus), h.GetStatus)
		f2bGroup.GET("/status/:name", middleware.RequirePermission(auth.PermF2BViewJail), h.GetJailDetails)
		f2bGroup.POST("/unban", middleware.RequirePermission(auth.PermF2BControlUnban), h.Unban)
	}
}
