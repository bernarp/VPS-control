// auth/jwt_test.go
package auth

import (
	"strings"
	"testing"
	"time"

	"VPS-control/internal/config"

	"go.uber.org/zap"
)

func newTestJwtService(ttl time.Duration) *AuthJwtService {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-minimum-32-chars!",
			Issuer: "test-issuer",
			TTL:    ttl,
		},
	}
	return NewAuthJwtService(cfg, zap.NewNop())
}

func testTokenData(username string) TokenData {
	return TokenData{
		UserID:      1,
		Username:    username,
		JTI:         "test-jti-12345",
		Roles:       []string{"user"},
		Permissions: []string{"pm2.view.basic"},
	}
}

func TestGenerateToken(t *testing.T) {
	svc := newTestJwtService(time.Hour)

	t.Run(
		"success", func(t *testing.T) {
			token, err := svc.GenerateToken(testTokenData("testuser"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			parts := strings.Split(token, ".")
			if len(parts) != 3 {
				t.Errorf("invalid token format, got %d parts", len(parts))
			}
		},
	)

	t.Run(
		"empty username", func(t *testing.T) {
			_, err := svc.GenerateToken(
				TokenData{
					UserID:   1,
					Username: "",
					JTI:      "some-jti",
				},
			)
			if err == nil {
				t.Error("expected error for empty username")
			}
		},
	)

	t.Run(
		"empty JTI allowed", func(t *testing.T) {
			token, err := svc.GenerateToken(
				TokenData{
					UserID:   1,
					Username: "testuser",
					JTI:      "",
				},
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token == "" {
				t.Error("expected non-empty token")
			}
		},
	)
}

func TestValidateToken(t *testing.T) {
	svc := newTestJwtService(time.Hour)

	t.Run(
		"valid token", func(t *testing.T) {
			data := testTokenData("testuser")
			token, _ := svc.GenerateToken(data)

			claims, err := svc.ValidateToken(token)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if claims.Username != "testuser" {
				t.Errorf("Username = %q, want %q", claims.Username, "testuser")
			}

			if claims.Issuer != "test-issuer" {
				t.Errorf("Issuer = %q, want %q", claims.Issuer, "test-issuer")
			}

			if claims.UserID != 1 {
				t.Errorf("UserID = %d, want %d", claims.UserID, 1)
			}

			if claims.JTI != data.JTI {
				t.Errorf("JTI = %q, want %q", claims.JTI, data.JTI)
			}

		},
	)

	t.Run(
		"empty token", func(t *testing.T) {
			_, err := svc.ValidateToken("")
			if err == nil {
				t.Error("expected error for empty token")
			}
		},
	)

	t.Run(
		"invalid token", func(t *testing.T) {
			_, err := svc.ValidateToken("invalid.token.here")
			if err == nil {
				t.Error("expected error for invalid token")
			}
		},
	)

	t.Run(
		"tampered token", func(t *testing.T) {
			token, _ := svc.GenerateToken(testTokenData("testuser"))

			parts := strings.Split(token, ".")
			if len(parts) != 3 {
				t.Fatal("invalid token format")
			}

			signature := parts[2]
			if len(signature) > 5 {
				mid := len(signature) / 2
				modified := []byte(signature)
				if modified[mid] == 'a' {
					modified[mid] = 'b'
				} else {
					modified[mid] = 'a'
				}
				parts[2] = string(modified)
			}
			tampered := strings.Join(parts, ".")

			_, err := svc.ValidateToken(tampered)
			if err == nil {
				t.Error("expected error for tampered token")
			}
		},
	)

	t.Run(
		"wrong issuer", func(t *testing.T) {
			otherCfg := &config.Config{
				JWT: config.JWTConfig{
					Secret: "test-secret-key-minimum-32-chars!",
					Issuer: "wrong-issuer",
					TTL:    time.Hour,
				},
			}
			otherSvc := NewAuthJwtService(otherCfg, zap.NewNop())
			token, _ := otherSvc.GenerateToken(testTokenData("testuser"))

			_, err := svc.ValidateToken(token)
			if err == nil {
				t.Error("expected error for wrong issuer")
			}
		},
	)

	t.Run(
		"wrong secret", func(t *testing.T) {
			otherCfg := &config.Config{
				JWT: config.JWTConfig{
					Secret: "different-secret-key-32-chars!!!",
					Issuer: "test-issuer",
					TTL:    time.Hour,
				},
			}
			otherSvc := NewAuthJwtService(otherCfg, zap.NewNop())
			token, _ := otherSvc.GenerateToken(testTokenData("testuser"))

			_, err := svc.ValidateToken(token)
			if err == nil {
				t.Error("expected error for wrong secret")
			}
		},
	)
}

func TestExpiredToken(t *testing.T) {
	svc := newTestJwtService(1 * time.Millisecond)

	token, _ := svc.GenerateToken(testTokenData("testuser"))
	time.Sleep(10 * time.Millisecond)

	_, err := svc.ValidateToken(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestGetIssuer(t *testing.T) {
	svc := newTestJwtService(time.Hour)

	if svc.GetIssuer() != "test-issuer" {
		t.Errorf("GetIssuer() = %q, want %q", svc.GetIssuer(), "test-issuer")
	}
}

func TestGetTTL(t *testing.T) {
	expectedTTL := 2 * time.Hour
	svc := newTestJwtService(expectedTTL)

	if svc.GetTTL() != expectedTTL {
		t.Errorf("GetTTL() = %v, want %v", svc.GetTTL(), expectedTTL)
	}
}

func TestClaimsHasPermission(t *testing.T) {
	claims := &CustomClaims{
		Username:    "testuser",
		UserID:      1,
		JTI:         "test-jti",
		Roles:       []string{"admin", "user"},
		Permissions: []string{"pm2.view.basic", "pm2.control.restart"},
	}

	t.Run(
		"has permission", func(t *testing.T) {
			if !claims.HasPermission("pm2.view.basic") {
				t.Error("should have pm2.view.basic")
			}
		},
	)

	t.Run(
		"no permission", func(t *testing.T) {
			if claims.HasPermission("f2b.control.unban") {
				t.Error("should not have f2b.control.unban")
			}
		},
	)

	t.Run(
		"has any permission", func(t *testing.T) {
			if !claims.HasAnyPermission("f2b.view.status", "pm2.view.basic") {
				t.Error("should have at least one")
			}
		},
	)

	t.Run(
		"has no any permission", func(t *testing.T) {
			if claims.HasAnyPermission("f2b.view.status", "user.delete") {
				t.Error("should not have any")
			}
		},
	)

	t.Run(
		"has role", func(t *testing.T) {
			if !claims.HasRole("admin") {
				t.Error("should have admin role")
			}
		},
	)

	t.Run(
		"no role", func(t *testing.T) {
			if claims.HasRole("superadmin") {
				t.Error("should not have superadmin role")
			}
		},
	)

	t.Run(
		"empty permissions", func(t *testing.T) {
			emptyClaims := &CustomClaims{
				Username:    "testuser",
				Permissions: []string{},
			}
			if emptyClaims.HasPermission("any") {
				t.Error("should not have any permission with empty slice")
			}
			if emptyClaims.HasAnyPermission("any", "other") {
				t.Error("should not have any permission with empty slice")
			}
		},
	)

	t.Run(
		"empty roles", func(t *testing.T) {
			emptyClaims := &CustomClaims{
				Username: "testuser",
				Roles:    []string{},
			}
			if emptyClaims.HasRole("any") {
				t.Error("should not have any role with empty slice")
			}
		},
	)
}
