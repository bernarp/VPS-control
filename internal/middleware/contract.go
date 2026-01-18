package middleware

import "github.com/gin-gonic/gin"

type Sanitizer interface {
	Middleware() gin.HandlerFunc
}
