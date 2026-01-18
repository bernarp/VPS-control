package sqlite3_local

type LocalDatabase interface {
	Close()
}

type TokenStore interface {
	GenerateJTI(username string) string
	SaveToken(
		jti, username string,
		expiresAt int64,
	) error
	SaveTokenExclusive(
		jti, username string,
		expiresAt int64,
	) (int64, error)
	ValidateToken(jti string) error
	RevokeToken(
		jti string,
		byID int,
		byUsername string,
	) error
	GetAllTokens() ([]TokenEntity, error)
}
