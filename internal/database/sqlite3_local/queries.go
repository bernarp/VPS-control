package sqlite3_local

const (
	QueryInsertToken = `INSERT INTO tokens (jti, username, revoked, expires_at) VALUES (?, ?, 0, ?)`

	QuerySelectRevoked = `SELECT revoked FROM tokens WHERE jti = ?`

	QuerySelectTokenValidation = `SELECT revoked, expires_at FROM tokens WHERE jti = ?`

	QueryRevokeToken = `UPDATE tokens SET revoked = 1, revoked_by_id = ?, revoked_by_username = ? WHERE jti = ?` //nolint:gosec // SQL query, not credentials

	QueryRevokeAllUserTokens = `UPDATE tokens SET revoked = 1, revoked_by_id = ?, revoked_by_username = ? WHERE username = ? AND revoked = 0` //nolint:gosec // SQL query, not credentials

	QueryDeleteExpired = `DELETE FROM tokens WHERE expires_at < ?`

	QueryCountActiveTokens = `SELECT COUNT(*) FROM tokens WHERE username = ? AND revoked = 0 AND expires_at > ?` //nolint:gosec // SQL query, not credentials

	QuerySelectAllTokens = `SELECT id, jti, username, revoked, revoked_by_id, revoked_by_username, expires_at, created_at FROM tokens ORDER BY created_at DESC` //nolint:gosec // SQL query, not credentials
)
