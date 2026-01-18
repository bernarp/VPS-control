package auth

import (
	"context"

	"VPS-control/internal/database/postgresql"
)

var _ AuthManager = (*ManagerService)(nil)

type ManagerService struct {
	userRepo postgresql.UserStore
	permRepo postgresql.PermissionStore
}

func NewAuthManagerService(
	userRepo postgresql.UserStore,
	permRepo postgresql.PermissionStore,
) *ManagerService {
	return &ManagerService{
		userRepo: userRepo,
		permRepo: permRepo,
	}
}

type AuthResult struct {
	User        *postgresql.UserResponseDTO
	Roles       []string
	Permissions []string
}

func (s *ManagerService) Login(
	ctx context.Context,
	username, password string,
) (*AuthResult, error) {
	user, err := s.userRepo.Authenticate(ctx, username, password)
	if err != nil {
		return nil, err
	}

	roles, err := s.permRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.permRepo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:        user,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func (s *ManagerService) GetUserPermissions(
	ctx context.Context,
	userID int,
) ([]string, error) {
	return s.permRepo.GetUserPermissions(ctx, userID)
}

func (s *ManagerService) HasPermission(
	ctx context.Context,
	userID int,
	permission string,
) (bool, error) {
	return s.permRepo.HasPermission(ctx, userID, permission)
}
