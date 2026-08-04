package lint

import (
	"os"
	"strings"
	"testing"

	"distilly/internal/dedupe"
)

func TestRunFindsExactDuplicates(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/exact_duplicates.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "")

	if len(report.Duplicates) == 0 {
		t.Fatal("expected at least one duplicate group, got none")
	}
	if report.PotentialSavings <= 0 {
		t.Fatalf("expected positive potential savings, got %f", report.PotentialSavings)
	}
}

func TestRunNoFalsePositivesOnSimilarButDistinctLines(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/example.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "")

	// example.txt has semantically similar but textually distinct lines
	// (e.g. "Use markdown." / "Respond in markdown." / "Format as markdown.").
	// v1's exact-match dedupe must not flag these as duplicates — that's
	// near-duplicate detection, deferred to Milestone 3.
	if len(report.Duplicates) != 0 {
		t.Fatalf("expected no exact duplicates in example.txt, got %+v", report.Duplicates)
	}
}

func TestRunFlagsLongHistoryAndPopulatesSections(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/long_history.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "")

	if report.Sections.System == 0 {
		t.Error("expected non-zero System section tokens")
	}
	if report.Sections.History == 0 {
		t.Error("expected non-zero History section tokens")
	}
	if report.Sections.Question == 0 {
		t.Error("expected non-zero Question section tokens")
	}

	found := false
	for _, s := range report.Suggestions {
		if s == "Compress history" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected \"Compress history\" suggestion, got %+v", report.Suggestions)
	}
}

func TestRunEstimatesCostForKnownModel(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/exact_duplicates.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "gpt-4")

	if !report.CostKnown {
		t.Fatal("expected gpt-4 to be a known model")
	}
	if report.EstimatedCostUSD <= 0 {
		t.Errorf("expected positive estimated cost, got %f", report.EstimatedCostUSD)
	}
	if report.EstimatedSavingsUSD <= 0 {
		t.Errorf("expected positive estimated savings given duplicate lines, got %f", report.EstimatedSavingsUSD)
	}
}

func TestRunFindsNearDuplicates(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/near_duplicates.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "")

	if len(report.NearDuplicates) != 2 {
		t.Fatalf("expected 2 near-duplicate groups, got %d: %+v", len(report.NearDuplicates), report.NearDuplicates)
	}

	found := false
	for _, s := range report.Suggestions {
		if s == "Review near-duplicate instructions" {
			found = true
		}
	}
	if !found {
		t.Errorf(`expected "Review near-duplicate instructions" suggestion, got %+v`, report.Suggestions)
	}
}

func TestRunNearDuplicatesIgnoreVariationInExamples(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/exact_duplicates.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "")

	// exact_duplicates.txt has no explicit "System:" header, so everything
	// from the first "Example N:" line onward — including enumerated
	// "Example 1:"/"Example 2:" headers and "Q: ...France?"/"...Japan?"
	// lines — falls into the Examples section. Near-duplicate detection
	// must not flag that expected example-to-example variation.
	if len(report.NearDuplicates) != 0 {
		t.Fatalf("expected 0 near-duplicate groups (examples should vary freely), got %d: %+v",
			len(report.NearDuplicates), report.NearDuplicates)
	}
}

func TestDiffForDuplicateShowsEveryOccurrenceCollapsingToKeep(t *testing.T) {
	d := dedupe.Duplicate{
		Lines: []string{
			"Always respond in JSON format.",
			"Always respond in JSON format.",
			"Always respond in JSON format.",
		},
		Keep:       "Always respond in JSON format.",
		Confidence: 1.0,
	}

	got := DiffForDuplicate(d)

	if strings.Count(got, "- Always respond in JSON format.") != 3 {
		t.Errorf("expected 3 removed lines in diff, got:\n%s", got)
	}
	if strings.Count(got, "+ Always respond in JSON format.") != 1 {
		t.Errorf("expected 1 added (kept) line in diff, got:\n%s", got)
	}
}

func TestApplyCollapsesExactDuplicatesKeepingFirstOccurrence(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/exact_duplicates.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	optimized := Apply(string(data), ApplyOptions{})

	if strings.Count(optimized, "Always respond in JSON format.") != 1 {
		t.Errorf("expected exactly 1 occurrence of the collapsed line, got:\n%s", optimized)
	}

	report := Run(optimized, "")
	if len(report.Duplicates) != 0 {
		t.Errorf("expected Apply's output to have no remaining exact duplicates, got %+v", report.Duplicates)
	}
}

func TestApplyLeavesNearDuplicatesUntouched(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/near_duplicates.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	optimized := Apply(string(data), ApplyOptions{})

	if optimized != string(data) {
		t.Errorf("expected near-duplicates to be left for user review, but Apply changed the prompt:\n%s", optimized)
	}
}

