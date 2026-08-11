package lint

import (
	"fmt"
	"strings"
)

// Per-finding penalty weights for Score. Tuned for a simple MVP curve:
// a prompt with one exact-dupe group lands around 85; stacking several
// issue types drops into the 40–60 band; pathological reports clamp at 0.
const (
	penaltyExactDupe     = 15
	penaltyNearDupe      = 10
	penaltyExactExample  = 15
	penaltyNearExample   = 10
	penaltyStructured    = 8
	penaltyLongHistory   = 20
)

// ScoreResult is a 0–100 prompt quality score plus the human-readable
// deductions that produced it.
type ScoreResult struct {
	Score  int
	Issues []string
}

// Score derives a 0–100 quality score from a lint Report. Starts at 100
// and deducts weighted penalties for exact/near duplicates, redundant
// examples, structured-data opportunities, and a long-history flag.
// The result is clamped to [0, 100].
func Score(r Report) ScoreResult {
	score := 100
	var issues []string

	if n := len(r.Duplicates); n > 0 {
		p := penaltyExactDupe * n
		score -= p
		issues = append(issues, fmt.Sprintf(
			"Exact duplicate instructions (%d group%s): -%d", n, plural(n), p))
	}
	if n := len(r.NearDuplicates); n > 0 {
		p := penaltyNearDupe * n
		score -= p
		issues = append(issues, fmt.Sprintf(
			"Near-duplicate instructions (%d group%s): -%d", n, plural(n), p))
	}
	if n := len(r.DuplicateExamples); n > 0 {
		p := penaltyExactExample * n
		score -= p
		issues = append(issues, fmt.Sprintf(
			"Duplicate examples (%d group%s): -%d", n, plural(n), p))
	}
	if n := len(r.NearDuplicateExamples); n > 0 {
		p := penaltyNearExample * n
		score -= p
		issues = append(issues, fmt.Sprintf(
			"Near-duplicate examples (%d group%s): -%d", n, plural(n), p))
	}
	if n := len(r.StructuredData); n > 0 {
		p := penaltyStructured * n
		score -= p
		issues = append(issues, fmt.Sprintf(
			"Structured data opportunities (%d block%s): -%d", n, plural(n), p))
	}
	if hasCompressHistorySuggestion(r.Suggestions) {
		score -= penaltyLongHistory
		issues = append(issues, fmt.Sprintf(
			"Long conversation history: -%d", penaltyLongHistory))
	}

	if score < 0 {
		score = 0
	}
	return ScoreResult{Score: score, Issues: issues}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func hasCompressHistorySuggestion(suggestions []string) bool {
	for _, s := range suggestions {
		if strings.EqualFold(s, "Compress history") {
			return true
		}
	}
	return false
}
