package gitops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CognisiveLabs/recall-cli/internal/config"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

func Sync(cfg *config.Config, store storage.Storage) error {
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

		fmt.Fprintf(os.Stderr, "Syncing %s...\n", source.Name)
		repoPath := filepath.Join(sourcesDir, source.Name)

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			cmd := exec.Command("git", "clone", source.Git, repoPath)
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to clone %s: %v\n", source.Name, err)
				continue
			}
		} else {
			cmd := exec.Command("git", "-C", repoPath, "pull")
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to pull %s: %v\n", source.Name, err)
				continue
			}
		}

		imported, err := ImportFromRepo(store, repoPath, source.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to import from %s: %v\n", source.Name, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "Imported %d commands from %s\n", imported, source.Name)
	}

	return nil
}
