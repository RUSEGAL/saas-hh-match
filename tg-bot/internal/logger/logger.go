package logger

import (
	"fmt"
	"os"

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

func Fatal() *zerolog.Event {
	return log.Fatal()
}

func LogStart(name string) {
	log.Info().Str("service", name).Msg("started")
}

func LogStop(name string) {
	log.Info().Str("service", name).Msg("stopped")
}

func LogError(err error, msg string) {
	log.Error().Err(err).Msg(msg)
}

func LogRequest(action string, userID int64, success bool) {
	event := log.Info()
	if !success {
		event = log.Error()
	}
	event.
		Str("action", action).
		Int64("user_id", userID).
		Bool("success", success).
		Msg("request")
}
