package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type config struct {
	databaseURL  string
	port         string
	webURL       string
	cookieSecure bool
	sessionTTL   time.Duration
}

func loadConfig() (config, error) {
	cfg := config{
		databaseURL: os.Getenv("DATABASE_URL"),
		port:        envOrDefault("PORT", "3001"),
		webURL:      envOrDefault("WEB_URL", "http://localhost:3000"),
		sessionTTL:  7 * 24 * time.Hour,
	}

	if cfg.databaseURL == "" {
		return config{}, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	if value := os.Getenv("COOKIE_SECURE"); value != "" {
		secure, err := strconv.ParseBool(value)
		if err != nil {
			return config{}, fmt.Errorf("parse COOKIE_SECURE: %w", err)
		}
		cfg.cookieSecure = secure
	}

	if value := os.Getenv("SESSION_TTL"); value != "" {
		ttl, err := time.ParseDuration(value)
		if err != nil || ttl <= 0 {
			return config{}, fmt.Errorf("SESSION_TTL must be a positive duration")
		}
		cfg.sessionTTL = ttl
	}

	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
