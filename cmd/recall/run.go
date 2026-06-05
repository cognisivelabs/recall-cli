package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/placeholders"
	"github.com/CognisiveLabs/recall-cli/internal/shell"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/tui"

	"github.com/spf13/cobra"
)

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
		Run: func(cmd *cobra.Command, args []string) {
			query := strings.Join(args, " ")

			match, err := findBestMatch(store, query, tagFilter)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if match == nil {
				fmt.Fprintf(os.Stderr, "No command matching %q found.\n", query)
				os.Exit(1)
			}

			if match.ID != 0 {
				store.RecordUsage(match.ID)
			}

			resolved := match.Pattern

			if placeholders.HasPlaceholders(resolved) {
				var remaining []placeholders.Placeholder
				resolved, remaining = placeholders.AutoResolve(resolved)
				if len(remaining) > 0 {
					rm := tui.NewResolvingModelFromParsed(resolved, remaining)
					p := tui.NewResolverProgram(rm)
					result, err := p.Run()
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error resolving placeholders: %v\n", err)
						os.Exit(1)
					}
					if m, ok := result.(tui.ResolvingModel); ok && m.Done() {
						resolved = m.Resolved()
					} else {
						fmt.Fprintln(os.Stderr, "Cancelled.")
						return
					}
				}
			}

			if dryRun {
				fmt.Println(resolved)
				return
			}

			shell.Execute(resolved)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry", false, "Print the resolved command without executing")
	cmd.Flags().StringVarP(&tagFilter, "tag", "t", "", "Filter by tag before searching")
	return cmd
}

func findBestMatch(store storage.Storage, query string, tagFilter string) (*storage.Command, error) {
	cmds, err := store.List()
	if err != nil {
		return nil, err
	}

	cmds = storage.FilterByTag(cmds, tagFilter)
	query = strings.ToLower(query)

	// Exact pattern match
	for i := range cmds {
		if strings.ToLower(cmds[i].Pattern) == query {
			return &cmds[i], nil
		}
	}

	// Substring match on pattern
	var candidates []storage.Command
	for _, c := range cmds {
		if strings.Contains(strings.ToLower(c.Pattern), query) {
			candidates = append(candidates, c)
		}
	}
	if len(candidates) == 1 {
		return &candidates[0], nil
	}
	if len(candidates) > 1 {
		fmt.Fprintf(os.Stderr, "Multiple matches for %q:\n", query)
		for i, c := range candidates {
			fmt.Fprintf(os.Stderr, "  [%d] %s — %s\n", i+1, c.Pattern, c.Description)
		}
		return &candidates[0], nil
	}

	// Substring match on description/tags
	for i := range cmds {
		searchable := strings.ToLower(cmds[i].Description + " " + cmds[i].Tags)
		if strings.Contains(searchable, query) {
			return &cmds[i], nil
		}
	}

	return nil, nil
}
