package main

import (
	"DiscordBotControl/internal/apierror"
	"DiscordBotControl/internal/auth"
	"DiscordBotControl/internal/config"
	"DiscordBotControl/internal/database/postgresql"
	"DiscordBotControl/internal/database/sqlite3_local"
	"DiscordBotControl/internal/middleware"
	"DiscordBotControl/internal/vps"
	"DiscordBotControl/internal/vps/fail2ban"
	"DiscordBotControl/internal/vps/pm2"
	"context"
	_ "embed"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed errors_code.yaml
var errorConfigData []byte

type application struct {
	cfg        *config.Config
	logger     *zap.Logger
	db         *postgresql.Database
	localDB    *sqlite3_local.LocalDB
	authHdl    *auth.Handler
	pm2Hdl     *pm2.Handler
	f2bHdl     *fail2ban.Handler
	authJwt    *auth.AuthJwtService
	authCookie *auth.AuthCookieService
	tokenRepo  *sqlite3_local.TokenRepository
	sanitizer  *middleware.InputSanitizer
}

func initApp(
	cfg *config.Config,
	logger *zap.Logger,
) *application {
	if err := apierror.Init(errorConfigData, logger); err != nil {
		logger.Fatal("Failed to initialize api errors registry", zap.Error(err))
	}

	db := initDatabase(cfg.Database, logger)
	localDB := initLocalDB(cfg, logger)

	userRepo := postgresql.NewUserRepository(db.Pool, logger)
	permRepo := postgresql.NewPermissionRepository(db.Pool, logger)
	tokenRepo := sqlite3_local.NewTokenRepository(localDB, logger)
	baseVpsSvc := vps.NewBaseVpsService()
	sanitizer := middleware.NewInputSanitizer(logger)

	authJwt := auth.NewAuthJwtService(cfg, logger)
	authCookie := auth.NewAuthCookieService(cfg)
	authMgr := auth.NewAuthManagerService(userRepo, permRepo)
	authHdl := auth.NewHandler(authMgr, authJwt, authCookie, tokenRepo, logger)

	pm2ListSvc := pm2.NewListService(baseVpsSvc)
	pm2ControlSvc := pm2.NewControlService(pm2ListSvc)
	pm2Hdl := pm2.NewHandler(pm2ListSvc, pm2ControlSvc, logger)

	f2bControlSvc := fail2ban.NewControlService(baseVpsSvc, logger)
	f2bHdl := fail2ban.NewHandler(f2bControlSvc, logger)

	return &application{
		cfg:        cfg,
		logger:     logger,
		db:         db,
		localDB:    localDB,
		authHdl:    authHdl,
		pm2Hdl:     pm2Hdl,
		f2bHdl:     f2bHdl,
		authJwt:    authJwt,
		authCookie: authCookie,
		tokenRepo:  tokenRepo,
		sanitizer:  sanitizer,
	}
}

func (app *application) Run() error {
	router := setupRouter(app.logger)
	app.registerRoutes(router)

	srv := &http.Server{
		Addr:         app.cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	serverErrors := make(chan error, 1)

	go func() {
		app.logger.Info("Server is starting", zap.String("port", app.cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-quit:
		app.logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

		shutdownTimeout := 25 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		app.logger.Info("Starting graceful shutdown", zap.Duration("timeout", shutdownTimeout))

		if err := srv.Shutdown(ctx); err != nil {
			app.logger.Error("Graceful shutdown failed", zap.Error(err))
			_ = srv.Close()
		} else {
			app.logger.Info("HTTP server stopped successfully")
		}
	}

	app.closeResources()

	if err := app.logger.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Logger sync error: %v\n", err)
	}

	app.logger.Info("Application terminated gracefully")
	return nil
}

func initDatabase(
	dbCfg config.DatabaseConfig,
	logger *zap.Logger,
) *postgresql.Database {
	db, err := postgresql.NewConnection(dbCfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	return db
}

func initLocalDB(
	cfg *config.Config,
	logger *zap.Logger,
) *sqlite3_local.LocalDB {
	localDB, err := sqlite3_local.NewLocalDB(cfg.Storage.LocalDBPath, logger)
	if err != nil {
		logger.Fatal("Failed to init local db", zap.Error(err), zap.String("path", cfg.Storage.LocalDBPath))
	}
	return localDB
}

func (app *application) closeResources() {
	app.logger.Info("Closing application resources")

	if app.localDB != nil {
		app.localDB.Close()
	}

	if app.db != nil {
		app.db.Close()
	}
}
