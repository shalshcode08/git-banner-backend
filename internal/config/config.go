package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port string
	Env  string
}

// Load reads configuration from environment variables and applies sane defaults.
func Load() (*Config, error) {
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

	return &Config{
		Port: port,
		Env:  env,
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
