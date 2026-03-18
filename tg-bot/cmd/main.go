package main

import (
	"os"
	"os/signal"
	"syscall"

	"telegram-bot/internal/bot"
	"telegram-bot/internal/config"
	"telegram-bot/internal/logger"
)

func main() {
	logger.Init()

	cfg := config.Load()

	if cfg.BotToken == "" {
		logger.Error().Msg("BOT_TOKEN is required")
		os.Exit(1)
	}

	tgBot, err := bot.NewBot(cfg)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create bot")
		os.Exit(1)
	}

	go func() {
		if err := tgBot.Start(); err != nil {
			logger.Error().Err(err).Msg("failed to start bot")
			os.Exit(1)
		}
	}()

	logger.Info().Str("name", cfg.BotToken[:10]+"...").Msg("bot started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down...")
	tgBot.Stop()
	logger.LogStop("telegram-bot")
}
