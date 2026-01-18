package auth

import (
	"errors"
	"fmt"
	"time"

	"VPS-control/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var _ JwtProvider = (*AuthJwtService)(nil)

type AuthJwtService struct {
	secretKey []byte
	issuer    string
	ttl       time.Duration
	logger    *zap.Logger
}

type CustomClaims struct {
	Username    string   `json:"username"`
	UserID      int      `json:"uid"`
	JTI         string   `json:"jti"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

type TokenData struct {
	UserID      int
	Username    string
	JTI         string
	Roles       []string
	Permissions []string
}

func NewAuthJwtService(
	cfg *config.Config,
	logger *zap.Logger,
) *AuthJwtService {
	if cfg.JWT.Secret == "" {
		logger.Fatal("JWT_SECRET environment variable is not set")
	}

	return &AuthJwtService{
		secretKey: []byte(cfg.JWT.Secret),
		issuer:    cfg.JWT.Issuer,
		ttl:       cfg.JWT.TTL,
		logger:    logger.Named("jwt"),
	}
}

func (s *AuthJwtService) GenerateToken(data TokenData) (string, error) {
	if data.Username == "" {
		s.logger.Warn("Token generation failed: empty username")
		return "", errors.New("username cannot be empty")
	}

	now := time.Now()
	claims := CustomClaims{
		Username:    data.Username,
		UserID:      data.UserID,
		JTI:         data.JTI,
		Roles:       data.Roles,
		Permissions: data.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        data.JTI,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   data.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		s.logger.Error(
			"Failed to sign token",
			zap.String("user", data.Username),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	s.logger.Debug(
		"Token generated successfully",
		zap.String("user", data.Username),
		zap.Int("user_id", data.UserID),
		zap.String("jti", data.JTI),
		zap.Strings("roles", data.Roles),
		zap.Int("permissions_count", len(data.Permissions)),
		zap.Time("expires_at", claims.ExpiresAt.Time),
	)

	return signedToken, nil
}

func (s *AuthJwtService) ValidateToken(tokenString string) (*CustomClaims, error) {
	if tokenString == "" {
		s.logger.Warn("Token validation failed: empty token")
		return nil, errors.New("token is empty")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				s.logger.Warn(
					"Unexpected signing method",
					zap.String("method", fmt.Sprintf("%v", token.Header["alg"])),
				)
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.secretKey, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(s.issuer),
		jwt.WithLeeway(5*time.Second),
	)

	if err != nil {
		s.logger.Warn("Token parse error", zap.Error(err))
		return nil, fmt.Errorf("token parse error: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		s.logger.Warn("Invalid token claims")
		return nil, errors.New("invalid token claims")
	}

	now := time.Now()
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(now) {
		s.logger.Warn(
			"Token expired",
			zap.String("user", claims.Username),
			zap.Time("expired_at", claims.ExpiresAt.Time),
		)
		return nil, errors.New("token expired")
	}

	if claims.NotBefore != nil && claims.NotBefore.After(now) {
		s.logger.Warn(
			"Token not yet valid",
			zap.String("user", claims.Username),
			zap.Time("valid_from", claims.NotBefore.Time),
		)
		return nil, errors.New("token not yet valid")
	}

	s.logger.Debug(
		"Token validated successfully",
		zap.String("user", claims.Username),
		zap.Int("user_id", claims.UserID),
		zap.String("jti", claims.JTI),
	)

	return claims, nil
}

func (c *CustomClaims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func (c *CustomClaims) HasAnyPermission(permissions ...string) bool {
	for _, required := range permissions {
		for _, p := range c.Permissions {
			if p == required {
				return true
			}
		}
	}
	return false
}

func (c *CustomClaims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (s *AuthJwtService) GetIssuer() string {
	return s.issuer
}

func (s *AuthJwtService) GetTTL() time.Duration {
	return s.ttl
}
