package main

import (
	"fmt"
	"os"

	"github.com/CognisiveLabs/recall-cli/internal/shell"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/tui"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

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
		Run: func(cmd *cobra.Command, args []string) {
			selected, err := tui.Start(store)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
				os.Exit(1)
			}
			if selected == "" {
				return
			}

			// If stdout is a terminal, user is running directly — execute the command.
			// If stdout is piped (shell widget), just print so the wrapper can eval it.
			if term.IsTerminal(int(os.Stdout.Fd())) {
				exitCode, err := shell.Execute(selected)
				if err != nil {
					fmt.Fprintf(os.Stderr, "execution error: %v\n", err)
					os.Exit(1)
				}
				if exitCode != 0 {
					os.Exit(exitCode)
				}
			} else {
				fmt.Println(selected)
			}
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