func TestApplyCollapsesNearDuplicatesWhenApproved(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/near_duplicates.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	optimized := Apply(string(data), ApplyOptions{ApproveNearDuplicates: true})

	report := Run(optimized, "")
	if len(report.NearDuplicates) != 0 {
		t.Errorf("expected approved near-duplicates to be collapsed, got %+v", report.NearDuplicates)
	}

	// Each group's Keep line (the longest variant) must still be present.
	if !strings.Contains(optimized, "Please always respond in JSON format.") {
		t.Errorf("expected the kept variant of the JSON group to survive, got:\n%s", optimized)
	}
	if !strings.Contains(optimized, "Do not explain your reasoning, just answer.") {
		t.Errorf("expected the kept variant of the reasoning group to survive, got:\n%s", optimized)
	}
}

func TestRunFindsDuplicateAndNearDuplicateExamples(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/redundant_examples.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "")

	if len(report.DuplicateExamples) != 1 {
		t.Fatalf("expected 1 exact-duplicate example group (the two identical France examples), got %d: %+v",
			len(report.DuplicateExamples), report.DuplicateExamples)
	}
	if len(report.NearDuplicateExamples) != 1 {
		t.Fatalf("expected 1 near-duplicate example group (the two 2+2 examples), got %d: %+v",
			len(report.NearDuplicateExamples), report.NearDuplicateExamples)
	}

	for _, want := range []string{"Remove duplicate examples", "Review near-duplicate examples"} {
		found := false
		for _, s := range report.Suggestions {
			if s == want {
				found = true
			}
		}
		if !found {
			t.Errorf("expected %q suggestion, got %+v", want, report.Suggestions)
		}
	}
}

// TestRunExactDuplicateLinesIgnoreCoincidentalOverlapBetweenExamples
// guards a real bug found while building example-block clustering:
// line-level exact-duplicate detection used to run over every line in
// the prompt regardless of section, so two genuinely different examples
// that happened to share one identical line (here, both answering
// "4") would have that shared line silently stripped out of the second
// example — corrupting it — even though the examples themselves are not
// duplicates.
func TestRunExactDuplicateLinesIgnoreCoincidentalOverlapBetweenExamples(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/redundant_examples.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "")

	for _, d := range report.Duplicates {
		if strings.Contains(d.Keep, `"answer": 4`) {
			t.Errorf(`did not expect A: {"answer": 4} to be treated as a leaked/duplicated line — `+
				`it only repeats between two distinct examples, got %+v`, d)
		}
	}
}

func TestApplyDoesNotCorruptDistinctExamplesSharingALine(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/redundant_examples.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	optimized := Apply(string(data), ApplyOptions{})

	// Example 3 ("What's the sum of 2 and 2?") is a near-duplicate of
	// Example 1, not an exact one, so it must survive Apply() with the
	// zero-value ApplyOptions untouched — including its answer line,
	// which coincidentally matches Example 1's.
	if !strings.Contains(optimized, "Q: What's the sum of 2 and 2?") {
		t.Fatalf("expected Example 3's question to survive, got:\n%s", optimized)
	}
	block := optimized[strings.Index(optimized, "Q: What's the sum of 2 and 2?"):]
	if !strings.Contains(block, `A: {"answer": 4}`) {
		t.Errorf("expected Example 3's answer line to survive alongside its question, got:\n%s", block)
	}
}

func TestApplyCollapsesDuplicateExamplesKeepingFirstOccurrence(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/redundant_examples.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	optimized := Apply(string(data), ApplyOptions{})

	if strings.Count(optimized, "capital of France") != 1 {
		t.Errorf("expected the exact-duplicate France example to collapse to 1 occurrence, got:\n%s", optimized)
	}
	// The near-duplicate "sum of 2 and 2" examples are untouched by
	// default — both must survive.
	if strings.Count(optimized, "sum of 2 and 2") != 2 {
		t.Errorf("expected both near-duplicate examples to survive without approval, got:\n%s", optimized)
	}
}

func TestApplyCollapsesNearDuplicateExamplesWhenApproved(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/redundant_examples.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	optimized := Apply(string(data), ApplyOptions{ApproveNearDuplicates: true})

	if strings.Count(optimized, "sum of 2 and 2") != 1 {
		t.Errorf("expected the near-duplicate examples to collapse to 1 once approved, got:\n%s", optimized)
	}
	if !strings.Contains(optimized, "capital of France") {
		t.Errorf("expected the distinct France example to survive, got:\n%s", optimized)
	}
}

func TestRunReportsUnknownModel(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/example.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "not-a-real-model")

	if report.CostKnown {
		t.Fatal("expected not-a-real-model to be unknown")
	}
	if report.EstimatedCostUSD != 0 {
		t.Errorf("expected zero cost for unknown model, got %f", report.EstimatedCostUSD)
	}
}
