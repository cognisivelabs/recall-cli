package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_DefaultsWhenNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Theme != "default" {
		t.Errorf("expected theme 'default', got %q", cfg.Theme)
	}
	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Name != "personal" {
		t.Errorf("expected source name 'personal', got %q", cfg.Sources[0].Name)
	}
}

func TestWriteDefault_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	err := WriteDefault()
	if err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}

	path := filepath.Join(tmpDir, ".config", "recall", "config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Loading should work
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig after WriteDefault: %v", err)
	}
	if cfg.Theme != "default" {
		t.Errorf("expected theme 'default', got %q", cfg.Theme)
	}
}

func TestWriteDefault_ErrorsIfExists(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// First write succeeds
	if err := WriteDefault(); err != nil {
		t.Fatalf("first WriteDefault: %v", err)
	}

	// Second write should error
	err := WriteDefault()
	if err == nil {
		t.Fatal("expected error on second WriteDefault, got nil")
	}
}

func TestLoadConfig_ParsesCustomFile(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "recall")
	os.MkdirAll(configDir, 0755)

	content := `
theme: dark
sources:
  - name: personal
    path: /my/db
  - name: team
    git: git@github.com:org/runbooks.git
`
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Theme != "dark" {
		t.Errorf("expected theme 'dark', got %q", cfg.Theme)
	}
	if len(cfg.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(cfg.Sources))
	}
	if cfg.Sources[1].Git != "git@github.com:org/runbooks.git" {
		t.Errorf("expected git source, got %q", cfg.Sources[1].Git)
	}
}
