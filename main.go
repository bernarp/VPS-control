package main

import (
	"log"

	_ "DiscordBotControl/docs"
	"DiscordBotControl/internal/config"

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
	defer func() { _ = logger.Sync() }()

	app := initApp(cfg, logger)
	defer app.Close()

	r := setupRouter(logger)
	app.registerRoutes(r)
	logger.Info(
		"DiscordBotControl API started",
		zap.String("port", cfg.Server.Port),
		zap.String("local_db", cfg.Storage.LocalDBPath),
	)
	if err := r.Run(cfg.Server.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
