package gitops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CognisiveLabs/recall-cli/internal/config"
)

func Sync(cfg *config.Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sourcesDir := filepath.Join(home, ".local", "share", "recall", "sources")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		return err
	}

	for _, source := range cfg.Sources {
		if source.Git == "" {
			continue
		}

		fmt.Printf("Syncing %s...\n", source.Name)
		repoPath := filepath.Join(sourcesDir, source.Name)

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			// Clone
			cmd := exec.Command("git", "clone", source.Git, repoPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Failed to clone %s: %v\n", source.Name, err)
			}
		} else {
			// Pull
			cmd := exec.Command("git", "pull")
			cmd.Dir = repoPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Failed to pull %s: %v\n", source.Name, err)
			}
		}

		// TODO: After sync, scan the repo for commands and import them into SQLite
		// For MVP, we just ensure the repo is synced.
	}

	return nil
}
