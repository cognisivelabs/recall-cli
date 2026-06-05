package main

import (
	"fmt"
	"os"
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
		Run: func(cmd *cobra.Command, args []string) {
			var lastCmd string
			var err error

			if lastCmdFlag != "" {
				lastCmd = lastCmdFlag
			} else {
				lastCmd, err = shell.GetLastCommand()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading history: %v\n", err)
					os.Exit(1)
				}
			}

			lastCmd = strings.TrimSpace(lastCmd)
			fmt.Fprintf(os.Stderr, "Last command: %s\n", lastCmd)

			existing, _ := store.GetByPattern(lastCmd)
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
				// Open form pre-filled with existing data so user can edit
				if err := tui.StartForm(store, existing); err != nil {
					fmt.Fprintf(os.Stderr, "Error running form: %v\n", err)
					os.Exit(1)
				}
				return
			}

			cmdObj := &storage.Command{Pattern: lastCmd}
			if err := tui.StartForm(store, cmdObj); err != nil {
				fmt.Fprintf(os.Stderr, "Error running form: %v\n", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVarP(&lastCmdFlag, "last-cmd", "c", "", "Explicitly set the command to save")
	return cmd
}
