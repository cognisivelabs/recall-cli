package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/tui"

	"github.com/spf13/cobra"
)

func NewAddCmd(store storage.Storage) *cobra.Command {
	var desc string
	var tags string

	cmd := &cobra.Command{
		Use:   "add [command]",
		Short: "Add a new command to recall",
		Long:  `Add a command directly. Opens a form to fill in details, or use flags for non-interactive mode.`,
		Example: `  recall add "kubectl logs -f deploy/{{service}}"
  recall add "docker compose up -d" -d "Start all services" -t "docker,dev"
  recall add  # opens empty form`,
		Run: func(cmd *cobra.Command, args []string) {
			var pattern string
			if len(args) > 0 {
				pattern = strings.TrimSpace(args[0])
			}

			// Non-interactive mode with flags
			if pattern != "" && desc != "" {
				existing, _ := store.GetByPattern(pattern)
				if existing != nil {
					fmt.Fprintf(os.Stderr, "Already saved as #%d: %s\n", existing.ID, existing.Description)
					fmt.Fprintf(os.Stderr, "Updating with new values.\n")
					existing.Description = desc
					if tags != "" {
						existing.Tags = tags
					}
					if err := store.Update(*existing); err != nil {
						fmt.Fprintf(os.Stderr, "Error updating: %v\n", err)
						os.Exit(1)
					}
					fmt.Fprintf(os.Stderr, "Updated: %s\n", pattern)
					return
				}

				c := storage.Command{
					Pattern:     pattern,
					Description: desc,
					Tags:        tags,
				}
				if err := store.Upsert(c); err != nil {
					fmt.Fprintf(os.Stderr, "Error saving command: %v\n", err)
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "Saved: %s\n", pattern)
				return
			}

			// Interactive mode — check for dupe before opening form
			if pattern != "" {
				existing, _ := store.GetByPattern(pattern)
				if existing != nil {
					fmt.Fprintf(os.Stderr, "Already saved as #%d: %s\n", existing.ID, existing.Description)
					fmt.Fprintf(os.Stderr, "Edit it? [y/N] ")
					var input string
					fmt.Scanln(&input)
					input = strings.ToLower(strings.TrimSpace(input))
					if input != "y" && input != "yes" {
						fmt.Fprintln(os.Stderr, "Skipped.")
						return
					}
					if err := tui.StartForm(store, existing); err != nil {
						fmt.Fprintf(os.Stderr, "Error running form: %v\n", err)
						os.Exit(1)
					}
					return
				}
			}

			cmdObj := &storage.Command{Pattern: pattern, Description: desc, Tags: tags}
			if err := tui.StartForm(store, cmdObj); err != nil {
				fmt.Fprintf(os.Stderr, "Error running form: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&desc, "desc", "d", "", "Description of the command")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "Comma-separated tags")
	return cmd
}
