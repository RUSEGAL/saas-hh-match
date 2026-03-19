package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	NATSURL              string  `mapstructure:"nats_url"`
	APIWebhookURL        string  `mapstructure:"api_webhook_url"`
	APIWebhookMatchesURL string  `mapstructure:"api_webhook_matches_url"`
	AIAPIKey             string  `mapstructure:"ai_api_key"`
	AIModel              string  `mapstructure:"ai_model"`
	AIBaseURL            string  `mapstructure:"ai_base_url"`
	CacheSize            int     `mapstructure:"cache_size"`
	ResumeWorkers        int     `mapstructure:"resume_workers"`
	VacancyWorkers       int     `mapstructure:"vacancy_workers"`
	MainAPIURL           string  `mapstructure:"main_api_url"`
	HHAPIURL             string  `mapstructure:"hh_api_url"`
	MatchScoreThreshold  float64 `mapstructure:"match_score_threshold"`
	MatchBatchSize       int     `mapstructure:"match_batch_size"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app")

	viper.SetDefault("nats_url", "nats://localhost:4222")
	viper.SetDefault("api_webhook_url", "http://localhost:8080/ai/webhook/analyze")
	viper.SetDefault("api_webhook_matches_url", "http://localhost:8080/ai/webhook/matches")
	viper.SetDefault("ai_model", "deepseek-chat")
	viper.SetDefault("ai_base_url", "https://llms.dotpoin.com/v1/")
	viper.SetDefault("cache_size", 1000)
	viper.SetDefault("resume_workers", 0)
	viper.SetDefault("vacancy_workers", 2)
	viper.SetDefault("main_api_url", "http://localhost:8080")
	viper.SetDefault("hh_api_url", "https://api.hh.ru")
	viper.SetDefault("match_score_threshold", 0.70)
	viper.SetDefault("match_batch_size", 20)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	if err := viper.BindEnv("nats_url", "NATS_URL"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("api_webhook_url", "API_WEBHOOK_URL"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("ai_api_key", "AI_API_KEY"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("ai_model", "AI_MODEL"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("ai_base_url", "AI_BASE_URL"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("cache_size", "CACHE_SIZE"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("resume_workers", "RESUME_WORKERS"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("vacancy_workers", "VACANCY_WORKERS"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("api_webhook_matches_url", "API_WEBHOOK_MATCHES_URL"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("main_api_url", "MAIN_API_URL"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("hh_api_url", "HH_API_URL"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("match_score_threshold", "MATCH_SCORE_THRESHOLD"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("match_batch_size", "MATCH_BATCH_SIZE"); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
