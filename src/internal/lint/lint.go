// Package lint orchestrates the v1 checks (token counting, duplicate
// detection, history flagging) and produces a Report. This is the
// "ESLint for prompts" core — fully deterministic, no AI calls.
package lint

import (
	"fmt"
	"io"
	"strings"

	"distilly/internal/cost"
	"distilly/internal/dedupe"
	"distilly/internal/diff"
	"distilly/internal/history"
	"distilly/internal/tokenizer"
)

// Report is the result of linting a single prompt.
type Report struct {
	InputTokens int
	Sections    SectionTokens
	Duplicates  []dedupe.Duplicate
	// NearDuplicates are lower-confidence groups (see dedupe.Duplicate.
	// Confidence) that read as the same instruction but aren't
	// byte/case-identical. These are reported for review, not folded
	// into PotentialSavings/EstimatedSavingsUSD, since they require
	// user approval before removal.
	NearDuplicates []dedupe.Duplicate

	// DuplicateExamples/NearDuplicateExamples are the same idea as
	// Duplicates/NearDuplicates, but clustered at the whole few-shot
	// example level (see SplitExamples) instead of per-line. A single
	// redundant example teaches the model nothing new no matter how
	// different its individual lines look from other sections' content.
	DuplicateExamples     []dedupe.Duplicate
	NearDuplicateExamples []dedupe.Duplicate

	Suggestions      []string
	PotentialSavings float64 // fraction, e.g. 0.46 for 46%

	// Model is the model name passed to Run for cost estimation, or ""
	// if none was given.
	Model string
	// CostKnown is false if Model isn't in cost.Table, in which case
	// EstimatedCostUSD/EstimatedSavingsUSD are zero.
	CostKnown           bool
	EstimatedCostUSD    float64
	EstimatedSavingsUSD float64
}

// Run lints a raw prompt string and returns a Report. model selects a
// pricing entry from internal/cost for cost estimation; pass "" to skip
// cost estimation.
func Run(prompt, model string) Report {
	lines := strings.Split(prompt, "\n")
	sections := SplitSections(prompt)

	total := tokenizer.Count(prompt)
	// Scoped to lines that also occur outside example blocks — see
	// nonExampleLines. Otherwise two genuinely different examples that
	// happen to share one identical line (e.g. the same answer) would
	// be misread as a leaked/duplicated instruction.
	dupes := filterToLinesSeenOutsideExamples(dedupe.FindExact(lines), nonExampleLines(prompt))
	// Near-duplicate detection is scoped to the System section only.
	// Examples and History are expected to vary line-to-line (different
	// Q/A pairs, different turns) — running lexical similarity over them
	// would flag that expected variation as a false positive (e.g.
	// "Example 1:"/"Example 2:" or "...capital of France?"/"...Japan?"
	// score just as similar as a genuinely reworded instruction).
	nearDupes := dedupe.FindNear(strings.Split(sections.System, "\n"), dedupe.DefaultNearThreshold)
	turns := history.ParseTurns(sections.History)

	exampleBodies := SplitExamples(prompt)
	dupeExamples := dedupe.FindExact(exampleBodies)
	nearDupeExamples := dedupe.FindNear(exampleBodies, dedupe.DefaultExampleNearThreshold)

	var suggestions []string
	savedTokens := 0

	if len(dupes) > 0 {
		suggestions = append(suggestions, "Remove duplicate instructions")
		for _, d := range dupes {
			// crude savings estimate: every repeat beyond the first is waste
			for _, l := range d.Lines[1:] {
				savedTokens += tokenizer.Count(l)
			}
		}
	}

	if len(nearDupes) > 0 {
		suggestions = append(suggestions, "Review near-duplicate instructions")
	}

	if len(dupeExamples) > 0 {
		suggestions = append(suggestions, "Remove duplicate examples")
		for _, d := range dupeExamples {
			for _, body := range d.Lines[1:] {
				savedTokens += tokenizer.Count(body)
			}
		}
	}

	if len(nearDupeExamples) > 0 {
		suggestions = append(suggestions, "Review near-duplicate examples")
	}

	if history.Flag(turns) {
		suggestions = append(suggestions, "Compress history")
	}

	var savings float64
	if total > 0 {
		savings = float64(savedTokens) / float64(total)
	}

	var estCost, estSavings float64
	var costKnown bool
	if model != "" {
		if usd, ok := cost.Estimate(model, total, 0); ok {
			costKnown = true
			estCost = usd
			estSavings, _ = cost.Estimate(model, savedTokens, 0)
		}
	}

	return Report{
		InputTokens: total,
		Sections: SectionTokens{
			System:   tokenizer.Count(sections.System),
			Examples: tokenizer.Count(sections.Examples),
			History:  tokenizer.Count(sections.History),
			Question: tokenizer.Count(sections.Question),
		},
		Duplicates:            dupes,
		NearDuplicates:        nearDupes,
		DuplicateExamples:     dupeExamples,
		NearDuplicateExamples: nearDupeExamples,
		Suggestions:           suggestions,
		PotentialSavings:      savings,
		Model:                 model,
		CostKnown:             costKnown,
		EstimatedCostUSD:      estCost,
		EstimatedSavingsUSD:   estSavings,
	}
}

