package main

import (
	"VPS-control/internal/apierror"
	"VPS-control/internal/auth"
	"VPS-control/internal/config"
	"VPS-control/internal/database/postgresql"
	"VPS-control/internal/database/sqlite3_local"
	"VPS-control/internal/middleware"
	"VPS-control/internal/nats"
	"VPS-control/internal/vps"
	"VPS-control/internal/vps/fail2ban"
	"VPS-control/internal/vps/pm2"
	_ "embed"
	"net/http"
	"time"

	"go.uber.org/zap"
)

//go:embed errors_code.yaml
var errorConfigData []byte

type application struct {
	cfg        *config.Config
	logger     *zap.Logger
	db         postgresql.DB
	localDB    sqlite3_local.LocalDatabase
	nats       nats.Connection
	broker     nats.Broker
	authHdl    auth.Handler
	pm2Hdl     pm2.Handler
	f2bHdl     fail2ban.Handler
	authJwt    auth.JwtProvider
	authCookie auth.SetAuthCookie
	tokenRepo  sqlite3_local.TokenStore
	sanitizer  middleware.Sanitizer
}

func initApp(
	cfg *config.Config,
	logger *zap.Logger,
) *application {
	if err := apierror.Init(errorConfigData, logger); err != nil {
		logger.Fatal("Failed to initialize api errors registry", zap.Error(err))
	}

	pgDB := initDatabase(cfg.Database, logger)
	s3DB := initLocalDB(cfg, logger)
	natsConn, err := nats.NewConnection(cfg.NATS, logger)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}

	userRepo := postgresql.NewUserRepository(pgDB.Pool, logger)
	permRepo := postgresql.NewPermissionRepository(pgDB.Pool, logger)
	tokenRepo := sqlite3_local.NewTokenRepository(s3DB, logger)
	baseVpsSvc := vps.NewBaseVpsService()
	sanitizer := middleware.NewInputSanitizer(logger)

	broker := nats.NewNatsBroker(natsConn)
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
		db:         pgDB,
		localDB:    s3DB,
		nats:       natsConn,
		broker:     broker,
		authHdl:    authHdl,
		pm2Hdl:     pm2Hdl,
		f2bHdl:     f2bHdl,
		authJwt:    authJwt,
		authCookie: authCookie,
		tokenRepo:  tokenRepo,
		sanitizer:  sanitizer,
	}
}

func (app *application) newServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         app.cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}
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

	if app.nats != nil {
		app.nats.Close()
	}

	if app.localDB != nil {
		app.localDB.Close()
	}

	if app.db != nil {
		app.db.Close()
	}
}
