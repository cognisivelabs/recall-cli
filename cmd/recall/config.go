package main

import (
	"fmt"
	"os"

	"github.com/CognisiveLabs/recall-cli/internal/config"

	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage recall configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Generate a starter config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.WriteDefault(); err != nil {
				return fmt.Errorf("creating config: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Config created at %s\n", config.ConfigPath())
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.ConfigPath())
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print the current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Config: %s\n\n", config.ConfigPath())
			fmt.Printf("Theme: %s\n\n", cfg.Theme)
			fmt.Println("Sources:")
			for _, s := range cfg.Sources {
				if s.Git != "" {
					fmt.Printf("  - %s (git: %s)\n", s.Name, s.Git)
				} else {
					fmt.Printf("  - %s (path: %s)\n", s.Name, s.Path)
				}
			}
			return nil
		},
	})

	return cmd
}