// DiffForDuplicate renders a before/after diff for a single duplicate
// group: every occurrence collapsing down to the one line that would be
// kept.
func DiffForDuplicate(d dedupe.Duplicate) string {
	return diff.Render(diff.Lines(d.Lines, []string{d.Keep}))
}

// ApplyOptions gates the lower-confidence optimization tiers behind
// explicit user approval. The zero value applies only the tier that's
// always safe: exact-duplicate collapse.
type ApplyOptions struct {
	// ApproveNearDuplicates opts in to collapsing near-duplicate lines
	// in the System section (dedupe.Duplicate.Confidence < 1.0) down to
	// each group's Keep line, and near-duplicate example blocks (see
	// SplitExamples) down to each group's Keep example. Off by default:
	// near-duplicates are a cosmetic-similarity guess, not a
	// proven-identical match, so collapsing them without explicit
	// approval risks the over-optimization internal/regression guards
	// against.
	ApproveNearDuplicates bool
}

// Apply returns prompt with exact-duplicate lines collapsed down to
// their first occurrence — the only tier of optimization considered
// safe to auto-apply without review (dedupe.Duplicate.Confidence == 1.0;
// see the confidence-tier design in docs/roadmap.md's Milestone 3
// section). Set opts.ApproveNearDuplicates to also fold in near-duplicate
// lines once a user has reviewed the report's near-duplicate diff.
func Apply(prompt string, opts ApplyOptions) string {
	lines := strings.Split(prompt, "\n")
	dupes := filterToLinesSeenOutsideExamples(dedupe.FindExact(lines), nonExampleLines(prompt))

	dropKey := make(map[string]bool, len(dupes))
	for _, d := range dupes {
		dropKey[strings.ToLower(d.Keep)] = true
	}

	// Near-duplicate groups don't share a single normalized key the way
	// exact-duplicate groups do (that's the point — they're lexically
	// distinct), so track the individual lines to remove instead.
	removeLine := map[string]bool{}
	if opts.ApproveNearDuplicates {
		sections := SplitSections(prompt)
		for _, d := range dedupe.FindNear(strings.Split(sections.System, "\n"), dedupe.DefaultNearThreshold) {
			keepKey := strings.ToLower(strings.TrimSpace(d.Keep))
			for _, l := range d.Lines {
				if key := strings.ToLower(strings.TrimSpace(l)); key != keepKey {
					removeLine[key] = true
				}
			}
		}
	}

	kept := map[string]bool{}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		key := strings.ToLower(strings.TrimSpace(line))
		if removeLine[key] {
			continue
		}
		if dropKey[key] {
			if kept[key] {
				continue
			}
			kept[key] = true
		}
		out = append(out, line)
	}
	prompt = strings.Join(out, "\n")

	// Example blocks: exact duplicates keep their first occurrence and
	// drop the rest (always safe); near-duplicates only drop down to
	// their Keep example once approved.
	exampleBodies := SplitExamples(prompt)
	dropExactBody := map[string]bool{}
	for _, d := range dedupe.FindExact(exampleBodies) {
		dropExactBody[strings.ToLower(strings.TrimSpace(d.Keep))] = true
	}

	removeNearBody := map[string]bool{}
	if opts.ApproveNearDuplicates {
		for _, d := range dedupe.FindNear(exampleBodies, dedupe.DefaultExampleNearThreshold) {
			keepKey := strings.ToLower(strings.TrimSpace(d.Keep))
			for _, b := range d.Lines {
				if key := strings.ToLower(strings.TrimSpace(b)); key != keepKey {
					removeNearBody[key] = true
				}
			}
		}
	}

	return removeExampleBlocks(prompt, dropExactBody, removeNearBody)
}

