package postgresql

import "time"

type UserLoginDTO struct {
	Username string `json:"username" binding:"required,min=3,max=32,alphanum"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserResponseDTO struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	LastLogin *time.Time `json:"last_login"`
	Active    bool       `json:"active"`
}

type PermissionDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type RoleDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UserPermissionsDTO struct {
	UserID      int      `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
