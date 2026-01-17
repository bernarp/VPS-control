package auth

import (
	"github.com/gin-gonic/gin"
)

const (
	CtxUsername    = "username"
	CtxUserID      = "user_id"
	CtxJTI         = "jti"
	CtxRoles       = "roles"
	CtxPermissions = "permissions"
	CtxClaims      = "claims"
)

func GetClaims(c *gin.Context) (*CustomClaims, bool) {
	claims, exists := c.Get(CtxClaims)
	if !exists {
		return nil, false
	}
	customClaims, ok := claims.(*CustomClaims)
	return customClaims, ok
}

func GetUserID(c *gin.Context) (int, bool) {
	id, exists := c.Get(CtxUserID)
	if !exists {
		return 0, false
	}
	userID, ok := id.(int)
	return userID, ok
}

func GetJTI(c *gin.Context) (string, bool) {
	jti, exists := c.Get(CtxJTI)
	if !exists {
		return "", false
	}
	val, ok := jti.(string)
	return val, ok
}
