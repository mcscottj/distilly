// Package dedupe finds exact and near-duplicate lines/instructions
// within a prompt, such as "Always respond in JSON format." repeated
// verbatim, or repeated with only cosmetic differences (casing,
// punctuation, a stray extra word).
package dedupe

import (
	"regexp"
	"strings"
)

// Duplicate represents a group of lines judged to express the same
// instruction.
type Duplicate struct {
	Lines []string
	// Keep is the suggested single line to retain.
	Keep string
	// Confidence is 1.0 for exact matches. Near-duplicates carry a score
	// in (0,1] reflecting how similar the lines are, so callers can gate
	// auto-apply behavior — see Milestone 3's confidence-tier design in
	// docs/roadmap.md ("high confidence -> auto-applied, low confidence
	// -> requires user approval").
	Confidence float64
}

// FindExact returns groups of lines that are identical after trimming
// whitespace and case. This is the safe, always-auto-apply tier.
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
			dupes = append(dupes, Duplicate{Lines: group, Keep: group[0], Confidence: 1.0})
		}
	}
	return dupes
}

// DefaultNearThreshold is a reasonable starting point for FindNear: high
// enough to avoid grouping merely topically-related lines (see
// TestRunNoFalsePositivesOnSimilarButDistinctLines), low enough to catch
// cosmetic rewordings.
const DefaultNearThreshold = 0.7

// DefaultExampleNearThreshold is FindNear's threshold when clustering
// whole few-shot example blocks instead of single lines. It's higher
// than DefaultNearThreshold: two genuinely different examples still
// share a lot of boilerplate structure (the "Q: ... / A: ..." shape,
// short shared words), so on longer, multi-line text the same absolute
// edit distance lands at a much higher similarity ratio than it would
// for a single instruction line. A threshold tuned for lines would
// cluster distinct examples together; see internal/lint.SplitExamples
// and the redundant_examples.txt fixture it's tested against.
const DefaultExampleNearThreshold = 0.85

var (
	punctRe = regexp.MustCompile(`[^\w\s]`)
	spaceRe = regexp.MustCompile(`\s+`)
)

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = punctRe.ReplaceAllString(s, "")
	s = spaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// FindNear groups lines that are similar but not byte/case-identical,
// using normalized Levenshtein similarity. threshold is the minimum
// similarity ratio (0..1) required to group two lines together.
//
// This is still a fully deterministic, non-AI heuristic — it catches
// cosmetic near-identical repeats (punctuation, minor rewording, a
// dropped/added word). True semantic paraphrase detection (e.g. "Use
// markdown." / "Respond in markdown.") requires embeddings and is
// deferred to Milestone 3; at DefaultNearThreshold, lines like those
// score too low on lexical similarity to be grouped here.
func FindNear(lines []string, threshold float64) []Duplicate {
	type entry struct {
		raw  string
		norm string
	}

	var entries []entry
	seenExact := map[string]bool{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// FindExact already owns lines that are identical (case-insensitive);
		// only keep one representative here so near-dup groups don't
		// re-report exact matches under a different name.
		key := strings.ToLower(trimmed)
		if seenExact[key] {
			continue
		}
		seenExact[key] = true

		norm := normalize(trimmed)
		if norm == "" {
			continue
		}
		entries = append(entries, entry{raw: trimmed, norm: norm})
	}

	parent := make([]int, len(entries))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}
		return parent[i]
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[rb] = ra
		}
	}

	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if similarity(entries[i].norm, entries[j].norm) >= threshold {
				union(i, j)
			}
		}
	}

	groups := map[int][]string{}
	var order []int
	for i, e := range entries {
		root := find(i)
		if _, ok := groups[root]; !ok {
			order = append(order, root)
		}
		groups[root] = append(groups[root], e.raw)
	}

	var dupes []Duplicate
	for _, root := range order {
		group := groups[root]
		if len(group) < 2 {
			continue
		}
		dupes = append(dupes, Duplicate{
			Lines:      group,
			Keep:       longest(group),
			Confidence: groupConfidence(group),
		})
	}
	return dupes
}

func longest(lines []string) string {
	best := lines[0]
	for _, l := range lines[1:] {
		if len(l) > len(best) {
			best = l
		}
	}
	return best
}

// groupConfidence reports the weakest pairwise similarity within the
// group, so a group's confidence reflects its least-similar member pair
// rather than an average that could mask a borderline outlier.
func groupConfidence(lines []string) float64 {
	min := 1.0
	for i := 0; i < len(lines); i++ {
		for j := i + 1; j < len(lines); j++ {
			s := similarity(normalize(lines[i]), normalize(lines[j]))
			if s < min {
				min = s
			}
		}
	}
	return min
}

// similarity returns a Levenshtein-based similarity ratio in [0,1]: 1.0
// means identical, 0.0 means completely different.
func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(levenshtein(a, b))/float64(maxLen)
}

// levenshtein computes the edit distance between a and b (byte-wise;
// fine for the ASCII-heavy instruction text this operates on).
func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}
