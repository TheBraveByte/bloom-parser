package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/TheBraveByte/bloom-parser/internal/powerbi"
)

type Config struct {
	Port         int
	HTTPPort     int
	LogFormat    string
	LogLevel     slog.Level
	OTelEnabled  bool
	MaxBytes     int
	MaxPages     int
	OCRLanguages string
	OCRDefault   bool
	PowerBI      powerbi.Config
}

// Load reads configuration from the environment, with .env as a fallback.
func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:         envInt("SERVER_PORT", 50051),
		HTTPPort:     envInt("HTTP_PORT", 8080),
		LogFormat:    envStr("LOG_FORMAT", "json"),
		LogLevel:     envLevel("LOG_LEVEL", slog.LevelInfo),
		OTelEnabled:  envBool("OTEL_ENABLED", false),
		MaxBytes:     envInt("MAX_DOCUMENT_BYTES", 32<<20),
		MaxPages:     envInt("MAX_DOCUMENT_PAGES", 200),
		OCRLanguages: envStr("OCR_LANGUAGES", "eng"),
		OCRDefault:   envBool("OCR_DEFAULT", false),
		PowerBI: powerbi.Config{
			TenantID:     os.Getenv("POWERBI_TENANT_ID"),
			ClientID:     os.Getenv("POWERBI_CLIENT_ID"),
			ClientSecret: os.Getenv("POWERBI_CLIENT_SECRET"),
			WorkspaceID:  os.Getenv("POWERBI_WORKSPACE_ID"),
			BaseURL:      os.Getenv("POWERBI_BASE_URL"),
			AuthURL:      os.Getenv("POWERBI_AUTH_URL"),
		},
	}
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	if v, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

func envLevel(key string, def slog.Level) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(os.Getenv(key))); err == nil {
		return l
	}
	return def
}
