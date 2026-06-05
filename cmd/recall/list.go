package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"github.com/spf13/cobra"
)

func NewListCmd(store storage.Storage) *cobra.Command {
	var tagFilter string
	var sourceFilter string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved commands",
		Aliases: []string{"ls"},
		Example: `  recall list
  recall list -t docker
  recall list -s team-ops
  recall list --json`,
		Run: func(cmd *cobra.Command, args []string) {
			cmds, err := store.List()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
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
				fmt.Fprintln(os.Stderr, "No commands found.")
				return
			}

			if jsonOutput {
				printJSON(filtered)
				return
			}

			printTable(filtered)
		},
	}

	cmd.Flags().StringVarP(&tagFilter, "tag", "t", "", "Filter by tag")
	cmd.Flags().StringVarP(&sourceFilter, "source", "s", "", "Filter by source (e.g. local, team-ops)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func printTable(cmds []storage.Command) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tCOMMAND\tDESCRIPTION\tTAGS\tUSED\tSOURCE\n")
	fmt.Fprintf(w, "──\t───────\t───────────\t────\t────\t──────\n")
	for _, c := range cmds {
		pattern := c.Pattern
		if len(pattern) > 50 {
			pattern = pattern[:47] + "..."
		}
		desc := c.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		tags := c.Tags
		if len(tags) > 20 {
			tags = tags[:17] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\t%s\n",
			c.ID, pattern, desc, tags, c.UsageCount, c.Source,
		)
	}
	w.Flush()
}

func printJSON(cmds []storage.Command) {
	fmt.Print("[")
	for i, c := range cmds {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Printf("\n  {\"id\":%d,\"pattern\":%q,\"description\":%q,\"tags\":%q,\"source\":%q,\"usage_count\":%d}",
			c.ID, c.Pattern, c.Description, c.Tags, c.Source, c.UsageCount,
		)
	}
	fmt.Println("\n]")
}
