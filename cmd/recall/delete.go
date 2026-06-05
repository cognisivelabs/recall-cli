package main

import (
	"fmt"
	"os"
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
		Run: func(cmd *cobra.Command, args []string) {
			arg := args[0]

			if id, err := strconv.Atoi(arg); err == nil {
				deleteByID(store, id, force)
				return
			}

			deleteByPattern(store, arg, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	return cmd
}

func deleteByID(store storage.Storage, id int, force bool) {
	target, err := store.GetByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if target == nil {
		fmt.Fprintf(os.Stderr, "No command with ID %d found.\n", id)
		os.Exit(1)
	}

	if !force {
		fmt.Fprintf(os.Stderr, "Delete: %s\n", target.Pattern)
		if !confirm() {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return
		}
	}

	if err := store.Delete(id); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Deleted command #%d\n", id)
}

func deleteByPattern(store storage.Storage, pattern string, force bool) {
	cmds, err := store.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	pattern = strings.ToLower(pattern)
	var matches []storage.Command
	for _, c := range cmds {
		if strings.Contains(strings.ToLower(c.Pattern), pattern) {
			matches = append(matches, c)
		}
	}

	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "No command matching %q found.\n", pattern)
		os.Exit(1)
	}

	if len(matches) > 1 {
		fmt.Fprintf(os.Stderr, "Multiple matches for %q:\n", pattern)
		for _, c := range matches {
			fmt.Fprintf(os.Stderr, "  [%d] %s — %s\n", c.ID, c.Pattern, c.Description)
		}
		fmt.Fprintln(os.Stderr, "Use the ID to delete a specific command.")
		os.Exit(1)
	}

	target := matches[0]
	if !force {
		fmt.Fprintf(os.Stderr, "Delete: %s\n", target.Pattern)
		if !confirm() {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return
		}
	}

	if err := store.Delete(target.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Deleted command #%d\n", target.ID)
}

func confirm() bool {
	fmt.Fprintf(os.Stderr, "Are you sure? [y/N] ")
	var input string
	fmt.Scanln(&input)
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}
