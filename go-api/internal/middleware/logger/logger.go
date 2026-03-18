package logger

import (
	"go-api/internal/logger"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer logger.LogRequest(c)()
		c.Next()
	}
}
