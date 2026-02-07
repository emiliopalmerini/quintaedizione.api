package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Run("missing DATABASE_URL fails", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("API_KEY", "")
		t.Setenv("APP_VERSION", "dev")

		_, err := Load()

		if err == nil {
			t.Fatal("expected error for missing DATABASE_URL")
		}
	})

	t.Run("non-dev without API_KEY fails", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("API_KEY", "")
		t.Setenv("APP_VERSION", "1.0.0")

		_, err := Load()

		if err == nil {
			t.Fatal("expected error for missing API_KEY in non-dev")
		}
	})

	t.Run("dev without API_KEY succeeds", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("API_KEY", "")
		t.Setenv("APP_VERSION", "dev")

		cfg, err := Load()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Version != "dev" {
			t.Errorf("expected version 'dev', got %q", cfg.Version)
		}
		if cfg.APIKey != "" {
			t.Errorf("expected empty API key, got %q", cfg.APIKey)
		}
	})

	t.Run("defaults are applied", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("APP_VERSION", "dev")
		t.Setenv("API_KEY", "")
		t.Setenv("API_PORT", "")
		t.Setenv("DB_MAX_OPEN_CONNS", "")
		t.Setenv("RATE_LIMIT_RPM", "")
		t.Setenv("RATE_LIMIT_ENABLED", "")

		cfg, err := Load()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Server.Port != "8080" {
			t.Errorf("expected default port 8080, got %q", cfg.Server.Port)
		}
		if cfg.Database.MaxOpenConns != 25 {
			t.Errorf("expected default max open conns 25, got %d", cfg.Database.MaxOpenConns)
		}
		if cfg.RateLimit.RequestsPerMinute != 60 {
			t.Errorf("expected default RPM 60, got %d", cfg.RateLimit.RequestsPerMinute)
		}
		if !cfg.RateLimit.Enabled {
			t.Error("expected rate limit enabled by default")
		}
	})

	t.Run("custom env values override defaults", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://custom/db")
		t.Setenv("APP_VERSION", "dev")
		t.Setenv("API_KEY", "")
		t.Setenv("API_PORT", "9090")
		t.Setenv("DB_MAX_OPEN_CONNS", "50")
		t.Setenv("SERVER_READ_TIMEOUT", "5s")
		t.Setenv("RATE_LIMIT_ENABLED", "false")
		t.Setenv("CORS_ALLOWED_ORIGINS", "http://a.com, http://b.com")

		cfg, err := Load()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Server.Port != "9090" {
			t.Errorf("expected port 9090, got %q", cfg.Server.Port)
		}
		if cfg.Database.MaxOpenConns != 50 {
			t.Errorf("expected max open conns 50, got %d", cfg.Database.MaxOpenConns)
		}
		if cfg.Server.ReadTimeout != 5*time.Second {
			t.Errorf("expected read timeout 5s, got %v", cfg.Server.ReadTimeout)
		}
		if cfg.RateLimit.Enabled {
			t.Error("expected rate limit disabled")
		}
		if len(cfg.CORS.AllowedOrigins) != 2 {
			t.Errorf("expected 2 allowed origins, got %d", len(cfg.CORS.AllowedOrigins))
		}
	})
}

func TestGetEnvHelpers(t *testing.T) {
	t.Run("getIntEnv with invalid value returns default", func(t *testing.T) {
		t.Setenv("TEST_INT", "not-a-number")
		got := getIntEnv("TEST_INT", 42)
		if got != 42 {
			t.Errorf("expected default 42, got %d", got)
		}
	})

	t.Run("getDurationEnv with invalid value returns default", func(t *testing.T) {
		t.Setenv("TEST_DUR", "bad")
		got := getDurationEnv("TEST_DUR", 10*time.Second)
		if got != 10*time.Second {
			t.Errorf("expected default 10s, got %v", got)
		}
	})

	t.Run("getBoolEnv with invalid value returns default", func(t *testing.T) {
		t.Setenv("TEST_BOOL", "maybe")
		got := getBoolEnv("TEST_BOOL", true)
		if !got {
			t.Error("expected default true")
		}
	})

	t.Run("getSliceEnv with empty entries", func(t *testing.T) {
		t.Setenv("TEST_SLICE", "a, , b")
		got := getSliceEnv("TEST_SLICE", nil)
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Errorf("expected [a b], got %v", got)
		}
	})

	t.Run("getSliceEnv with all empty returns default", func(t *testing.T) {
		t.Setenv("TEST_SLICE2", ", ,")
		got := getSliceEnv("TEST_SLICE2", []string{"default"})
		if len(got) != 1 || got[0] != "default" {
			t.Errorf("expected [default], got %v", got)
		}
	})
}
