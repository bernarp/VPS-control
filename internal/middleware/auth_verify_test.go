package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"VPS-control/internal/auth"
	"VPS-control/internal/config"
	"VPS-control/internal/database/sqlite3_local"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func setupAuthServices(t *testing.T) (*auth.AuthJwtService, *auth.AuthCookieService, *sqlite3_local.TokenRepository) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-minimum-32-chars!",
			Issuer: "test-issuer",
			TTL:    time.Hour,
		},
		Cookie: config.CookieConfig{
			Name:     "test_token",
			Secure:   false,
			HttpOnly: true,
			SameSite: "strict",
		},
	}

	jwtSvc := auth.NewAuthJwtService(cfg, zap.NewNop())
	cookieSvc := auth.NewAuthCookieService(cfg)

	tmpFile, err := os.CreateTemp("", "test_tokens_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

	localDB, err := sqlite3_local.NewLocalDB(tmpFile.Name(), zap.NewNop())
	if err != nil {
		t.Fatalf("failed to create local db: %v", err)
	}
	t.Cleanup(func() { localDB.Close() })

	tokenRepo := sqlite3_local.NewTokenRepository(localDB, zap.NewNop())

	return jwtSvc, cookieSvc, tokenRepo
}

func testTokenData(username string) auth.TokenData {
	return auth.TokenData{
		UserID:      1,
		Username:    username,
		JTI:         "test-jti-" + username,
		Roles:       []string{"user"},
		Permissions: []string{"pm2.view.basic", "pm2.control.restart"},
	}
}

func TestAuthMiddleware_ValidCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	data := testTokenData("testuser")
	token, _ := jwtSvc.GenerateToken(data)

	expiresAt := time.Now().Add(time.Hour).Unix()
	_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.AddCookie(
		&http.Cookie{
			Name:  "test_token",
			Value: token,
		},
	)

	middleware := AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())
	middleware(c)

	if c.IsAborted() {
		t.Error("valid cookie should not abort")
	}

	username, exists := c.Get(auth.CtxUsername)
	if !exists || username != "testuser" {
		t.Errorf("username = %v, want 'testuser'", username)
	}

	userID, exists := c.Get(auth.CtxUserID)
	if !exists || userID != 1 {
		t.Errorf("user_id = %v, want 1", userID)
	}

	jti, exists := c.Get(auth.CtxJTI)
	if !exists || jti != data.JTI {
		t.Errorf("jti = %v, want %q", jti, data.JTI)
	}

	permissions, exists := c.Get(auth.CtxPermissions)
	if !exists {
		t.Error("permissions should be set")
	}
	perms := permissions.([]string)
	if len(perms) != 2 {
		t.Errorf("permissions count = %d, want 2", len(perms))
	}

	claims, exists := c.Get(auth.CtxClaims)
	if !exists {
		t.Error("claims should be set")
	}
	if _, ok := claims.(*auth.CustomClaims); !ok {
		t.Error("claims should be *auth.CustomClaims")
	}
}

func TestAuthMiddleware_ValidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	data := testTokenData("testuser")
	token, _ := jwtSvc.GenerateToken(data)
	expiresAt := time.Now().Add(time.Hour).Unix()
	_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())
	middleware(c)

	if c.IsAborted() {
		t.Error("valid bearer token should not abort")
	}
}

func TestAuthMiddleware_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	middleware := AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())
	middleware(c)

	if !c.IsAborted() {
		t.Error("request without auth should be aborted")
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token.here")

	middleware := AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())
	middleware(c)

	if !c.IsAborted() {
		t.Error("invalid token should be aborted")
	}
}

func TestAuthMiddleware_MalformedBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	tests := []struct {
		name   string
		header string
	}{
		{"no_bearer_prefix", "token123"},
		{"wrong_prefix", "Basic token123"},
		{"no_space", "Bearertoken123"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set("Authorization", tt.header)

				middleware := AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())
				middleware(c)

				if !c.IsAborted() {
					t.Error("malformed auth should be aborted")
				}
			},
		)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-minimum-32-chars!",
			Issuer: "test-issuer",
			TTL:    1 * time.Millisecond,
		},
		Cookie: config.CookieConfig{
			Name:     "test_token",
			SameSite: "strict",
		},
	}

	jwtSvc := auth.NewAuthJwtService(cfg, zap.NewNop())
	cookieSvc := auth.NewAuthCookieService(cfg)

	tmpFile, _ := os.CreateTemp("", "test_tokens_expired_*.db")
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	localDB, _ := sqlite3_local.NewLocalDB(tmpFile.Name(), zap.NewNop())
	defer localDB.Close()

	tokenRepo := sqlite3_local.NewTokenRepository(localDB, zap.NewNop())

	data := testTokenData("testuser")
	token, _ := jwtSvc.GenerateToken(data)
	expiresAt := time.Now().Add(time.Hour).Unix()
	_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)

	time.Sleep(10 * time.Millisecond)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())
	middleware(c)

	if !c.IsAborted() {
		t.Error("expired token should be aborted")
	}
}

