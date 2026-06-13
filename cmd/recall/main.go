package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/CognisiveLabs/recall-cli/internal/shell"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/tui"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// exitCodeError is a sentinel error that carries a non-zero process exit code.
// Sub-commands return this instead of calling os.Exit directly, keeping RunE
// functions testable. main() is the only place that converts it to os.Exit.
type exitCodeError int

func (e exitCodeError) Error() string { return fmt.Sprintf("exit status %d", int(e)) }

// main initializes the SQLite store, wires up all sub-commands, and hands off to Cobra.
// If the user runs `recall` with no sub-command the TUI launches; if the result is a
// command string it is executed directly (or printed to stdout when piped).
func main() {
	store, err := storage.NewSQLiteStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	rootCmd := &cobra.Command{
		Use:          "recall",
		Short:        "Recall: Your external memory for the terminal",
		Long:         `Recall is a command manager that replaces history search with a context-aware, team-syncable dashboard.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			selected, err := tui.Start(store)
			if err != nil {
				return fmt.Errorf("running TUI: %w", err)
			}
			if selected == "" {
				return nil
			}

			// If stdout is a terminal the user is running recall directly — execute.
			// If stdout is piped (shell widget) just print so the wrapper can eval it.
			if term.IsTerminal(int(os.Stdout.Fd())) {
				exitCode, err := shell.Execute(selected)
				if err != nil {
					return fmt.Errorf("execution error: %w", err)
				}
				if exitCode != 0 {
					return exitCodeError(exitCode)
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), selected)
			}
			return nil
		},
	}

	rootCmd.AddCommand(NewSaveCmd(store))
	rootCmd.AddCommand(NewAddCmd(store))
	rootCmd.AddCommand(NewEditCmd(store))
	rootCmd.AddCommand(NewRunCmd(store))
	rootCmd.AddCommand(NewListCmd(store))
	rootCmd.AddCommand(NewDeleteCmd(store))
	rootCmd.AddCommand(NewSyncCmd(store))
	rootCmd.AddCommand(NewConfigCmd())
	rootCmd.AddCommand(NewInitCmd())

	if err := rootCmd.Execute(); err != nil {
		var ec exitCodeError
		if errors.As(err, &ec) {
			os.Exit(int(ec))
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
