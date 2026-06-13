package main

import (
	"fmt"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"github.com/spf13/cobra"
)

// NewDeleteCmd returns the `recall delete` (alias: rm) command.
// Accepts a numeric ID or an exact pattern string to identify the command.
// Prompts for confirmation unless --force is passed.
func NewDeleteCmd(store storage.Storage) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete <id or pattern>",
		Short:   "Delete a saved command by ID or pattern",
		Aliases: []string{"rm"},
		Example: `  recall delete 3
					recall delete "kubectl logs"
					recall delete 3 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := strings.TrimSpace(args[0])

			target, err := findCommandByIDOrPattern(store, arg)
			if err != nil {
				return err
			}
			return confirmAndDelete(cmd, store, *target, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	return cmd
}

// confirmAndDelete asks for confirmation (unless force), then deletes.
// Shared by both ID-based and pattern-based delete paths.
// Note: --force skips the y/N prompt, but does NOT skip disambiguation.
// If multiple commands match, the caller must resolve ambiguity before calling this.
func confirmAndDelete(cmd *cobra.Command, store storage.Storage, target storage.Command, force bool) error {
	if !force {
		prompt := fmt.Sprintf("Delete: %s\nAre you sure? [y/N] ", target.Pattern)
		if !confirmFromReader(cmd.InOrStdin(), cmd.ErrOrStderr(), prompt) {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
			return nil
		}
	}

	if err := store.Delete(target.ID); err != nil {
		return fmt.Errorf("deleting command: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Deleted command #%d\n", target.ID)
	return nil
}
