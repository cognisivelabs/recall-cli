package main

import (
	"os"
	"path/filepath"

	"github.com/CognisiveLabs/recall-cli/internal/shell"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [shell]",
		Short: "Generate shell integration script (zsh, bash)",
		Long:  `Outputs a shell integration script that wires up Ctrl+Space to launch Recall and wraps the recall command for interactive use. Defaults to detecting your shell from $SHELL.`,
		Example: `  # Add to ~/.zshrc
  eval "$(recall init zsh)"

  # Add to ~/.bashrc
  eval "$(recall init bash)"

  # Auto-detect from $SHELL
  eval "$(recall init)"

  # Also bind Ctrl+R (replaces default history search)
  RECALL_BIND_CTRL_R=1 eval "$(recall init zsh)"`,
		ValidArgs: []string{"zsh", "bash"},
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			sh := ""
			if len(args) > 0 {
				sh = args[0]
			} else {
				sh = filepath.Base(os.Getenv("SHELL"))
			}
			return shell.PrintInitScript(cmd.OutOrStdout(), sh)
		},
	}
}
