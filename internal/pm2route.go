package internal

import (
	"VPS-control/internal/auth"
	"VPS-control/internal/middleware"
	"VPS-control/internal/vps/pm2"

	"github.com/gin-gonic/gin"
)

func RegisterPM2Routes(
	rg *gin.RouterGroup,
	h *pm2.Handler,
) {
	pm2Group := rg.Group("/pm2")
	{
		pm2Group.GET("/processes/basic", middleware.RequirePermission(auth.PermPM2ViewBasic), h.GetProcessesBasic)
		pm2Group.GET("/processes/cwd", middleware.RequirePermission(auth.PermPM2ViewCwd), h.GetProcessesWithCwd)
		pm2Group.GET("/processes/full", middleware.RequirePermission(auth.PermPM2ViewFull), h.GetProcessesFull)

		pm2Group.POST("/:name/restart", middleware.RequirePermission(auth.PermPM2ControlRestart), h.Restart)
		pm2Group.POST("/:name/start", middleware.RequirePermission(auth.PermPM2ControlStart), h.Start)
		pm2Group.POST("/:name/stop", middleware.RequirePermission(auth.PermPM2ControlStop), h.Stop)
	}
}
