package main

import (
	"DiscordBotControl/internal"
	"DiscordBotControl/internal/middleware"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func setupRouter(logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(ginzap.Ginzap(logger.Named("http"), time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	r.Use(middleware.SecurityHeadersMiddleware())
	return r
}

func (app *application) registerRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(app.sanitizer.Middleware())

	authGroup := api.Group("/auth")
	authGroup.Use(
		middleware.AuthRateLimitMiddleware(
			app.cfg.RateLimit.Auth.MaxAttempts,
			app.cfg.RateLimit.Auth.BlockTime,
			app.logger,
		),
	)

	authMW := middleware.AuthMiddleware(app.authJwt, app.authCookie, app.tokenRepo, app.logger)
	internal.RegisterAuthRoutes(authGroup, app.authHdl, authMW)

	vpsGroup := api.Group("/vps")
	vpsGroup.Use(
		middleware.RateLimitMiddleware(
			app.cfg.RateLimit.API.Limit,
			app.cfg.RateLimit.API.Window,
		),
	)
	vpsGroup.Use(authMW)

	internal.RegisterPM2Routes(vpsGroup, app.pm2Hdl)
	internal.RegisterFail2BanRoutes(vpsGroup, app.f2bHdl)

	r.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
