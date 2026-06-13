package main

import (
	"fmt"

	"github.com/CognisiveLabs/recall-cli/internal/config"
	"github.com/CognisiveLabs/recall-cli/internal/gitops"
	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"github.com/spf13/cobra"
)

// NewSyncCmd returns the `recall sync` command.
// Reads git sources from the config file and pulls/clones each repo, then
// imports any YAML command files found inside into the local store.
func NewSyncCmd(store storage.Storage) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync commands from git sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if err := gitops.Sync(cfg, store); err != nil {
				return fmt.Errorf("syncing: %w", err)
			}

			fmt.Fprintln(cmd.ErrOrStderr(), "Sync complete.")
			return nil
		},
	}
}
