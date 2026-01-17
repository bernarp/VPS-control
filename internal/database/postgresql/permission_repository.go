package postgresql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrPermissionNotFound = errors.New("permission not found")
	ErrRoleNotFound       = errors.New("role not found")
)

type PermissionRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewPermissionRepository(
	db *pgxpool.Pool,
	logger *zap.Logger,
) *PermissionRepository {
	return &PermissionRepository{
		db:     db,
		logger: logger.Named("permission_repository"),
	}
}

func (r *PermissionRepository) GetUserPermissions(
	ctx context.Context,
	userID int,
) ([]string, error) {
	query := `
        SELECT DISTINCT p.name
        FROM permissions p
        JOIN role_permissions rp ON p.id = rp.permission_id
        JOIN user_roles ur ON rp.role_id = ur.role_id
        WHERE ur.user_id = $1
        ORDER BY p.name
    `

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get user permissions", zap.Int("user_id", userID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		permissions = append(permissions, name)
	}

	return permissions, nil
}

func (r *PermissionRepository) GetUserRoles(
	ctx context.Context,
	userID int,
) ([]string, error) {
	query := `
        SELECT r.name
        FROM roles r
        JOIN user_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = $1
        ORDER BY r.name
    `

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get user roles", zap.Int("user_id", userID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}

	return roles, nil
}

func (r *PermissionRepository) HasPermission(
	ctx context.Context,
	userID int,
	permissionName string,
) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM permissions p
            JOIN role_permissions rp ON p.id = rp.permission_id
            JOIN user_roles ur ON rp.role_id = ur.role_id
            WHERE ur.user_id = $1 AND p.name = $2
        )
    `

	var exists bool
	err := r.db.QueryRow(ctx, query, userID, permissionName).Scan(&exists)
	if err != nil {
		r.logger.Error(
			"failed to check permission",
			zap.Int("user_id", userID),
			zap.String("permission", permissionName),
			zap.Error(err),
		)
		return false, err
	}

	return exists, nil
}

func (r *PermissionRepository) HasAnyPermission(
	ctx context.Context,
	userID int,
	permissionNames []string,
) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM permissions p
            JOIN role_permissions rp ON p.id = rp.permission_id
            JOIN user_roles ur ON rp.role_id = ur.role_id
            WHERE ur.user_id = $1 AND p.name = ANY($2)
        )
    `

	var exists bool
	err := r.db.QueryRow(ctx, query, userID, permissionNames).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check permissions", zap.Int("user_id", userID), zap.Error(err))
		return false, err
	}

	return exists, nil
}

func (r *PermissionRepository) GetUserFullPermissions(
	ctx context.Context,
	userID int,
) (*UserPermissionsDTO, error) {
	var username string
	err := r.db.QueryRow(ctx, "SELECT username FROM vps_data_auth WHERE id = $1", userID).Scan(&username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	roles, err := r.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissions, err := r.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserPermissionsDTO{
		UserID:      userID,
		Username:    username,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func (r *PermissionRepository) GetAllPermissions(
	ctx context.Context,
) ([]PermissionDTO, error) {
	query := `SELECT id, name, COALESCE(description, '') FROM permissions ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to get all permissions", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var permissions []PermissionDTO
	for rows.Next() {
		var p PermissionDTO
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	return permissions, nil
}

func (r *PermissionRepository) GetAllRoles(
	ctx context.Context,
) ([]RoleDTO, error) {
	query := `SELECT id, name, COALESCE(description, '') FROM roles ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to get all roles", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var roles []RoleDTO
	for rows.Next() {
		var role RoleDTO
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *PermissionRepository) AssignRoleToUser(
	ctx context.Context,
	userID int,
	roleName string,
) error {
	query := `
        INSERT INTO user_roles (user_id, role_id)
        SELECT $1, id FROM roles WHERE name = $2
        ON CONFLICT DO NOTHING
    `

	result, err := r.db.Exec(ctx, query, userID, roleName)
	if err != nil {
		r.logger.Error(
			"failed to assign role",
			zap.Int("user_id", userID),
			zap.String("role", roleName),
			zap.Error(err),
		)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrRoleNotFound
	}

	r.logger.Info(
		"role assigned",
		zap.Int("user_id", userID),
		zap.String("role", roleName),
	)

	return nil
}

func (r *PermissionRepository) RemoveRoleFromUser(
	ctx context.Context,
	userID int,
	roleName string,
) error {
	query := `
        DELETE FROM user_roles
        WHERE user_id = $1 AND role_id = (SELECT id FROM roles WHERE name = $2)
    `

	_, err := r.db.Exec(ctx, query, userID, roleName)
	if err != nil {
		r.logger.Error(
			"failed to remove role",
			zap.Int("user_id", userID),
			zap.String("role", roleName),
			zap.Error(err),
		)
		return err
	}

	return nil
}
