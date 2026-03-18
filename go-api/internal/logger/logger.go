package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}
	output.FormatLevel = func(i interface{}) string {
		return fmt.Sprintf("%-3s", i)
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	log.Logger = log.Output(output)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func Info() *zerolog.Event {
	return log.Info()
}

func Error() *zerolog.Event {
	return log.Error()
}

func Debug() *zerolog.Event {
	return log.Debug()
}

func Warn() *zerolog.Event {
	return log.Warn()
}

func Request(method, path string, status int, latency time.Duration, ip, userID string) {
	log.Info().
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Dur("latency", latency).
		Str("ip", ip).
		Str("user_id", userID).
		Msg("request")
}

func LogRequest(c *gin.Context) func() {
	start := time.Now()
	return func() {
		var userID string
		if id, exists := c.Get("user_id"); exists {
			userID = fmt.Sprintf("%v", id)
		}
		Request(
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
			c.ClientIP(),
			userID,
		)
	}
}
