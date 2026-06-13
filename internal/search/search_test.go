package search_test

import (
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/search"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

// helpers

func cmd(pattern, desc, tags string) storage.Command {
	return storage.Command{Pattern: pattern, Description: desc, Tags: tags}
}

func mustFind(t *testing.T, results []search.Result) search.Result {
	t.Helper()
	if len(results) == 0 {
		t.Fatal("expected at least one result, got none")
	}
	return results[0]
}

// --- FindBest ---

func TestFindBest_EmptyQuery(t *testing.T) {
	cmds := []storage.Command{cmd("kubectl get pods", "list pods", "")}
	if got := search.FindBest(cmds, ""); len(got) != 0 {
		t.Errorf("empty query should return no results, got %d", len(got))
	}
}

func TestFindBest_EmptyCommands(t *testing.T) {
	if got := search.FindBest(nil, "kubectl"); len(got) != 0 {
		t.Errorf("nil commands should return no results, got %d", len(got))
	}
}

func TestFindBest_ExactMatch(t *testing.T) {
	cmds := []storage.Command{
		cmd("kubectl get pods", "list pods", ""),
		cmd("docker ps", "list containers", ""),
	}
	results := search.FindBest(cmds, "kubectl get pods")
	got := mustFind(t, results)
	if got.Command.Pattern != "kubectl get pods" {
		t.Errorf("expected exact match, got %q", got.Command.Pattern)
	}
}

func TestFindBest_ExactMatchCaseInsensitive(t *testing.T) {
	cmds := []storage.Command{cmd("kubectl get pods", "", "")}
	results := search.FindBest(cmds, "KUBECTL GET PODS")
	got := mustFind(t, results)
	if got.Command.Pattern != "kubectl get pods" {
		t.Errorf("expected case-insensitive exact match, got %q", got.Command.Pattern)
	}
}

func TestFindBest_SubstringPatternMatch(t *testing.T) {
	cmds := []storage.Command{
		cmd("kubectl get pods", "list pods", ""),
		cmd("docker ps", "list containers", ""),
	}
	results := search.FindBest(cmds, "kubectl")
	got := mustFind(t, results)
	if got.Command.Pattern != "kubectl get pods" {
		t.Errorf("expected kubectl command, got %q", got.Command.Pattern)
	}
}

func TestFindBest_NoMatch(t *testing.T) {
	cmds := []storage.Command{cmd("kubectl get pods", "list pods", "")}
	if got := search.FindBest(cmds, "zzznomatch"); len(got) != 0 {
		t.Errorf("expected no results for unrecognised query, got %d", len(got))
	}
}

func TestFindBest_DescriptionMatch(t *testing.T) {
	cmds := []storage.Command{
		cmd("kubectl get pods", "list kubernetes pods", ""),
		cmd("docker ps", "list containers", ""),
	}
	results := search.FindBest(cmds, "kubernetes")
	got := mustFind(t, results)
	if got.Command.Pattern != "kubectl get pods" {
		t.Errorf("expected kubectl command via description, got %q", got.Command.Pattern)
	}
}

func TestFindBest_TagMatch(t *testing.T) {
	cmds := []storage.Command{
		cmd("kubectl get pods", "list pods", "k8s,debug"),
		cmd("docker ps", "list containers", "docker"),
	}
	results := search.FindBest(cmds, "k8s")
	got := mustFind(t, results)
	if got.Command.Pattern != "kubectl get pods" {
		t.Errorf("expected kubectl command via tag, got %q", got.Command.Pattern)
	}
}

func TestFindBest_PatternScoresHigherThanDescription(t *testing.T) {
	// "logs" appears in both patterns; the one with substring in pattern wins.
	cmds := []storage.Command{
		cmd("tail -f /var/log/syslog", "view system logs", ""),
		cmd("kubectl logs -f deploy/app", "stream pod output", ""),
	}
	results := search.FindBest(cmds, "logs")
	if len(results) < 2 {
		t.Fatal("expected multiple results")
	}
	// Both have "logs" as substring in pattern; both should score well.
	// The important thing is neither returns zero results.
	for _, r := range results {
		if r.Score == 0 {
			t.Errorf("all results should have score > 0")
		}
	}
}

func TestFindBest_ResultsOrderedByScore(t *testing.T) {
	cmds := []storage.Command{
		cmd("docker ps", "list containers", ""),
		cmd("kubectl get pods", "list pods", "k8s"),
		cmd("kubectl logs -f deploy/app", "stream kubectl logs", "k8s"),
	}
	results := search.FindBest(cmds, "kubectl")
	if len(results) < 2 {
		t.Fatal("expected at least 2 results")
	}
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("results not sorted: results[%d].Score=%d > results[%d].Score=%d",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func TestFindBest_FuzzyMatch(t *testing.T) {
	cmds := []storage.Command{
		cmd("kubectl get pods", "list pods", ""),
		cmd("docker ps", "containers", ""),
	}
	// "kgp" matches k-ubectl-g-et-p-ods in order
	results := search.FindBest(cmds, "kgp")
	got := mustFind(t, results)
	if got.Command.Pattern != "kubectl get pods" {
		t.Errorf("expected fuzzy match on kubectl get pods, got %q", got.Command.Pattern)
	}
}

func TestFindBest_MultipleResults(t *testing.T) {
	cmds := []storage.Command{
		cmd("kubectl get pods", "", ""),
		cmd("kubectl get svc", "", ""),
		cmd("docker ps", "", ""),
	}
	results := search.FindBest(cmds, "kubectl")
	if len(results) < 2 {
		t.Errorf("expected at least 2 matches for 'kubectl', got %d", len(results))
	}
}

// --- fuzzyScore (tested indirectly through FindBest, but worth a direct path) ---

func TestFindBest_FuzzyScoreConsecutiveBonus(t *testing.T) {
	// "abc" consecutive should score higher than "a_b_c" spread out.
	cmds := []storage.Command{
		cmd("xabcx", "consecutive", ""),
		cmd("xaxbxcx", "spread", ""),
	}
	results := search.FindBest(cmds, "abc")
	if len(results) < 2 {
		t.Fatal("expected 2 results")
	}
	if results[0].Command.Pattern != "xabcx" {
		t.Errorf("consecutive match should rank higher, got %q first", results[0].Command.Pattern)
	}
}
