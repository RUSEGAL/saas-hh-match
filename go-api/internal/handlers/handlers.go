package handler404

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Ret404(c *gin.Context) {
	c.JSON(404, gin.H{
		"error":     "route_not_found",
		"message":   "route does not exist",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"path":      c.Request.URL.Path,
	})
	log.Info()
}
