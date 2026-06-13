// Package config loads and writes the recall YAML configuration file.
package config

import (
	"fmt"
	"os"

	"github.com/CognisiveLabs/recall-cli/internal/paths"
	"gopkg.in/yaml.v3"
)

// Source describes one place recall can load commands from.
// Either Path (local SQLite file) or Git (remote repo URL) must be set.
type Source struct {
	Name string `yaml:"name"`
	Path string `yaml:"path,omitempty"`
	Git  string `yaml:"git,omitempty"`
}

// Config is the top-level structure of ~/.config/recall/config.yaml.
type Config struct {
	Sources []Source `yaml:"sources"`
	Theme   string   `yaml:"theme"`
}

// ConfigDir returns the directory that holds the config file.
// Delegates to paths.ConfigDir so XDG_CONFIG_HOME is respected.
func ConfigDir() (string, error) {
	return paths.ConfigDir()
}

// ConfigPath returns the full path to the YAML config file.
func ConfigPath() (string, error) {
	return paths.ConfigPath()
}

// LoadConfig reads the config file and returns the parsed Config.
// If the file does not exist, a default config is returned — no file is created.
// Callers that need the directory to exist (e.g. config init) must call
// os.MkdirAll themselves before writing.
func LoadConfig() (*Config, error) {
	dbPath, err := paths.DBPath()
	if err != nil {
		return nil, err
	}
	cfg := defaultConfig(dbPath)

	cfgPath, err := paths.ConfigPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	return cfg, nil
}

// defaultConfig builds a minimal Config pointing at the resolved DB path.
func defaultConfig(dbPath string) *Config {
	return &Config{
		Sources: []Source{
			{Name: "personal", Path: dbPath},
		},
		Theme: "default",
	}
}

// WriteDefault creates the config directory (if needed) and writes a commented
// starter config file. Returns an error if the file already exists.
func WriteDefault() error {
	cfgDir, err := paths.ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	cfgPath, err := paths.ConfigPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(cfgPath); err == nil {
		return fmt.Errorf("config already exists at %s", cfgPath)
	}

	dbPath, err := paths.DBPath()
	if err != nil {
		return err
	}
	cfg := defaultConfig(dbPath)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	header := []byte("# Recall CLI configuration\n# https://github.com/CognisiveLabs/recall-cli\n#\n# Add git sources to sync team commands:\n#\n# sources:\n#   - name: personal\n#     path: ~/.local/share/recall/recall.db\n#   - name: team-ops\n#     git: git@github.com:my-org/ops-runbooks.git\n#\n# theme: default\n\n")
	if err := os.WriteFile(cfgPath, append(header, data...), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
