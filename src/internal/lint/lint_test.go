package lint

import (
	"os"
	"testing"
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
