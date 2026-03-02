package config

import (
	"os"
	"testing"
)

// clearConfigEnv unsets all config-related env vars and restores them after test.
func clearConfigEnv(t *testing.T) {
	t.Helper()
	vars := []string{"PORT", "ENV", "GITHUB_TOKEN", "CACHE_TTL"}
	saved := make(map[string]string, len(vars))
	for _, k := range vars {
		saved[k] = os.Getenv(k)
		os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	})
}

// inTempDir changes to a fresh temp dir for the duration of the test so that
// Load() doesn't pick up a real .env file from the project root.
func inTempDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
}

func TestLoad_Defaults(t *testing.T) {
	inTempDir(t)
	clearConfigEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port: expected %q, got %q", "8080", cfg.Port)
	}
	if cfg.Env != "development" {
		t.Errorf("Env: expected %q, got %q", "development", cfg.Env)
	}
	if cfg.CacheTTL.Seconds() != 300 {
		t.Errorf("CacheTTL: expected 300s, got %v", cfg.CacheTTL)
	}
	if cfg.GitHubToken != "" {
		t.Errorf("GitHubToken: expected empty, got %q", cfg.GitHubToken)
	}
}

func TestLoad_ValidOverrides(t *testing.T) {
	inTempDir(t)
	clearConfigEnv(t)
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "production")
	os.Setenv("GITHUB_TOKEN", "ghp_test")
	os.Setenv("CACHE_TTL", "60")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port: expected %q, got %q", "9090", cfg.Port)
	}
	if cfg.Env != "production" {
		t.Errorf("Env: expected %q, got %q", "production", cfg.Env)
	}
	if cfg.GitHubToken != "ghp_test" {
		t.Errorf("GitHubToken: expected %q, got %q", "ghp_test", cfg.GitHubToken)
	}
	if cfg.CacheTTL.Seconds() != 60 {
		t.Errorf("CacheTTL: expected 60s, got %v", cfg.CacheTTL)
	}
}

func TestLoad_InvalidPORT(t *testing.T) {
	inTempDir(t)
	clearConfigEnv(t)
	os.Setenv("PORT", "not_a_number")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid PORT, got nil")
	}
}

func TestLoad_InvalidCacheTTL(t *testing.T) {
	inTempDir(t)
	clearConfigEnv(t)
	os.Setenv("CACHE_TTL", "bad")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid CACHE_TTL, got nil")
	}
}

func TestLoad_NegativeCacheTTL(t *testing.T) {
	inTempDir(t)
	clearConfigEnv(t)
	os.Setenv("CACHE_TTL", "-5")

	_, err := Load()
	if err == nil {
		t.Error("expected error for negative CACHE_TTL, got nil")
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	cfg := &Config{Env: "development"}
	if !cfg.IsDevelopment() {
		t.Error("expected IsDevelopment() true")
	}
	if cfg.IsProduction() {
		t.Error("expected IsProduction() false")
	}
}

func TestConfig_IsProduction(t *testing.T) {
	cfg := &Config{Env: "production"}
	if !cfg.IsProduction() {
		t.Error("expected IsProduction() true")
	}
	if cfg.IsDevelopment() {
		t.Error("expected IsDevelopment() false")
	}
}
