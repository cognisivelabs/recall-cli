package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Name string `yaml:"name"`
	Path string `yaml:"path,omitempty"`
	Git  string `yaml:"git,omitempty"`
}

type Config struct {
	Sources []Source `yaml:"sources"`
	Theme   string   `yaml:"theme"`
}

func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(home, ".config", "recall")
	configPath := filepath.Join(configDir, "config.yaml")

	// Default config
	cfg := &Config{
		Sources: []Source{
			{Name: "personal", Path: filepath.Join(home, ".local", "share", "recall", "recall.db")},
		},
		Theme: "default",
	}

	f, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Ensure dir exists
			os.MkdirAll(configDir, 0755)
			// Write default? Nah, just return default.
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
