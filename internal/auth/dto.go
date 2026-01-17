package auth

type LoginRequest struct {
	Username string `json:"username" example:"admin" binding:"required,min=3,max=32,alphanum"`
	Password string `json:"password" example:"secret_pass" binding:"required,min=8,max=128,excludes= "`
}

type LoginResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Logged in successfully"`
}

type AuthStatusResponse struct {
	Success  bool   `json:"success" example:"true"`
	Message  string `json:"message" example:"ok"`
	Username string `json:"username,omitempty" example:"bernarp"`
}

type SessionResponse struct {
	ID                int64  `json:"id"`
	JTI               string `json:"jti"`
	Username          string `json:"username"`
	Revoked           bool   `json:"revoked"`
	RevokedByID       *int64 `json:"revoked_by_id,omitempty"`
	RevokedByUsername string `json:"revoked_by_username,omitempty"`
	ExpiresAt         int64  `json:"expires_at"`
	CreatedAt         int64  `json:"created_at"`
}

type SessionListResponse struct {
	Sessions []SessionResponse `json:"sessions"`
	Total    int               `json:"total"`
}

type RevokeSessionRequest struct {
	JTI string `json:"jti" binding:"required"`
}
