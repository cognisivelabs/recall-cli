package main

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"github.com/spf13/cobra"
)

// NewListCmd returns the `recall list` command.
// Supports filtering by tag (-t) or source (-s). Outputs a human-readable table
// by default; use --json for machine-readable output.
func NewListCmd(store storage.Storage) *cobra.Command {
	var tagFilter string
	var sourceFilter string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all saved commands",
		Aliases: []string{"ls"},
		Example: `  recall list
  recall list -t docker
  recall list -s team-ops
  recall list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmds, err := store.List()
			if err != nil {
				return fmt.Errorf("listing commands: %w", err)
			}

			filtered := storage.FilterByTag(cmds, tagFilter)
			if sourceFilter != "" {
				var srcFiltered []storage.Command
				for _, c := range filtered {
					if c.Source == sourceFilter {
						srcFiltered = append(srcFiltered, c)
					}
				}
				filtered = srcFiltered
			}

			if len(filtered) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "No commands found.")
				return nil
			}

			if jsonOutput {
				return printJSON(cmd.OutOrStdout(), filtered)
			}
			printTable(cmd.OutOrStdout(), filtered)
			return nil
		},
	}

	cmd.Flags().StringVarP(&tagFilter, "tag", "t", "", "Filter by tag")
	cmd.Flags().StringVarP(&sourceFilter, "source", "s", "", "Filter by source (e.g. local, team-ops)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

// commandJSON is the DTO used for --json output. Field names are stable across
// refactors because json tags are explicit.
type commandJSON struct {
	ID          int    `json:"id"`
	Pattern     string `json:"pattern"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Source      string `json:"source"`
	UsageCount  int    `json:"usage_count"`
}

// printTable writes commands as an aligned tab-separated table to w.
// Long values are truncated so columns stay readable in a normal terminal width.
func printTable(w io.Writer, cmds []storage.Command) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "ID\tCOMMAND\tDESCRIPTION\tTAGS\tUSED\tSOURCE\n")
	fmt.Fprintf(tw, "──\t───────\t───────────\t────\t────\t──────\n")
	for _, c := range cmds {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%d\t%s\n",
			c.ID,
			truncate(c.Pattern, 50),
			truncate(c.Description, 30),
			truncate(c.Tags, 20),
			c.UsageCount,
			c.Source,
		)
	}
	tw.Flush()
}

// printJSON marshals commands as a JSON array to w using encoding/json.
func printJSON(w io.Writer, cmds []storage.Command) error {
	dtos := make([]commandJSON, len(cmds))
	for i, c := range cmds {
		dtos[i] = commandJSON{
			ID:          c.ID,
			Pattern:     c.Pattern,
			Description: c.Description,
			Tags:        c.Tags,
			Source:      c.Source,
			UsageCount:  c.UsageCount,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(dtos)
}

// truncate shortens s to max runes, appending "..." if it was cut.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}
