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
	KioskGroup  string
	StaticDir   string
	TemplateDir string
	Organization string
	AppName      string
}

// Load reads configuration from environment variables and returns a Config.
// Returns an error if required variables are missing.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: os.Getenv("DECKEL_DATABASE_URL"),
		ListenAddr:  envOrDefault("DECKEL_LISTEN_ADDR", ":8080"),
		DevMode:     strings.EqualFold(os.Getenv("DECKEL_DEV_MODE"), "true"),
		AdminGroup:  envOrDefault("DECKEL_ADMIN_GROUP", "admin"),
		KioskGroup:  envOrDefault("DECKEL_KIOSK_GROUP", "kiosk"),
		StaticDir:   envOrDefault("DECKEL_STATIC_DIR", "./static"),
		TemplateDir: envOrDefault("DECKEL_TEMPLATE_DIR", "./templates"),
		Organization: envOrDefault("DECKEL_ORGANIZATION", "K4-Bar"),
		AppName:      envOrDefault("DECKEL_APP_NAME", "Deckel"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DECKEL_DATABASE_URL environment variable is required")
	}

	return cfg, nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
