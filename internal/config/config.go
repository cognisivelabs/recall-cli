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

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "recall")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig(home)

	f, err := os.Open(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ConfigDir(), 0755)
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

func DefaultConfig(home string) *Config {
	return &Config{
		Sources: []Source{
			{Name: "personal", Path: filepath.Join(home, ".local", "share", "recall", "recall.db")},
		},
		Theme: "default",
	}
}

func WriteDefault() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	path := ConfigPath()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists at %s", path)
	}

	cfg := DefaultConfig(home)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	header := []byte("# Recall CLI configuration\n# https://github.com/CognisiveLabs/recall-cli\n#\n# Add git sources to sync team commands:\n#\n# sources:\n#   - name: personal\n#     path: ~/.local/share/recall/recall.db\n#   - name: team-ops\n#     git: git@github.com:my-org/ops-runbooks.git\n#\n# theme: default\n\n")

	if err := os.WriteFile(path, append(header, data...), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
