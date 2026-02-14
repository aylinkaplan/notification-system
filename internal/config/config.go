package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RabbitMQURL     string // e.g. amqp://guest:guest@localhost:5672/
	WebhookBaseURL  string
	WebhookUUID     string
	WorkerCount     int
	RateLimitPerSec int
}

func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/notifications?sslmode=disable")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	webhookBase := getEnv("WEBHOOK_BASE_URL", "https://webhook.site")
	webhookUUID := getEnv("WEBHOOK_UUID", "")
	workerCount := getEnvInt("WORKER_COUNT", 5)
	rateLimit := getEnvInt("RATE_LIMIT_PER_SEC", 100)

	return &Config{
		Port:            port,
		DatabaseURL:     dbURL,
		RabbitMQURL:     rabbitURL,
		WebhookBaseURL:  webhookBase,
		WebhookUUID:     webhookUUID,
		WorkerCount:     workerCount,
		RateLimitPerSec: rateLimit,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
