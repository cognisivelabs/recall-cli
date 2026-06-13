package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatches_ExactPrefix(t *testing.T) {
	if !Matches("/home/user/work/billing", "/home/user/work/billing") {
		t.Error("exact match should return true")
	}
}

func TestMatches_Prefix(t *testing.T) {
	if !Matches("/home/user/work/billing/src", "/home/user/work/billing") {
		t.Error("prefix match should return true")
	}
}

func TestMatches_Glob(t *testing.T) {
	if !Matches("/home/user/work/billing-service", "/home/user/work/billing-*") {
		t.Error("glob match should return true")
	}
}

func TestMatches_NoMatch(t *testing.T) {
	if Matches("/home/user/personal/blog", "/home/user/work/billing") {
		t.Error("non-matching path should return false")
	}
}

func TestMatches_EmptyFilter(t *testing.T) {
	if Matches("/home/user/work", "") {
		t.Error("empty filter should return false")
	}
}

func TestMatches_EmptyCwd(t *testing.T) {
	if Matches("", "/home/user/work") {
		t.Error("empty cwd should return false")
	}
}

func TestMatches_HomeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	cwd := filepath.Join(home, "projects", "myapp")
	if !Matches(cwd, "~/projects/myapp") {
		t.Error("~ expansion should match")
	}
}

func TestDetect_NotEmpty(t *testing.T) {
	cwd := Detect()
	if cwd == "" {
		t.Error("Detect should return a non-empty path")
	}
}
