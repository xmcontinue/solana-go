package auth

import "github.com/gin-gonic/gin"

// Authenticator 网关认证接口
type Authenticator interface {
	Name() string
	NoAuth() gin.HandlerFunc
}
