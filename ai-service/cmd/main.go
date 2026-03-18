package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-service/internal/ai"
	"ai-service/internal/config"
	"ai-service/internal/consumer"
	"ai-service/internal/logger"
	"ai-service/internal/vacancy"
	"ai-service/internal/webhook"
)

func main() {
	logger.Init()
	logger.LogStart("ai-service")

	cfg, err := config.Load()
	if err != nil {
		logger.Error().Err(err).Msg("failed to load config")
		os.Exit(1)
	}

	if cfg.AIAPIKey == "" {
		logger.Error().Msg("AI_API_KEY is required")
		os.Exit(1)
	}

	nc, err := consumer.ReconnectWithBackoff(cfg.NATSURL, 10, 30*time.Second)
	if err != nil {
		logger.Error().Err(err).Msg("failed to connect to NATS")
		os.Exit(1)
	}
	defer nc.Close()

	logger.Info().Str("url", cfg.NATSURL).Msg("connected to NATS")

	analyzer, err := ai.NewAnalyzer(cfg.AIAPIKey, cfg.AIModel, cfg.AIBaseURL, "prompts/resume_analyze.txt", cfg.CacheSize)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create AI analyzer")
		os.Exit(1)
	}

	sender := webhook.NewSender(cfg.APIWebhookURL)

	natsConsumer := consumer.NewConsumer(nc, analyzer, sender, cfg.ResumeWorkers, cfg.VacancyWorkers)

	if cfg.MainAPIURL != "" && cfg.HHAPIURL != "" {
		vacancyFetcher := vacancy.NewFetcher(cfg.HHAPIURL, 5*time.Minute)
		vacancyMatcher, err := vacancy.NewMatcher(cfg.AIAPIKey, cfg.AIModel, cfg.AIBaseURL, "prompts/vacancy_match.txt", cfg.MatchScoreThreshold, cfg.MatchBatchSize)
		if err != nil {
			logger.Error().Err(err).Msg("failed to create vacancy matcher")
			os.Exit(1)
		}
		matchesSender := webhook.NewMatchesSender(cfg.APIWebhookMatchesURL)

		natsConsumer = natsConsumer.WithVacancyService(vacancyFetcher, vacancyMatcher, cfg.MainAPIURL, matchesSender)
		logger.Info().Int("batchSize", cfg.MatchBatchSize).Msg("vacancy matching service enabled")
	}

	if err := natsConsumer.Start(); err != nil {
		logger.Error().Err(err).Msg("failed to start NATS consumer")
		os.Exit(1)
	}

	logger.Info().
		Int("resume_workers", cfg.ResumeWorkers).
		Int("vacancy_workers", cfg.VacancyWorkers).
		Msg("ai-service started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info().Msg("shutting down...")
	natsConsumer.Stop()
	logger.LogStop("ai-service")
}
