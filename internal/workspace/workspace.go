package workspace

import (
	"os"
	"path/filepath"
	"strings"
)

func Detect() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

func Matches(cwd, filter string) bool {
	if filter == "" || cwd == "" {
		return false
	}

	filter = expandHome(filter)

	if matched, _ := filepath.Match(filter, cwd); matched {
		return true
	}

	return strings.HasPrefix(cwd, filter)
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
