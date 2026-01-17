package postgresql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user inactive")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewUserRepository(
	db *pgxpool.Pool,
	logger *zap.Logger,
) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.Named("user_repository"),
	}
}

func (r *UserRepository) Authenticate(
	ctx context.Context,
	username, rawPassword string,
) (*UserResponseDTO, error) {
	query := `SELECT id, username, password, active, last_login FROM vps_data_auth WHERE username = $1`

	var entity UserEntity
	err := r.db.QueryRow(ctx, query, username).Scan(
		&entity.ID,
		&entity.Username,
		&entity.Password,
		&entity.Active,
		&entity.LastLogin,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("database query error", zap.Error(err))
		return nil, err
	}

	if !entity.Active {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(entity.Password), []byte(rawPassword)); err != nil {
		return nil, ErrInvalidCredentials
	}

	go func() {
		_ = r.UpdateLastLogin(context.Background(), entity.ID)
	}()

	return &UserResponseDTO{
		ID:        entity.ID,
		Username:  entity.Username,
		LastLogin: entity.LastLogin,
	}, nil
}

func (r *UserRepository) UpdateLastLogin(
	ctx context.Context,
	userID int,
) error {
	_, err := r.db.Exec(ctx, "UPDATE vps_data_auth SET last_login = NOW() WHERE id = $1", userID)
	if err != nil {
		r.logger.Error("failed to update last login", zap.Int("user_id", userID), zap.Error(err))
	}
	return err
}
