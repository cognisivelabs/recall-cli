package main

import (
	"fmt"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/shell"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/tui"

	"github.com/spf13/cobra"
)

func NewSaveCmd(store storage.Storage) *cobra.Command {
	var lastCmdFlag string

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save the last executed command",
		RunE: func(cmd *cobra.Command, args []string) error {
			var lastCmd string
			var err error

			if lastCmdFlag != "" {
				lastCmd = lastCmdFlag
			} else {
				lastCmd, err = shell.GetLastCommand()
				if err != nil {
					return fmt.Errorf("reading shell history: %w", err)
				}
			}

			lastCmd = strings.TrimSpace(lastCmd)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Last command: %s\n", lastCmd)

			existing, err := store.GetByPattern(lastCmd)
			if err != nil {
				return fmt.Errorf("checking for existing command: %w", err)
			}

			if existing != nil {
				prompt := fmt.Sprintf("Already saved as #%d: %s\nEdit it? [y/N] ", existing.ID, existing.Description)
				if !confirmFromReader(cmd.InOrStdin(), cmd.ErrOrStderr(), prompt) {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Skipped.")
					return nil
				}
				return tui.StartForm(store, existing)
			}

			cmdObj := &storage.Command{Pattern: lastCmd}
			return tui.StartForm(store, cmdObj)
		},
	}
	cmd.Flags().StringVarP(&lastCmdFlag, "last-cmd", "c", "", "Explicitly set the command to save")
	return cmd
}
