package lint

import (
	"regexp"
	"strings"

	"distilly/internal/dedupe"
)

// exampleHeaderRe matches a header line that starts a new few-shot
// example block, e.g. "Example 1:", "Example 2:", "Examples:".
var exampleHeaderRe = regexp.MustCompile(`(?i)^\s*examples?\s*\d*\s*:\s*(.*)$`)

// exampleBlock is one few-shot example. Body is what deduplication
// compares — two examples with different numbering ("Example 1:" vs
// "Example 3:") but identical Q/A content are still the same example —
// while Raw (header line included) is what gets reproduced or dropped
// when reconstructing an optimized prompt.
type exampleBlock struct {
	Raw  []string
	Body string
}

// splitExampleBlocks walks prompt and extracts each few-shot example:
// everything from one "Example N:" header up to (but not including) the
// next section or example header.
//
// This is deliberately separate from SplitSections, which flattens all
// examples into one Examples blob with header lines discarded — fine
// for section-level token counts, but it throws away the block
// boundaries example-level deduplication needs.
func splitExampleBlocks(prompt string) []exampleBlock {
	var blocks []exampleBlock
	var raw, body []string
	inExample := false

	flush := func() {
		if inExample {
			if b := strings.TrimSpace(strings.Join(body, "\n")); b != "" {
				blocks = append(blocks, exampleBlock{Raw: raw, Body: b})
			}
		}
		raw, body = nil, nil
	}

	for _, line := range strings.Split(prompt, "\n") {
		if m := exampleHeaderRe.FindStringSubmatch(line); m != nil {
			flush()
			inExample = true
			raw = []string{line}
			if rest := strings.TrimSpace(m[1]); rest != "" {
				body = append(body, rest)
			}
			continue
		}
		if sectionHeaderRe.MatchString(line) {
			flush()
			inExample = false
			continue
		}
		if inExample {
			raw = append(raw, line)
			body = append(body, line)
		}
	}
	flush()

	return blocks
}

// SplitExamples returns the body text of each few-shot example block in
// prompt (header line stripped). Feed this into dedupe.FindExact /
// dedupe.FindNear with dedupe.DefaultExampleNearThreshold to cluster
// redundant examples — see Report.DuplicateExamples/NearDuplicateExamples
// in lint.go.
func SplitExamples(prompt string) []string {
	blocks := splitExampleBlocks(prompt)
	bodies := make([]string, len(blocks))
	for i, b := range blocks {
		bodies[i] = b.Body
	}
	return bodies
}

// nonExampleLines returns every line of prompt that is not part of an
// example block (see splitExampleBlocks), in original order.
//
// Line-level exact-duplicate detection (dedupe.FindExact over the whole
// prompt) is scoped against this set: a duplicate is only treated as a
// leaked/repeated instruction — safe to strip wherever it appears,
// including inside an example — if it also occurs outside every example
// block. A line that repeats only *between* examples (e.g. two
// genuinely different examples that happen to share the same answer)
// is left alone here; that's incidental example structure, and
// collapsing it would corrupt the example it was pulled out of. Whole
// redundant examples are instead caught at the block level — see
// SplitExamples.
func nonExampleLines(prompt string) []string {
	var out []string
	inExample := false
	for _, line := range strings.Split(prompt, "\n") {
		if exampleHeaderRe.MatchString(line) {
			inExample = true
			continue
		}
		if sectionHeaderRe.MatchString(line) {
			inExample = false
		}
		if !inExample {
			out = append(out, line)
		}
	}
	return out
}

// filterToLinesSeenOutsideExamples drops any duplicate group from dupes
// whose lines occur only inside example blocks — see nonExampleLines.
func filterToLinesSeenOutsideExamples(dupes []dedupe.Duplicate, outside []string) []dedupe.Duplicate {
	outsideKeys := make(map[string]bool, len(outside))
	for _, l := range outside {
		outsideKeys[strings.ToLower(strings.TrimSpace(l))] = true
	}

	var kept []dedupe.Duplicate
	for _, d := range dupes {
		if outsideKeys[strings.ToLower(strings.TrimSpace(d.Keep))] {
			kept = append(kept, d)
		}
	}
	return kept
}

// removeExampleBlocks reconstructs prompt with whole example blocks
// (header line included) dropped based on their body text:
//   - a body in removeBody is dropped every time it occurs (used for
//     near-duplicate examples once approved — every occurrence but the
//     kept one should go)
//   - a body in dropExactBody is kept on its first occurrence and
//     dropped on every subsequent one (used for exact-duplicate
//     examples, mirroring Apply's line-level "keep first occurrence"
//     behavior)
//
// Non-example lines (System/History/Question, and text before the first
// example header) pass through untouched.
func removeExampleBlocks(prompt string, dropExactBody, removeBody map[string]bool) string {
	lines := strings.Split(prompt, "\n")
	var out, raw, body []string
	inExample := false
	keptExact := map[string]bool{}

	flush := func() {
		if !inExample {
			return
		}
		key := strings.ToLower(strings.TrimSpace(strings.Join(body, "\n")))
		switch {
		case removeBody[key]:
			// drop entirely: an approved near-duplicate that isn't the
			// group's kept variant.
		case dropExactBody[key]:
			if !keptExact[key] {
				keptExact[key] = true
				out = append(out, raw...)
			}
		default:
			out = append(out, raw...)
		}
		raw, body = nil, nil
	}

	for _, line := range lines {
		if m := exampleHeaderRe.FindStringSubmatch(line); m != nil {
			flush()
			inExample = true
			raw = []string{line}
			if rest := strings.TrimSpace(m[1]); rest != "" {
				body = append(body, rest)
			}
			continue
		}
		if sectionHeaderRe.MatchString(line) {
			flush()
			inExample = false
			out = append(out, line)
			continue
		}
		if inExample {
			raw = append(raw, line)
			body = append(body, line)
			continue
		}
		out = append(out, line)
	}
	flush()

	return strings.Join(out, "\n")
}
