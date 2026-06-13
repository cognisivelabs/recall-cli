package gitops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"gopkg.in/yaml.v3"
)

type CommandEntry struct {
	Pattern     string   `yaml:"pattern"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

type CommandFile struct {
	Commands []CommandEntry `yaml:"commands"`
}

// ImportFromRepo scans repoPath for YAML command files and upserts every valid
// command entry into store, tagging each with sourceName. Returns the count of
// successfully imported commands. Malformed files are logged and skipped.
func ImportFromRepo(store storage.Storage, repoPath string, sourceName string) (int, error) {
	files, err := findCommandFiles(repoPath)
	if err != nil {
		return 0, fmt.Errorf("scanning repo: %w", err)
	}

	var imported int
	for _, f := range files {
		entries, err := parseCommandFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", f, err)
			continue
		}

		for _, entry := range entries {
			if entry.Pattern == "" {
				continue
			}

			c := storage.Command{
				Pattern:     entry.Pattern,
				Description: entry.Description,
				Tags:        strings.Join(entry.Tags, ","),
				Source:      sourceName,
			}

			if err := store.Upsert(c); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to import %q: %v\n", entry.Pattern, err)
				continue
			}
			imported++
		}
	}

	return imported, nil
}

// findCommandFiles walks root and returns paths to all .yaml/.yml files,
// skipping the .git directory.
func findCommandFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		name := info.Name()
		ext := filepath.Ext(name)
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// parseCommandFile reads a YAML command file and returns its entries.
// Accepts two formats:
//   - Structured: { commands: [{pattern, description, tags}, ...] }
//   - Flat list:  [{pattern, description, tags}, ...]
func parseCommandFile(path string) ([]CommandEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try structured format first: { commands: [...] }
	var cf CommandFile
	if err := yaml.Unmarshal(data, &cf); err == nil && len(cf.Commands) > 0 {
		return cf.Commands, nil
	}

	// Fall back to flat list: [{ pattern: ..., description: ... }, ...]
	var entries []CommandEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("unrecognized format: %w", err)
	}

	return entries, nil
}
