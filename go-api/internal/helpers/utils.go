package helpers

import (
	"errors"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Respond(c *gin.Context, code int, errType, message string) {
	c.JSON(code, gin.H{
		"error":     errType,
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"path":      c.Request.URL.Path,
	})
}

func RandStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		// randomly select 1 character from given charset
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func GetUserID(c *gin.Context) (int64, error) {
	userIDRaw, exists := c.Get("user_id")

	if !exists {
		return 0, errors.New("no user")
	}

	userID, ok := userIDRaw.(int64)
	if !ok {
		return 0, errors.New("invalid type")
	}

	return userID, nil
}
