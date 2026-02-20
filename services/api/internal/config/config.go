package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env             string
	HTTPAddr        string
	DBDSN           string
	AdminEmail      string
	AdminPassword   string
	OpenAIAPIKey    string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
	OpenAIModel     string
	OpenAITimeout   time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Env:             getEnv("ENV", "dev"),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		DBDSN:           getEnv("DB_DSN", ""),
		AdminEmail:      getEnv("ADMIN_EMAIL", ""),
		AdminPassword:   getEnv("ADMIN_PASSWORD", ""),
		OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),
		ReadTimeout:     getEnvDuration("READ_TIMEOUT_SECONDS", 10) * time.Second,
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT_SECONDS", 15) * time.Second,
		IdleTimeout:     getEnvDuration("IDLE_TIMEOUT_SECONDS", 60) * time.Second,
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT_SECONDS", 10) * time.Second,
	}

	origins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	cfg.AllowedOrigins = splitCSV(origins)

	if cfg.DBDSN == "" {
		return Config{}, fmt.Errorf("DB_DSN is required")
	}

	cfg.OpenAIModel = getEnv("OPENAI_MODEL", "gpt-4o-mini")
	cfg.OpenAITimeout = getEnvDuration("OPENAI_TIMEOUT_SECONDS", 20) * time.Second

	return cfg, nil
}

func splitCSV(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallbackSeconds int) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return time.Duration(n)
		}
	}
	return time.Duration(fallbackSeconds)
}
