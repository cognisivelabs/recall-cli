package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"github.com/spf13/cobra"
)

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

			if id, err := strconv.Atoi(arg); err == nil {
				target, err := store.GetByID(id)
				if err != nil {
					return fmt.Errorf("looking up command: %w", err)
				}
				if target == nil {
					return fmt.Errorf("no command with ID %d found", id)
				}
				return confirmAndDelete(cmd, store, *target, force)
			}

			return deleteByPattern(cmd, store, arg, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	return cmd
}

func deleteByPattern(cmd *cobra.Command, store storage.Storage, pattern string, force bool) error {
	// Try exact match first — safest for a destructive op
	exact, err := store.GetByPattern(pattern)
	if err != nil {
		return fmt.Errorf("looking up command: %w", err)
	}
	if exact != nil {
		return confirmAndDelete(cmd, store, *exact, force)
	}

	// No exact match — fall back to substring search and show suggestions
	cmds, err := store.List()
	if err != nil {
		return fmt.Errorf("listing commands: %w", err)
	}

	patternLower := strings.ToLower(pattern)
	var matches []storage.Command
	for _, c := range cmds {
		if strings.Contains(strings.ToLower(c.Pattern), patternLower) {
			matches = append(matches, c)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no command matching %q found", pattern)
	}

	if len(matches) == 1 {
		return confirmAndDelete(cmd, store, matches[0], force)
	}

	// Multiple matches — show them but don't delete (ambiguous + destructive = bad)
	printMatchList(cmd.ErrOrStderr(), pattern, matches)
	return fmt.Errorf("multiple matches for %q — re-run with one of the IDs shown above", pattern)
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
