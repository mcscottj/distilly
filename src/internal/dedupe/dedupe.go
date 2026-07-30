// Package dedupe finds exact and near-duplicate lines/instructions
// within a prompt, such as "Use markdown." / "Respond in markdown." /
// "Format as markdown." collapsing to a single instruction.
package dedupe

import "strings"

// Duplicate represents a group of lines judged to express the same
// instruction.
type Duplicate struct {
	Lines []string
	// Keep is the suggested single line to retain.
	Keep string
}

// FindExact returns groups of lines that are byte-for-byte identical
// after trimming whitespace. This is the safe, always-auto-apply tier —
// see Milestone 3 in docs/roadmap.md for confidence-scored semantic
// dedupe on top of this.
func FindExact(lines []string) []Duplicate {
	seen := map[string][]string{}
	order := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; !ok {
			order = append(order, key)
		}
		seen[key] = append(seen[key], trimmed)
	}

	var dupes []Duplicate
	for _, key := range order {
		group := seen[key]
		if len(group) > 1 {
			dupes = append(dupes, Duplicate{Lines: group, Keep: group[0]})
		}
	}
	return dupes
}

// TODO(milestone 3): near-duplicate detection via embedding similarity
// or a local model, with a confidence score attached to each match.
