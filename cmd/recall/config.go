package main

import (
	"fmt"
	"os"

	"github.com/CognisiveLabs/recall-cli/internal/config"
	"github.com/CognisiveLabs/recall-cli/internal/paths"

	"github.com/spf13/cobra"
)

// NewConfigCmd returns the `recall config` parent command with three sub-commands:
//   - init  — create a default config file at the standard path
//   - path  — print the config file path (useful for scripting)
//   - show  — print the active configuration
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
			cfgPath, _ := paths.ConfigPath()
			fmt.Fprintf(os.Stderr, "Config created at %s\n", cfgPath)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := paths.ConfigPath()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), cfgPath)
			return nil
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

			cfgPath, _ := paths.ConfigPath()
			fmt.Fprintf(cmd.ErrOrStderr(), "Config: %s\n\n", cfgPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Theme: %s\n\n", cfg.Theme)
			fmt.Fprintln(cmd.OutOrStdout(), "Sources:")
			for _, s := range cfg.Sources {
				if s.Git != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s (git: %s)\n", s.Name, s.Git)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s (path: %s)\n", s.Name, s.Path)
				}
			}
			return nil
		},
	})

	return cmd
}