func TestAuthMiddleware_RevokedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	data := testTokenData("testuser")
	token, _ := jwtSvc.GenerateToken(data)

	expiresAt := time.Now().Add(time.Hour).Unix()
	_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)
	_ = tokenRepo.RevokeToken(data.JTI, 0, "TEST")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())
	middleware(c)

	if !c.IsAborted() {
		t.Error("revoked token should be aborted")
	}
}

func TestAuthMiddleware_TokenNotInRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	data := testTokenData("testuser")
	token, _ := jwtSvc.GenerateToken(data)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())
	middleware(c)

	if !c.IsAborted() {
		t.Error("token not in repo should be aborted")
	}
}

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	t.Run(
		"has permission", func(t *testing.T) {
			data := testTokenData("testuser")
			token, _ := jwtSvc.GenerateToken(data)
			expiresAt := time.Now().Add(time.Hour).Unix()
			_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Authorization", "Bearer "+token)

			AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())(c)
			RequirePermission("pm2.view.basic")(c)

			if c.IsAborted() {
				t.Error("should not abort when user has permission")
			}
		},
	)

	t.Run(
		"no permission", func(t *testing.T) {
			data := auth.TokenData{
				UserID:      2,
				Username:    "testuser2",
				JTI:         "test-jti-user2",
				Roles:       []string{"user"},
				Permissions: []string{"pm2.view.basic"},
			}
			token, _ := jwtSvc.GenerateToken(data)
			expiresAt := time.Now().Add(time.Hour).Unix()
			_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Authorization", "Bearer "+token)

			AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())(c)
			RequirePermission("f2b.control.unban")(c)

			if !c.IsAborted() {
				t.Error("should abort when user lacks permission")
			}
		},
	)

	t.Run(
		"no claims in context", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)

			RequirePermission("any.permission")(c)

			if !c.IsAborted() {
				t.Error("should abort when no claims in context")
			}
		},
	)
}

func TestRequireAnyPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	t.Run(
		"has one of permissions", func(t *testing.T) {
			data := testTokenData("testuser")
			token, _ := jwtSvc.GenerateToken(data)
			expiresAt := time.Now().Add(time.Hour).Unix()
			_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Authorization", "Bearer "+token)

			AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())(c)
			RequireAnyPermission("f2b.view.status", "pm2.view.basic")(c)

			if c.IsAborted() {
				t.Error("should not abort when user has at least one permission")
			}
		},
	)

	t.Run(
		"has none of permissions", func(t *testing.T) {
			data := auth.TokenData{
				UserID:      3,
				Username:    "testuser3",
				JTI:         "test-jti-user3",
				Roles:       []string{"user"},
				Permissions: []string{"pm2.view.basic"},
			}
			token, _ := jwtSvc.GenerateToken(data)
			expiresAt := time.Now().Add(time.Hour).Unix()
			_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Authorization", "Bearer "+token)

			AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())(c)
			RequireAnyPermission("f2b.view.status", "user.delete")(c)

			if !c.IsAborted() {
				t.Error("should abort when user has none of permissions")
			}
		},
	)
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	t.Run(
		"has role", func(t *testing.T) {
			data := testTokenData("testuser")
			token, _ := jwtSvc.GenerateToken(data)
			expiresAt := time.Now().Add(time.Hour).Unix()
			_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Authorization", "Bearer "+token)

			AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())(c)
			RequireRole("user")(c)

			if c.IsAborted() {
				t.Error("should not abort when user has role")
			}
		},
	)

	t.Run(
		"no role", func(t *testing.T) {
			data := auth.TokenData{
				UserID:      4,
				Username:    "testuser4",
				JTI:         "test-jti-user4",
				Roles:       []string{"user"},
				Permissions: []string{"pm2.view.basic"},
			}
			token, _ := jwtSvc.GenerateToken(data)
			expiresAt := time.Now().Add(time.Hour).Unix()
			_ = tokenRepo.SaveToken(data.JTI, data.Username, expiresAt)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Authorization", "Bearer "+token)

			AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())(c)
			RequireRole("admin")(c)

			if !c.IsAborted() {
				t.Error("should abort when user lacks role")
			}
		},
	)
}

func TestGetClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run(
		"claims exist", func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			expectedClaims := &auth.CustomClaims{
				Username: "testuser",
				UserID:   42,
			}
			c.Set(auth.CtxClaims, expectedClaims)

			// Исправлено: вызов через auth.GetClaims
			claims, ok := auth.GetClaims(c)

			if !ok {
				t.Error("should return true when claims exist")
			}
			if claims.Username != "testuser" {
				t.Errorf("Username = %q, want %q", claims.Username, "testuser")
			}
			if claims.UserID != 42 {
				t.Errorf("UserID = %d, want %d", claims.UserID, 42)
			}
		},
	)

	t.Run(
		"claims not exist", func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			claims, ok := auth.GetClaims(c)

			if ok {
				t.Error("should return false when claims not exist")
			}
			if claims != nil {
				t.Error("claims should be nil")
			}
		},
	)

	t.Run(
		"claims wrong type", func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set(auth.CtxClaims, "wrong type")

			claims, ok := auth.GetClaims(c)

			if ok {
				t.Error("should return false for wrong type")
			}
			if claims != nil {
				t.Error("claims should be nil for wrong type")
			}
		},
	)
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run(
		"user_id exists", func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set(auth.CtxUserID, 123)

			// Исправлено: вызов через auth.GetUserID
			userID, ok := auth.GetUserID(c)

			if !ok {
				t.Error("should return true when user_id exists")
			}
			if userID != 123 {
				t.Errorf("userID = %d, want %d", userID, 123)
			}
		},
	)

	t.Run(
		"user_id not exist", func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			userID, ok := auth.GetUserID(c)

			if ok {
				t.Error("should return false when user_id not exist")
			}
			if userID != 0 {
				t.Errorf("userID = %d, want 0", userID)
			}
		},
	)

	t.Run(
		"user_id wrong type", func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set(auth.CtxUserID, "not an int")

			userID, ok := auth.GetUserID(c)

			if ok {
				t.Error("should return false for wrong type")
			}
			if userID != 0 {
				t.Errorf("userID = %d, want 0", userID)
			}
		},
	)
}

func TestGetJTI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run(
		"jti exists", func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set(auth.CtxJTI, "test-jti-value")

			// Исправлено: вызов через auth.GetJTI
			jti, ok := auth.GetJTI(c)

			if !ok {
				t.Error("should return true when jti exists")
			}
			if jti != "test-jti-value" {
				t.Errorf("jti = %q, want %q", jti, "test-jti-value")
			}
		},
	)

	t.Run(
		"jti not exist", func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			jti, ok := auth.GetJTI(c)

			if ok {
				t.Error("should return false when jti not exist")
			}
			if jti != "" {
				t.Errorf("jti = %q, want empty", jti)
			}
		},
	)
}

func TestAuthMiddleware_CookiePriorityOverHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc, cookieSvc, tokenRepo := setupAuthServices(t)

	cookieData := auth.TokenData{
		UserID:      1,
		Username:    "cookie_user",
		JTI:         "cookie-jti",
		Roles:       []string{"user"},
		Permissions: []string{"pm2.view.basic"},
	}
	headerData := auth.TokenData{
		UserID:      2,
		Username:    "header_user",
		JTI:         "header-jti",
		Roles:       []string{"admin"},
		Permissions: []string{"admin.all"},
	}

	cookieToken, _ := jwtSvc.GenerateToken(cookieData)
	headerToken, _ := jwtSvc.GenerateToken(headerData)

	expiresAt := time.Now().Add(time.Hour).Unix()
	_ = tokenRepo.SaveToken(cookieData.JTI, cookieData.Username, expiresAt)
	_ = tokenRepo.SaveToken(headerData.JTI, headerData.Username, expiresAt)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.AddCookie(
		&http.Cookie{
			Name:  "test_token",
			Value: cookieToken,
		},
	)
	c.Request.Header.Set("Authorization", "Bearer "+headerToken)

	middleware := AuthMiddleware(jwtSvc, cookieSvc, tokenRepo, zap.NewNop())
	middleware(c)

	username, _ := c.Get(auth.CtxUsername)
	if username != "cookie_user" {
		t.Errorf("username = %v, want 'cookie_user' (cookie should have priority)", username)
	}
}
