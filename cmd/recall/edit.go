package main

import (
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/tui"

	"github.com/spf13/cobra"
)

func NewEditCmd(store storage.Storage) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <id or pattern>",
		Short: "Edit an existing saved command",
		Long:  `Opens a form pre-filled with the existing command's data for editing.`,
		Example: `  recall edit 3
					recall edit "kubectl logs"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := strings.TrimSpace(args[0])

			target, err := findCommandByIDOrPattern(store, arg)
			if err != nil {
				return err
			}

			return tui.StartForm(store, target)
		},
	}
}
