package config

import (
	"os"
	"strconv"
)

type Config struct {
	BotToken          string
	APIURL            string
	APIToken          string
	PaymentProvider   string
	YookassaShopID    string
	YookassaSecret    string
	SchedulerEnabled  bool
	Timezone          string
	SubscriptionDays  int
	SubscriptionPrice int
	DatabaseURL       string

	Mode          string
	WebhookURL    string
	WebhookListen string
	WebhookSecret string

	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() *Config {
	return &Config{
		BotToken:          getEnv("BOT_TOKEN", ""),
		APIURL:            getEnv("API_URL", "http://localhost:8080"),
		APIToken:          getEnv("API_TOKEN", ""),
		PaymentProvider:   getEnv("PAYMENT_PROVIDER", "yookassa"),
		YookassaShopID:    getEnv("YOOKASSA_SHOP_ID", ""),
		YookassaSecret:    getEnv("YOOKASSA_SECRET", ""),
		SchedulerEnabled:  getEnvBool("SCHEDULER_ENABLED", true),
		Timezone:          getEnv("TIMEZONE", "Europe/Moscow"),
		SubscriptionDays:  getEnvInt("SUBSCRIPTION_DAYS", 30),
		SubscriptionPrice: getEnvInt("SUBSCRIPTION_PRICE_MONTH", 299),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://localhost:5432/bot?sslmode=disable"),

		Mode:          getEnv("MODE", "polling"),
		WebhookURL:    getEnv("WEBHOOK_URL", ""),
		WebhookListen: getEnv("WEBHOOK_LISTEN", ":8443"),
		WebhookSecret: getEnv("WEBHOOK_SECRET", ""),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}
