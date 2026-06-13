// Package search implements fuzzy command matching for the recall CLI.
//
// The only public entry point is FindBest. It is a pure function — it takes
// a slice of commands and returns a ranked result, with no I/O side effects.
package search

import (
	"sort"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

// Result is a single match returned by FindBest.
type Result struct {
	Command storage.Command
	Score   int
}

// FindBest searches cmds for the best match against query.
//
// Priority order (highest to lowest):
//  1. Exact case-insensitive match on Pattern
//  2. Substring match in Pattern  (+100)
//  3. Fuzzy character-order match on Pattern  (×2 multiplier)
//  4. Substring match in Description  (+50)
//  5. Substring match in Tags  (+30)
//  6. Fuzzy match on Description  (+score)
//
// Returns all scored matches in descending order, or an empty slice when
// nothing scores above zero. The caller can inspect len(results) to decide
// whether to show a disambiguation prompt.
func FindBest(cmds []storage.Command, query string) []Result {
	if query == "" || len(cmds) == 0 {
		return nil
	}

	queryLower := strings.ToLower(query)

	// Exact match short-circuits everything else.
	for i := range cmds {
		if strings.ToLower(cmds[i].Pattern) == queryLower {
			return []Result{{Command: cmds[i], Score: 1000}}
		}
	}

	var results []Result
	for _, cmd := range cmds {
		score := scoreCommand(cmd, queryLower)
		if score > 0 {
			results = append(results, Result{Command: cmd, Score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// scoreCommand computes a relevance score for cmd against the lower-cased query.
func scoreCommand(cmd storage.Command, queryLower string) int {
	patternLower := strings.ToLower(cmd.Pattern)
	descLower := strings.ToLower(cmd.Description)
	tagsLower := strings.ToLower(cmd.Tags)

	score := 0

	if strings.Contains(patternLower, queryLower) {
		score += 100
	}
	score += fuzzyScore(queryLower, patternLower) * 2

	if strings.Contains(descLower, queryLower) {
		score += 50
	}
	if strings.Contains(tagsLower, queryLower) {
		score += 30
	}
	score += fuzzyScore(queryLower, descLower)

	return score
}

// fuzzyScore returns a score for how well query characters appear — in order —
// inside target. Returns 0 if not all query characters are found in order.
// Bonuses are awarded for consecutive matches and word-boundary positions
// (start of string, or preceded by space / / / - / _).
func fuzzyScore(query, target string) int {
	if len(query) == 0 {
		return 0
	}

	qi := 0
	score := 0
	prevMatchIdx := -1

	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if target[ti] != query[qi] {
			continue
		}
		score += 10
		if prevMatchIdx == ti-1 {
			score += 5 // consecutive bonus
		}
		if ti == 0 || target[ti-1] == ' ' || target[ti-1] == '/' ||
			target[ti-1] == '-' || target[ti-1] == '_' {
			score += 8 // word-boundary bonus
		}
		prevMatchIdx = ti
		qi++
	}

	if qi < len(query) {
		return 0 // not all query chars matched in order
	}
	return score
}
