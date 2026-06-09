package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/placeholders"
	"github.com/CognisiveLabs/recall-cli/internal/shell"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/tui"

	"github.com/spf13/cobra"
)

func NewRunCmd(store storage.Storage) *cobra.Command {
	var dryRun bool
	var tagFilter string

	cmd := &cobra.Command{
		Use:   "run <query>",
		Short: "Find and execute a saved command",
		Long:  `Search for a command by pattern or description, resolve placeholders, and execute it.`,
		Example: `  recall run "kubectl logs"
  recall run kl
  recall run docker --dry
  recall run logs -t k8s`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")

			match, err := findBestMatch(store, query, tagFilter)
			if err != nil {
				return fmt.Errorf("searching commands: %w", err)
			}
			if match == nil {
				return fmt.Errorf("no command matching %q found", query)
			}

			if match.ID != 0 {
				store.RecordUsage(match.ID)
			}

			resolved := match.Pattern

			if placeholders.HasPlaceholders(resolved) {
				var remaining []placeholders.Placeholder
				resolved, remaining = placeholders.AutoResolve(resolved)
				if len(remaining) > 0 {
					rm := tui.NewResolvingModelFromParsed(resolved, remaining)
					p := tui.NewResolverProgram(rm)
					result, err := p.Run()
					if err != nil {
						return fmt.Errorf("resolving placeholders: %w", err)
					}
					if m, ok := result.(tui.ResolvingModel); ok && m.Done() {
						resolved = m.Resolved()
					} else {
						fmt.Fprintln(os.Stderr, "Cancelled.")
						return nil
					}
				}
			}

			if dryRun {
				fmt.Println(resolved)
				return nil
			}

			exitCode, err := shell.Execute(resolved)
			if err != nil {
				return fmt.Errorf("executing command: %w", err)
			}
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry", false, "Print the resolved command without executing")
	cmd.Flags().StringVarP(&tagFilter, "tag", "t", "", "Filter by tag before searching")
	return cmd
}

// findBestMatch searches for a command using fuzzy matching.
// Priority: exact match > substring in pattern > fuzzy score across pattern+description+tags.
// Returns the highest-scoring match, or nil if nothing scores above threshold.
func findBestMatch(store storage.Storage, query string, tagFilter string) (*storage.Command, error) {
	cmds, err := store.List()
	if err != nil {
		return nil, err
	}

	cmds = storage.FilterByTag(cmds, tagFilter)
	queryLower := strings.ToLower(query)

	// Exact pattern match (highest priority)
	for i := range cmds {
		if strings.ToLower(cmds[i].Pattern) == queryLower {
			return &cmds[i], nil
		}
	}

	// Score all commands with fuzzy matching
	type scored struct {
		cmd   storage.Command
		score int
	}
	var results []scored

	for _, cmd := range cmds {
		patternLower := strings.ToLower(cmd.Pattern)
		descLower := strings.ToLower(cmd.Description)
		tagsLower := strings.ToLower(cmd.Tags)

		score := 0

		// Substring match on pattern (high value)
		if strings.Contains(patternLower, queryLower) {
			score += 100
		}

		// Fuzzy character match on pattern
		score += fuzzyScore(queryLower, patternLower) * 2

		// Substring match on description/tags (lower value)
		if strings.Contains(descLower, queryLower) {
			score += 50
		}
		if strings.Contains(tagsLower, queryLower) {
			score += 30
		}

		// Fuzzy match on description
		score += fuzzyScore(queryLower, descLower)

		if score > 0 {
			results = append(results, scored{cmd: cmd, score: score})
		}
	}

	if len(results) == 0 {
		return nil, nil
	}

	// Sort by score descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > 1 && results[0].score == results[1].score {
		limit := len(results)
		if limit > 5 {
			limit = 5
		}
		top := make([]storage.Command, limit)
		for i := 0; i < limit; i++ {
			top[i] = results[i].cmd
		}
		printMatchList(os.Stderr, query, top)
	}

	return &results[0].cmd, nil
}

// fuzzyScore returns a score for how well the query characters appear in order within the target.
// Higher score = better match. Returns 0 if not all query chars are found in order.
func fuzzyScore(query, target string) int {
	if len(query) == 0 {
		return 0
	}

	qi := 0
	score := 0
	prevMatchIdx := -1

	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if target[ti] == query[qi] {
			score += 10
			// Bonus for consecutive matches
			if prevMatchIdx == ti-1 {
				score += 5
			}
			// Bonus for matching at word boundaries (after space, /, -, _)
			if ti == 0 || target[ti-1] == ' ' || target[ti-1] == '/' || target[ti-1] == '-' || target[ti-1] == '_' {
				score += 8
			}
			prevMatchIdx = ti
			qi++
		}
	}

	// All query characters must be found in order
	if qi < len(query) {
		return 0
	}

	return score
}
