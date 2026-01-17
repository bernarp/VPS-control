package sqlite3_local

import "database/sql"

type TokenEntity struct {
	ID                int64          `db:"id"`
	JTI               string         `db:"jti"`
	Username          string         `db:"username"`
	Revoked           bool           `db:"revoked"`
	RevokedByID       sql.NullInt64  `db:"revoked_by_id"`
	RevokedByUsername sql.NullString `db:"revoked_by_username"`
	ExpiresAt         int64          `db:"expires_at"`
	CreatedAt         int64          `db:"created_at"`
}

type TokenStatus struct {
	Revoked   bool
	ExpiresAt int64
}

type RevokeResult struct {
	Count int64
}
