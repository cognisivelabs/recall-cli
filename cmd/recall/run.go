package main

import (
	"fmt"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/placeholders"
	"github.com/CognisiveLabs/recall-cli/internal/search"
	"github.com/CognisiveLabs/recall-cli/internal/shell"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/tui"

	"github.com/spf13/cobra"
)

// NewRunCmd returns the `recall run` command.
// Searches saved commands by the given query, resolves any {{placeholders}},
// and executes the result. Use --dry to print the resolved command without running it.
// Use -t to restrict the search to commands with a specific tag.
func NewRunCmd(store storage.Storage) *cobra.Command {
	var dryRun bool
	var tagFilter string

	cmd := &cobra.Command{
		Use:   "run <query>",
		Short: "Find and execute a saved command",
		Long:  `Search for a command by pattern or description, resolve placeholders, and execute it.`,
		Example: `  recall run "kubectl logs"
  recall run kl
  recall run docker --dry
  recall run logs -t k8s`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")

			match, err := findBestMatch(store, query, tagFilter, cmd)
			if err != nil {
				return fmt.Errorf("searching commands: %w", err)
			}
			if match == nil {
				return fmt.Errorf("no command matching %q found", query)
			}

			if match.ID != 0 {
				store.RecordUsage(match.ID)
			}

			resolved, cancelled, err := resolvePlaceholders(cmd, match.Pattern)
			if err != nil {
				return err
			}
			if cancelled {
				return nil
			}

			if dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), resolved)
				return nil
			}

			exitCode, err := shell.Execute(resolved)
			if err != nil {
				return fmt.Errorf("executing command: %w", err)
			}
			if exitCode != 0 {
				return exitCodeError(exitCode)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry", false, "Print the resolved command without executing")
	cmd.Flags().StringVarP(&tagFilter, "tag", "t", "", "Filter by tag before searching")
	return cmd
}

// findBestMatch looks up the best command for query, filtered by tagFilter.
// Prints an ambiguity notice to stderr when multiple commands tie at the top score.
func findBestMatch(store storage.Storage, query, tagFilter string, cmd *cobra.Command) (*storage.Command, error) {
	cmds, err := store.List()
	if err != nil {
		return nil, err
	}

	cmds = storage.FilterByTag(cmds, tagFilter)
	results := search.FindBest(cmds, query)
	if len(results) == 0 {
		return nil, nil
	}

	// Notify the user when the top two results tie (ambiguous match).
	if len(results) > 1 && results[0].Score == results[1].Score {
		limit := len(results)
		if limit > 5 {
			limit = 5
		}
		top := make([]storage.Command, limit)
		for i := range top {
			top[i] = results[i].Command
		}
		printMatchList(cmd.ErrOrStderr(), query, top)
	}

	return &results[0].Command, nil
}

// resolvePlaceholders interactively fills any {{placeholder}} tokens in pattern.
// Returns the resolved command string, a cancelled flag (user pressed esc), and
// any unexpected error.
func resolvePlaceholders(cmd *cobra.Command, pattern string) (resolved string, cancelled bool, err error) {
	if !placeholders.HasPlaceholders(pattern) {
		return pattern, false, nil
	}

	var remaining []placeholders.Placeholder
	resolved, remaining = placeholders.AutoResolve(pattern)
	if len(remaining) == 0 {
		return resolved, false, nil
	}

	rm := tui.NewResolvingModelFromParsed(resolved, remaining)
	p := tui.NewResolverProgram(rm)
	result, err := p.Run()
	if err != nil {
		return "", false, fmt.Errorf("resolving placeholders: %w", err)
	}

	if m, ok := result.(tui.ResolvingModel); ok && m.Done() {
		return m.Resolved(), false, nil
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
	return "", true, nil
}
