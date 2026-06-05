package main

import (
	"github.com/CognisiveLabs/recall-cli/internal/shell"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [shell]",
		Short: "Generate shell integration script (zsh, bash)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sh := "zsh"
			if len(args) > 0 {
				sh = args[0]
			}
			shell.PrintInitScript(sh)
		},
	}
}
