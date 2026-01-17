package postgresql

import (
	"context"
	"time"

	"DiscordBotControl/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Database struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewConnection(
	cfg config.DatabaseConfig,
	logger *zap.Logger,
) (*Database, error) {
	dbLogger := logger.Named("postgresql")
	databaseURL := cfg.GetDSN()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		dbLogger.Error("Failed to parse database DSN", zap.Error(err))
		return nil, err
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		dbLogger.Error("Failed to create connection pool", zap.Error(err))
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		dbLogger.Error("Database unreachable", zap.Error(err))
		return nil, err
	}

	dbLogger.Info("Successfully connected to PostgreSQL", zap.String("host", cfg.Host))

	return &Database{
		Pool:   pool,
		logger: dbLogger,
	}, nil
}

func (db *Database) Close() {
	if db.Pool != nil {
		db.logger.Info("Closing PostgreSQL connection pool")
		db.Pool.Close()
	}
}
