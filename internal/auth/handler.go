package auth

import (
	"errors"
	"net/http"
	"time"

	"VPS-control/internal/apierror"
	"VPS-control/internal/database/postgresql"
	"VPS-control/internal/database/sqlite3_local"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type handler struct {
	authManager   AuthManager
	jwtService    JwtProvider
	cookieService SetAuthCookie
	tokenRepo     sqlite3_local.TokenStore
	logger        *zap.Logger
}

func NewHandler(
	am AuthManager,
	aj JwtProvider,
	ac SetAuthCookie,
	tr sqlite3_local.TokenStore,
	l *zap.Logger,
) Handler {
	return &handler{
		authManager:   am,
		jwtService:    aj,
		cookieService: ac,
		tokenRepo:     tr,
		logger:        l,
	}
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and set HTTP-only cookie with JWT token containing permissions
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login Credentials"
// @Success      200 {object} LoginResponse
// @Failure      400 {object} apierror.AppError
// @Failure      401 {object} apierror.AppError
// @Failure      500 {object} apierror.AppError
// @Router       /auth/login [post]
func (h *handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST)
		return
	}

	result, err := h.authManager.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) ||
			errors.Is(err, postgresql.ErrInvalidCredentials) ||
			errors.Is(err, postgresql.ErrUserInactive) {
			apierror.Abort(c, apierror.Errors.INVALID_CREDENTIALS)
			return
		}
		apierror.Abort(c, apierror.Errors.DATABASE_ERROR.Wrap(err))
		return
	}

	jti := h.tokenRepo.GenerateJTI(result.User.Username)
	expiresAt := time.Now().Add(h.jwtService.GetTTL()).Unix()

	token, err := h.jwtService.GenerateToken(
		TokenData{
			UserID:      result.User.ID,
			Username:    result.User.Username,
			JTI:         jti,
			Roles:       result.Roles,
			Permissions: result.Permissions,
		},
	)
	if err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}

	revokedCount, err := h.tokenRepo.SaveTokenExclusive(jti, result.User.Username, expiresAt)
	if err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}

	h.cookieService.SetAuthCookie(c, token)

	h.logger.Info(
		LogUserLoggedIn,
		zap.String("username", result.User.Username),
		zap.Int("user_id", result.User.ID),
		zap.Int64("revoked_sessions", revokedCount),
	)

	c.JSON(http.StatusOK, LoginResponse{Success: true, Message: MsgLoginSuccess})
}

// Verify godoc
// @Summary      Verify session
// @Description  Check if the user has a valid JWT in cookies/header and return info
// @Tags         auth
// @Security     CookieAuth
// @Produce      json
// @Success      200 {object} AuthStatusResponse
// @Failure      401 {object} apierror.AppError
// @Router       /auth/verify [post]
func (h *handler) Verify(c *gin.Context) {
	claims, ok := GetClaims(c)
	if !ok {
		apierror.Abort(c, apierror.Errors.TOKEN_EXPIRED)
		return
	}

	c.JSON(
		http.StatusOK, AuthStatusResponse{
			Success:  true,
			Message:  MsgSessionOK,
			Username: claims.Username,
		},
	)
}

// Logout godoc
// @Summary      User logout
// @Description  Clear authentication cookie and revoke token
// @Tags         auth
// @Security     CookieAuth
// @Produce      json
// @Success      200 {object} AuthStatusResponse
// @Failure      401 {object} apierror.AppError
// @Router       /auth/logout [post]
func (h *handler) Logout(c *gin.Context) {
	claims, ok := GetClaims(c)
	if !ok {
		apierror.Abort(c, apierror.Errors.TOKEN_EXPIRED)
		return
	}
	if err := h.tokenRepo.RevokeToken(claims.JTI, claims.UserID, claims.Username); err != nil {
		h.logger.Warn("Failed to revoke token", zap.String("jti", claims.JTI), zap.Error(err))
	}

	h.cookieService.ClearAuthCookie(c)
	c.JSON(http.StatusOK, AuthStatusResponse{Success: true, Message: MsgLogoutSuccess})
}

// GetSessions godoc
// @Summary      Get all auth sessions
// @Description  Returns a list of all generated tokens (active and revoked) from local SQLite DB
// @Tags         auth
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  SessionListResponse
// @Failure      500  {object}  apierror.AppError
// @Router       /auth/sessions [get]
func (h *handler) GetSessions(c *gin.Context) {
	entities, err := h.tokenRepo.GetAllTokens()
	if err != nil {
		apierror.Abort(c, apierror.Errors.DATABASE_ERROR.Wrap(err))
		return
	}

	sessions := make([]SessionResponse, 0, len(entities))
	for _, e := range entities {
		resp := SessionResponse{
			ID:        e.ID,
			JTI:       e.JTI,
			Username:  e.Username,
			Revoked:   e.Revoked,
			ExpiresAt: e.ExpiresAt,
			CreatedAt: e.CreatedAt,
		}
		if e.RevokedByID.Valid {
			val := e.RevokedByID.Int64
			resp.RevokedByID = &val
		}
		if e.RevokedByUsername.Valid {
			resp.RevokedByUsername = e.RevokedByUsername.String
		}
		sessions = append(sessions, resp)
	}

	c.JSON(http.StatusOK, SessionListResponse{Sessions: sessions, Total: len(sessions)})
}

// RevokeSession godoc
// @Summary      Revoke an auth session
// @Description  Invalidates a specific session by its JTI
// @Tags         auth
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        request body RevokeSessionRequest true "Session JTI"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  apierror.AppError
// @Failure      404  {object}  apierror.AppError
// @Failure      500  {object}  apierror.AppError
// @Router       /auth/sessions/revoke [post]
func (h *handler) RevokeSession(c *gin.Context) {
	var req RevokeSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST)
		return
	}

	adminID, _ := GetUserID(c)
	adminUsername, _ := c.Get(CtxUsername)
	usernameStr, ok := adminUsername.(string)
	if !ok {
		usernameStr = "UNKNOWN"
	}

	err := h.tokenRepo.RevokeToken(req.JTI, adminID, usernameStr)
	if err != nil {
		if errors.Is(err, sqlite3_local.ErrTokenNotFound) {
			apierror.Abort(c, apierror.Errors.INVALID_REQUEST.Wrap(err))
			return
		}
		apierror.Abort(c, apierror.Errors.DATABASE_ERROR.Wrap(err))
		return
	}

	h.logger.Info("Session revoked", zap.String("jti", req.JTI), zap.String("by", usernameStr))
	c.JSON(http.StatusOK, gin.H{"message": "session revoked successfully"})
}
