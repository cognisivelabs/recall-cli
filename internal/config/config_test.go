package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/config"
)

// isolateEnv redirects all path-related env vars to a temp directory so each
// test starts from a clean slate without touching the real user home.
func isolateEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// HOME is used on Unix; USERPROFILE (and HOMEPATH/HOMEDRIVE) is used on Windows.
	// Unset all of them so os.UserHomeDir() reads from the temp HOME we set below.
	for _, key := range []string{"HOME", "USERPROFILE", "HOMEPATH", "HOMEDRIVE", "XDG_DATA_HOME", "XDG_CONFIG_HOME", "RECALL_DB_PATH"} {
		prev, existed := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if existed {
				os.Setenv(key, prev)
			} else {
				os.Unsetenv(key)
			}
		})
	}

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir) // Windows: os.UserHomeDir() reads USERPROFILE first
	t.Cleanup(func() {
		os.Unsetenv("HOME")
		os.Unsetenv("USERPROFILE")
	})

	return tmpDir
}

// TestLoadConfig_DefaultsWhenNoFile verifies that missing config returns defaults, no error.
func TestLoadConfig_DefaultsWhenNoFile(t *testing.T) {
	isolateEnv(t)

	cfg, err := config.LoadConfig()
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

// TestLoadConfig_NoFilesystemSideEffect ensures LoadConfig does NOT create directories.
func TestLoadConfig_NoFilesystemSideEffect(t *testing.T) {
	tmpDir := isolateEnv(t)

	_, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	configDir := filepath.Join(tmpDir, ".config", "recall")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Error("LoadConfig must not create directories — found", configDir)
	}
}

// TestWriteDefault_CreatesFile verifies that WriteDefault creates a valid config file.
func TestWriteDefault_CreatesFile(t *testing.T) {
	tmpDir := isolateEnv(t)

	if err := config.WriteDefault(); err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}

	path := filepath.Join(tmpDir, ".config", "recall", "config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created at", path)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig after WriteDefault: %v", err)
	}
	if cfg.Theme != "default" {
		t.Errorf("expected theme 'default', got %q", cfg.Theme)
	}
}

// TestWriteDefault_ErrorsIfExists verifies that a second WriteDefault fails gracefully.
func TestWriteDefault_ErrorsIfExists(t *testing.T) {
	isolateEnv(t)

	if err := config.WriteDefault(); err != nil {
		t.Fatalf("first WriteDefault: %v", err)
	}
	if err := config.WriteDefault(); err == nil {
		t.Fatal("expected error on second WriteDefault, got nil")
	}
}

// TestLoadConfig_ParsesCustomFile verifies that a real config file is parsed correctly.
func TestLoadConfig_ParsesCustomFile(t *testing.T) {
	tmpDir := isolateEnv(t)

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

	cfg, err := config.LoadConfig()
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

// TestLoadConfig_DefaultDBPathContainsRecall verifies the default source path is sane.
func TestLoadConfig_DefaultDBPathContainsRecall(t *testing.T) {
	isolateEnv(t)

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !strings.Contains(cfg.Sources[0].Path, "recall") {
		t.Errorf("default source path %q does not contain 'recall'", cfg.Sources[0].Path)
	}
}

// TestLoadConfig_MalformedYAML verifies LoadConfig returns an error for invalid YAML.
func TestLoadConfig_MalformedYAML(t *testing.T) {
	tmpDir := isolateEnv(t)

	configDir := filepath.Join(tmpDir, ".config", "recall")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("not: [valid yaml"), 0644)

	_, err := config.LoadConfig()
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

// TestConfigDir_ReturnsNonEmpty verifies ConfigDir returns a usable path.
func TestConfigDir_ReturnsNonEmpty(t *testing.T) {
	isolateEnv(t)

	dir, err := config.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir: %v", err)
	}
	if dir == "" {
		t.Error("ConfigDir returned empty string")
	}
	if !strings.Contains(dir, "recall") {
		t.Errorf("ConfigDir %q does not contain 'recall'", dir)
	}
}

// TestConfigPath_ReturnsYAMLFile verifies ConfigPath ends with config.yaml.
func TestConfigPath_ReturnsYAMLFile(t *testing.T) {
	isolateEnv(t)

	p, err := config.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}
	if !strings.HasSuffix(p, "config.yaml") {
		t.Errorf("ConfigPath %q should end with config.yaml", p)
	}
}

// TestConfigDir_XDGOverride verifies ConfigDir respects XDG_CONFIG_HOME.
func TestConfigDir_XDGOverride(t *testing.T) {
	isolateEnv(t)
	custom := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", custom)
	t.Cleanup(func() { os.Unsetenv("XDG_CONFIG_HOME") })

	dir, err := config.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir: %v", err)
	}
	if !strings.HasPrefix(dir, custom) {
		t.Errorf("expected dir under XDG_CONFIG_HOME %q, got %q", custom, dir)
	}
}

// TestWriteDefault_XDGOverride verifies that WriteDefault respects XDG_CONFIG_HOME.
func TestWriteDefault_XDGOverride(t *testing.T) {
	isolateEnv(t)
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("XDG_CONFIG_HOME") })

	if err := config.WriteDefault(); err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}

	expected := filepath.Join(tmpDir, "recall", "config.yaml")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("expected config at %s, not found", expected)
	}
}
