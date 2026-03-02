package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv_BlankLinesAndComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "# comment\n\nDOTENV_T1=hello\n# another\nDOTENV_T2=world\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Unsetenv("DOTENV_T1")
		os.Unsetenv("DOTENV_T2")
	})

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := os.Getenv("DOTENV_T1"); v != "hello" {
		t.Errorf("DOTENV_T1: expected %q, got %q", "hello", v)
	}
	if v := os.Getenv("DOTENV_T2"); v != "world" {
		t.Errorf("DOTENV_T2: expected %q, got %q", "world", v)
	}
}

func TestLoadDotEnv_Quotes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "DOTENV_DQ=\"double quoted\"\nDOTENV_SQ='single quoted'\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Unsetenv("DOTENV_DQ")
		os.Unsetenv("DOTENV_SQ")
	})

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := os.Getenv("DOTENV_DQ"); v != "double quoted" {
		t.Errorf("DOTENV_DQ: expected %q, got %q", "double quoted", v)
	}
	if v := os.Getenv("DOTENV_SQ"); v != "single quoted" {
		t.Errorf("DOTENV_SQ: expected %q, got %q", "single quoted", v)
	}
}

func TestLoadDotEnv_NoOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("DOTENV_EXISTING=from_file\n"), 0600); err != nil {
		t.Fatal(err)
	}
	os.Setenv("DOTENV_EXISTING", "from_env")
	t.Cleanup(func() { os.Unsetenv("DOTENV_EXISTING") })

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := os.Getenv("DOTENV_EXISTING"); v != "from_env" {
		t.Errorf("DOTENV_EXISTING: expected %q (no overwrite), got %q", "from_env", v)
	}
}

func TestLoadDotEnv_MissingFile(t *testing.T) {
	// silently returns nil when file does not exist
	if err := loadDotEnv("/tmp/does-not-exist-ever.env"); err != nil {
		t.Errorf("expected nil for missing file, got %v", err)
	}
}

func TestLoadDotEnv_LineWithoutEquals(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	// line without '=' should be skipped, not error
	content := "NOT_A_KV_PAIR\nDOTENV_VALID=ok\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Unsetenv("DOTENV_VALID") })

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := os.Getenv("DOTENV_VALID"); v != "ok" {
		t.Errorf("DOTENV_VALID: expected %q, got %q", "ok", v)
	}
}
