package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-api/internal/cache"
	"go-api/internal/service/tokens"

	"github.com/gin-gonic/gin"
)

const tokenCacheTTL = 15 * time.Minute

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing Authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid Authorization format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		cacheKey := fmt.Sprintf("token:%s", token)

		if cache.Client != nil {
			var userID int64
			if err := cache.Get(cacheKey, &userID); err == nil {
				c.Set("user_id", userID)
				c.Next()
				return
			}
		}

		id, err := tokens.FindToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "auth_error",
				"message": "invalid token",
			})
			c.Abort()
			return
		}

		userID := int64(id)

		if cache.Client != nil {
			cache.Set(cacheKey, userID, tokenCacheTTL)
		}

		c.Set("user_id", userID)

		c.Next()
	}
}

func InvalidateToken(token string) {
	if cache.Client != nil {
		cache.Delete(fmt.Sprintf("token:%s", token))
	}
}
