package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration values.
type Config struct {
	DatabaseURL string
	ListenAddr  string
	DevMode     bool
	AdminGroup  string
	StaticDir   string
	TemplateDir string
}

// Load reads configuration from environment variables and returns a Config.
// Returns an error if required variables are missing.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ListenAddr:  envOrDefault("LISTEN_ADDR", ":8080"),
		DevMode:     strings.EqualFold(os.Getenv("DEV_MODE"), "true"),
		AdminGroup:  envOrDefault("ADMIN_GROUP", "admin"),
		StaticDir:   envOrDefault("STATIC_DIR", "./static"),
		TemplateDir: envOrDefault("TEMPLATE_DIR", "./templates"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	return cfg, nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
