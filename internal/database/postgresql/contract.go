package postgresql

import (
	"context"
)

type DB interface {
	Close()
}

type PermissionStore interface {
	GetUserPermissions(
		ctx context.Context,
		userID int,
	) ([]string, error)
	GetUserRoles(
		ctx context.Context,
		userID int,
	) ([]string, error)
	HasPermission(
		ctx context.Context,
		userID int,
		permissionName string,
	) (bool, error)
	HasAnyPermission(
		ctx context.Context,
		userID int,
		permissionNames []string,
	) (bool, error)
	GetUserFullPermissions(
		ctx context.Context,
		userID int,
	) (*UserPermissionsDTO, error)
	GetAllPermissions(ctx context.Context) ([]PermissionDTO, error)
	GetAllRoles(ctx context.Context) ([]RoleDTO, error)
	AssignRoleToUser(
		ctx context.Context,
		userID int,
		roleName string,
	) error
	RemoveRoleFromUser(
		ctx context.Context,
		userID int,
		roleName string,
	) error
}

type UserStore interface {
	Authenticate(
		ctx context.Context,
		username, rawPassword string,
	) (*UserResponseDTO, error)
	UpdateLastLogin(
		ctx context.Context,
		userID int,
	) error
}
