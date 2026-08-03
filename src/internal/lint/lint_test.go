package lint

import (
	"os"
	"strings"
	"testing"

	"github.com/smcguire/distilly/internal/dedupe"
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
