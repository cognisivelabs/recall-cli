package main

import (
	"fmt"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/tui"

	"github.com/spf13/cobra"
)

// NewAddCmd returns the `recall add` command.
// With no flags it opens an interactive form. With a pattern + at least one flag
// (-d description or -t tags) it saves non-interactively. If the pattern already
// exists in interactive mode it refuses and suggests `recall edit` instead; in
// non-interactive mode it upserts (update if exists, insert otherwise).
func NewAddCmd(store storage.Storage) *cobra.Command {
	var desc string
	var tags string

	cmd := &cobra.Command{
		Use:   "add [command]",
		Short: "Add a new command to recall",
		Long:  `Add a command directly. Opens a form to fill in details, or use flags for non-interactive mode.`,
		Example: `  recall add "kubectl logs -f deploy/{{service}}"
  recall add "docker compose up -d" -d "Start all services" -t "docker,dev"
  recall add "kubectl get pods" -t prod
  recall add  # opens empty form`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern := ""
			if len(args) > 0 {
				pattern = strings.TrimSpace(args[0])
			}
			desc = strings.TrimSpace(desc)
			tags = strings.TrimSpace(tags)

			// Non-interactive: flags provided alongside a pattern
			nonInteractive := pattern != "" && (desc != "" || tags != "")
			if nonInteractive {
				existing, err := store.GetByPattern(pattern)
				if err != nil {
					return fmt.Errorf("checking for existing command: %w", err)
				}
				if err := store.Upsert(storage.Command{
					Pattern:     pattern,
					Description: desc,
					Tags:        tags,
				}); err != nil {
					return fmt.Errorf("saving command: %w", err)
				}
				verb := "Saved"
				if existing != nil {
					verb = "Updated"
				}
				cmd.Printf("%s: %s\n", verb, pattern)
				return nil
			}

			// Interactive: reject if duplicate exists
			if pattern != "" {
				existing, err := store.GetByPattern(pattern)
				if err != nil {
					return fmt.Errorf("checking for existing command: %w", err)
				}
				if existing != nil {
					return fmt.Errorf("command already exists as #%d: %s\nUse 'recall edit %q' to modify it", existing.ID, existing.Description, pattern)
				}
			}

			// Interactive: open form for new command
			return tui.StartForm(store, &storage.Command{
				Pattern:     pattern,
				Description: desc,
				Tags:        tags,
			})
		},
	}

	cmd.Flags().StringVarP(&desc, "desc", "d", "", "Description of the command")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "Comma-separated tags")
	return cmd
}
