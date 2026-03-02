package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port        string
	Env         string
	GitHubToken string
	CacheTTL    time.Duration
}

// Load reads configuration from environment variables and applies sane defaults.
// It also loads a .env file from the current working directory if one exists.
func Load() (*Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return nil, fmt.Errorf("loading .env: %w", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("invalid PORT %q: must be a number", port)
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	githubToken := os.Getenv("GITHUB_TOKEN")

	cacheTTLSec := 300
	if v := os.Getenv("CACHE_TTL"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("invalid CACHE_TTL %q: must be a positive integer (seconds)", v)
		}
		cacheTTLSec = n
	}

	return &Config{
		Port:        port,
		Env:         env,
		GitHubToken: githubToken,
		CacheTTL:    time.Duration(cacheTTLSec) * time.Second,
	}, nil
}

// IsDevelopment reports whether the app is running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction reports whether the app is running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
