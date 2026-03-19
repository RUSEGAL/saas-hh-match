package helpers

import (
	"errors"
	"math/rand"
	"strconv"
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
	if userIDRaw, exists := c.Get("user_id"); exists {
		if userID, ok := userIDRaw.(int64); ok {
			return userID, nil
		}
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			return userID, nil
		}
	}

	var body struct {
		UserID int64 `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&body); err == nil && body.UserID > 0 {
		return body.UserID, nil
	}

	return 0, errors.New("no user_id found in request")
}
