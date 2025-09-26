package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMergesYAMLFiles(t *testing.T) {
	dir := t.TempDir()

	base := filepath.Join(dir, "base.yaml")
	override := filepath.Join(dir, "override.yaml")

	if err := os.WriteFile(base, []byte("app:\n  port: \"8080\"\n  env: \"development\"\n"), 0o600); err != nil {
		t.Fatalf("write base config: %v", err)
	}
	if err := os.WriteFile(override, []byte("app:\n  port: \"9090\"\n"), 0o600); err != nil {
		t.Fatalf("write override config: %v", err)
	}

	v, err := New().WithYAML(base, override).Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if got := v.GetString("app.port"); got != "9090" {
		t.Fatalf("expected app.port=9090, got %q", got)
	}
	if got := v.GetString("app.env"); got != "development" {
		t.Fatalf("expected app.env=development, got %q", got)
	}
}

func TestLoadRespectsDotenvWithoutOverwritingExistingEnv(t *testing.T) {
	dir := t.TempDir()
	dotenv := filepath.Join(dir, "custos.env")

	content := "CUSTOS_APP_PORT=7777\nCUSTOS_DATABASE_HOST=dotenv-host\n"
	if err := os.WriteFile(dotenv, []byte(content), 0o600); err != nil {
		t.Fatalf("write dotenv: %v", err)
	}

	t.Setenv("CUSTOS_DATABASE_HOST", "pre-existing")

	v, err := New().WithDotenv(dotenv).WithEnvPrefix("CUSTOS").Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if got := v.GetString("app.port"); got != "7777" {
		t.Fatalf("expected app.port from dotenv, got %q", got)
	}
	if got := os.Getenv("CUSTOS_DATABASE_HOST"); got != "pre-existing" {
		t.Fatalf("expected pre-existing env preserved, got %q", got)
	}
}

func TestLoadAllowsEnvOverrides(t *testing.T) {
	t.Setenv("CUSTOS_APP_PORT", "8088")

	v, err := New().WithEnvPrefix("CUSTOS").Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if got := v.GetString("app.port"); got != "8088" {
		t.Fatalf("expected env override for app.port, got %q", got)
	}
}

func TestLoadIgnoresMissingFiles(t *testing.T) {
	if _, err := New().WithDotenv("./missing.env").WithYAML("./missing.yaml").Load(); err != nil {
		t.Fatalf("expected no error for missing files, got %v", err)
	}
}
