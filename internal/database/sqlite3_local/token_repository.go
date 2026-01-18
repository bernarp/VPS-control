package sqlite3_local

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenRevoked  = errors.New("token revoked")
)

const (
	jtiRandomBytesLen = 8
	jtiFormat         = "%s_%d%s"
)

var _ TokenStore = (*TokenRepository)(nil)

type TokenRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewTokenRepository(
	localDB *LocalDB,
	logger *zap.Logger,
) *TokenRepository {
	return &TokenRepository{
		db:     localDB.DB,
		logger: logger.Named("token_repository"),
	}
}

func (r *TokenRepository) GenerateJTI(username string) string {
	randomBytes := make([]byte, jtiRandomBytesLen)
	_, _ = rand.Read(randomBytes)
	return fmt.Sprintf(jtiFormat, username, time.Now().Unix(), hex.EncodeToString(randomBytes))
}

func (r *TokenRepository) SaveToken(
	jti, username string,
	expiresAt int64,
) error {
	_, err := r.db.Exec(QueryInsertToken, jti, username, expiresAt)
	return err
}

func (r *TokenRepository) SaveTokenExclusive(
	jti, username string,
	expiresAt int64,
) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			r.logger.Warn("rollback failed", zap.Error(rbErr))
		}
	}()

	result, err := tx.Exec(QueryRevokeAllUserTokens, 0, "SYSTEM", username)
	if err != nil {
		return 0, err
	}
	revokedCount, _ := result.RowsAffected()

	_, err = tx.Exec(QueryInsertToken, jti, username, expiresAt)
	if err != nil {
		return 0, err
	}

	return revokedCount, tx.Commit()
}

func (r *TokenRepository) ValidateToken(jti string) error {
	var status TokenStatus
	err := r.db.QueryRow(QuerySelectTokenValidation, jti).Scan(&status.Revoked, &status.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTokenNotFound
		}
		return err
	}
	if status.Revoked || time.Now().Unix() > status.ExpiresAt {
		return ErrTokenRevoked
	}
	return nil
}

func (r *TokenRepository) RevokeToken(
	jti string,
	byID int,
	byUsername string,
) error {
	result, err := r.db.Exec(QueryRevokeToken, byID, byUsername, jti)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrTokenNotFound
	}
	return nil
}

func (r *TokenRepository) GetAllTokens() ([]TokenEntity, error) {
	rows, err := r.db.Query(QuerySelectAllTokens)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tokens []TokenEntity
	for rows.Next() {
		var t TokenEntity
		var revoked int
		err := rows.Scan(
			&t.ID, &t.JTI, &t.Username, &revoked,
			&t.RevokedByID, &t.RevokedByUsername,
			&t.ExpiresAt, &t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		t.Revoked = revoked == 1
		tokens = append(tokens, t)
	}
	return tokens, nil
}
