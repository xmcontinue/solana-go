package app

import "github.com/gin-gonic/gin"

func NoAuth() gin.HandlerFunc {
	return jAuth.NoAuth()
}
