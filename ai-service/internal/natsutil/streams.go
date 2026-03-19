package natsutil

import (
	"fmt"

	"github.com/nats-io/nats.go"

	"ai-service/internal/logger"
)

var streams = []struct {
	Name     string
	Subjects []string
}{
	{Name: "RESUME_ANALYZE", Subjects: []string{"resume.analyze"}},
	{Name: "VACANCY_MATCH", Subjects: []string{"vacancy.match"}},
}

func EnsureStreams(nc *nats.Conn) error {
	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("failed to get JetStream context: %w", err)
	}

	for _, stream := range streams {
		info, err := js.StreamInfo(stream.Name)
		if err == nil && info != nil {
			logger.Info().Str("stream", stream.Name).Msg("stream already exists")
			continue
		}

		cfg := &nats.StreamConfig{
			Name:      stream.Name,
			Subjects:  stream.Subjects,
			Retention: nats.WorkQueuePolicy,
			Storage:   nats.FileStorage,
		}

		_, err = js.AddStream(cfg)
		if err != nil {
			return fmt.Errorf("failed to create stream %s: %w", stream.Name, err)
		}
		logger.Info().Str("stream", stream.Name).Msg("stream created")
	}

	return nil
}
