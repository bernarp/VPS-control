package sqlite3_local

import (
	"database/sql"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

type LocalDB struct {
	DB     *sql.DB
	logger *zap.Logger
}

func NewLocalDB(
	dbPath string,
	logger *zap.Logger,
) (*LocalDB, error) {
	log := logger.Named("sqlite_local")

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		log.Error("Failed to create db directory", zap.Error(err))
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Error("Failed to open sqlite database", zap.Error(err))
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Error("Failed to ping sqlite database", zap.Error(err))
		return nil, err
	}

	localDB := &LocalDB{
		DB:     db,
		logger: log,
	}

	if err := localDB.initSchema(); err != nil {
		log.Error("Failed to initialize schema", zap.Error(err))
		return nil, err
	}

	log.Info("SQLite local database connected", zap.String("path", dbPath))
	return localDB, nil
}

func (l *LocalDB) initSchema() error {
	schema := `
    CREATE TABLE IF NOT EXISTS tokens (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        jti TEXT NOT NULL UNIQUE,
        username TEXT NOT NULL,
        revoked INTEGER NOT NULL DEFAULT 0,
        revoked_by_id INTEGER,
        revoked_by_username TEXT,
        expires_at INTEGER NOT NULL,
        created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
    );

    CREATE INDEX IF NOT EXISTS idx_tokens_jti ON tokens(jti);
    CREATE INDEX IF NOT EXISTS idx_tokens_username ON tokens(username);
    CREATE INDEX IF NOT EXISTS idx_tokens_revoked ON tokens(revoked);
    `

	_, err := l.DB.Exec(schema)
	return err
}

func (l *LocalDB) Close() {
	if l.DB != nil {
		l.logger.Info("Closing SQLite connection")
		_ = l.DB.Close()
	}
}
