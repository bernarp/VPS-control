package main

import (
	"log"

	_ "VPS-control/docs"
	"VPS-control/internal/config"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// @title           VPS_API
// @version         2.7
// @description     VPS management system.
// @BasePath        /api
// @schemes         https
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name discord_bot_auth
// @description HTTP-only cookie with JWT token (set automatically after login)

func main() {
	_ = godotenv.Load(".env.dev")

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	logger := InitLogger(cfg.Server.Debug)
	app := initApp(cfg, logger)
	if err := app.Run(); err != nil {
		logger.Fatal("Application terminated with error", zap.Error(err))
	}
}
