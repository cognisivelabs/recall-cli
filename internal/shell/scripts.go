package shell

import (
	_ "embed"
	"fmt"
)

//go:embed scripts/init.zsh
var zshScript string

//go:embed scripts/init.bash
var bashScript string

// PrintInitScript outputs the shell integration script for the given shell type.
// The scripts are embedded at compile time from the scripts/ directory.
func PrintInitScript(shellType string) {
	switch shellType {
	case "zsh":
		fmt.Print(zshScript)
	case "bash":
		fmt.Print(bashScript)
	default:
		fmt.Printf("# Unsupported shell: %s. Defaulting to Zsh.\n%s", shellType, zshScript)
	}
}
