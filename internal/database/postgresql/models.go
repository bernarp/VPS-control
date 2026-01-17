package postgresql

import "time"

type UserEntity struct {
	ID        int        `db:"id"`
	Username  string     `db:"username"`
	Password  string     `db:"password"`
	LastLogin *time.Time `db:"last_login"`
	Active    bool       `db:"active"`
}

type PermissionEntity struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

type RoleEntity struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}
