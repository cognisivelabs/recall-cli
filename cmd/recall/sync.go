package main

import (
	"fmt"
	"os"

	"github.com/CognisiveLabs/recall-cli/internal/config"
	"github.com/CognisiveLabs/recall-cli/internal/gitops"
	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"github.com/spf13/cobra"
)

func NewSyncCmd(store storage.Storage) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync commands from git sources",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}

			if err := gitops.Sync(cfg, store); err != nil {
				fmt.Fprintf(os.Stderr, "Error syncing: %v\n", err)
				os.Exit(1)
			}

			fmt.Fprintln(os.Stderr, "Sync complete.")
		},
	}
}
