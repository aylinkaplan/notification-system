package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear env to get defaults
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("RABBITMQ_URL")
	os.Unsetenv("WEBHOOK_UUID")
	os.Unsetenv("WORKER_COUNT")
	os.Unsetenv("RATE_LIMIT_PER_SEC")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected Port 8080, got %s", cfg.Port)
	}
	if cfg.WorkerCount != 5 {
		t.Errorf("expected WorkerCount 5, got %d", cfg.WorkerCount)
	}
	if cfg.RateLimitPerSec != 100 {
		t.Errorf("expected RateLimitPerSec 100, got %d", cfg.RateLimitPerSec)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("WORKER_COUNT", "10")
	os.Setenv("RATE_LIMIT_PER_SEC", "50")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("WORKER_COUNT")
		os.Unsetenv("RATE_LIMIT_PER_SEC")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected Port 9090, got %s", cfg.Port)
	}
	if cfg.WorkerCount != 10 {
		t.Errorf("expected WorkerCount 10, got %d", cfg.WorkerCount)
	}
	if cfg.RateLimitPerSec != 50 {
		t.Errorf("expected RateLimitPerSec 50, got %d", cfg.RateLimitPerSec)
	}
}