// Print writes a human-readable report to w, in the style sketched out
// in docs/roadmap.md.
func (r Report) Print(w io.Writer) {
	fmt.Fprintf(w, "Input Tokens: %d\n\n", r.InputTokens)

	fmt.Fprintln(w, "Sections")
	fmt.Fprintln(w, "--------")
	fmt.Fprintf(w, "System Prompt   %d\n", r.Sections.System)
	fmt.Fprintf(w, "Examples        %d\n", r.Sections.Examples)
	fmt.Fprintf(w, "History         %d\n", r.Sections.History)
	fmt.Fprintf(w, "Question        %d\n", r.Sections.Question)
	fmt.Fprintln(w)

	if len(r.Duplicates) > 0 {
		fmt.Fprintln(w, "Duplicate Instructions")
		fmt.Fprintln(w, "----------------------")
		for _, d := range r.Duplicates {
			fmt.Fprintf(w, "Keep: %q (found %d times)\n", d.Keep, len(d.Lines))
		}
		fmt.Fprintln(w)

		fmt.Fprintln(w, "Diff (safe to auto-apply)")
		fmt.Fprintln(w, "-------------------------")
		for _, d := range r.Duplicates {
			fmt.Fprint(w, DiffForDuplicate(d))
		}
		fmt.Fprintln(w)
	}

	if len(r.NearDuplicates) > 0 {
		fmt.Fprintln(w, "Near-Duplicate Instructions (review before removing)")
		fmt.Fprintln(w, "-----------------------------------------------------")
		for _, d := range r.NearDuplicates {
			fmt.Fprintf(w, "Keep: %q (%.0f%% confidence, %d similar lines)\n", d.Keep, d.Confidence*100, len(d.Lines))
			for _, l := range d.Lines {
				fmt.Fprintf(w, "  - %q\n", l)
			}
		}
		fmt.Fprintln(w)

		fmt.Fprintln(w, "Suggested Diff (not applied automatically)")
		fmt.Fprintln(w, "-------------------------------------------")
		for _, d := range r.NearDuplicates {
			fmt.Fprint(w, DiffForDuplicate(d))
		}
		fmt.Fprintln(w)
	}

	if len(r.DuplicateExamples) > 0 {
		fmt.Fprintln(w, "Duplicate Examples")
		fmt.Fprintln(w, "------------------")
		for _, d := range r.DuplicateExamples {
			fmt.Fprintf(w, "Keep:\n%s\n(found %d times)\n\n", d.Keep, len(d.Lines))
		}

		fmt.Fprintln(w, "Diff (safe to auto-apply)")
		fmt.Fprintln(w, "-------------------------")
		for _, d := range r.DuplicateExamples {
			fmt.Fprint(w, DiffForDuplicate(d))
		}
		fmt.Fprintln(w)
	}

	if len(r.NearDuplicateExamples) > 0 {
		fmt.Fprintln(w, "Near-Duplicate Examples (review before removing)")
		fmt.Fprintln(w, "--------------------------------------------------")
		for _, d := range r.NearDuplicateExamples {
			fmt.Fprintf(w, "Keep (%.0f%% confidence, %d similar examples):\n%s\n\n", d.Confidence*100, len(d.Lines), d.Keep)
		}

		fmt.Fprintln(w, "Suggested Diff (not applied automatically)")
		fmt.Fprintln(w, "-------------------------------------------")
		for _, d := range r.NearDuplicateExamples {
			fmt.Fprint(w, DiffForDuplicate(d))
		}
		fmt.Fprintln(w)
	}

	if len(r.Suggestions) > 0 {
		fmt.Fprintln(w, "Suggestions")
		fmt.Fprintln(w, "-----------")
		for _, s := range r.Suggestions {
			fmt.Fprintf(w, "\u2713 %s\n", s)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "Potential Savings: %.0f%%\n", r.PotentialSavings*100)

	if r.Model != "" {
		if r.CostKnown {
			fmt.Fprintf(w, "Estimated Cost (%s): $%.4f", r.Model, r.EstimatedCostUSD)
			if r.EstimatedSavingsUSD > 0 {
				fmt.Fprintf(w, " (save ~$%.4f)", r.EstimatedSavingsUSD)
			}
			fmt.Fprintln(w)
		} else {
			fmt.Fprintf(w, "Estimated Cost: unknown model %q\n", r.Model)
		}
	}
}
