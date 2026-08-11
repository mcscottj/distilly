package lint

import (
	"os"
	"strings"
	"testing"

	"distilly/internal/dedupe"
)

func TestScoreCleanPromptIsPerfect(t *testing.T) {
	got := Score(Report{})

	if got.Score != 100 {
		t.Errorf("Score = %d, want 100", got.Score)
	}
	if len(got.Issues) != 0 {
		t.Errorf("Issues = %v, want empty", got.Issues)
	}
}

func TestScoreExactDuplicatesDeduct(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/exact_duplicates.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}
	report := Run(string(data), "")
	if len(report.Duplicates) == 0 {
		t.Fatal("fixture must have exact duplicates")
	}

	got := Score(report)

	wantPenalty := penaltyExactDupe * len(report.Duplicates)
	wantScore := 100 - wantPenalty
	if got.Score != wantScore {
		t.Errorf("Score = %d, want %d", got.Score, wantScore)
	}
	if !hasIssueContaining(got.Issues, "exact duplicate") {
		t.Errorf("Issues = %v, want an exact-duplicate issue", got.Issues)
	}
}

func TestScoreNearDuplicatesDeduct(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/near_duplicates.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}
	report := Run(string(data), "")
	if len(report.NearDuplicates) == 0 {
		t.Fatal("fixture must have near duplicates")
	}

	got := Score(report)

	if got.Score >= 100 {
		t.Errorf("Score = %d, want < 100", got.Score)
	}
	if !hasIssueContaining(got.Issues, "near-duplicate") {
		t.Errorf("Issues = %v, want a near-duplicate issue", got.Issues)
	}
}

func TestScoreRedundantExamplesDeduct(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/redundant_examples.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}
	report := Run(string(data), "")
	if len(report.DuplicateExamples) == 0 && len(report.NearDuplicateExamples) == 0 {
		t.Fatal("fixture must have redundant examples")
	}

	got := Score(report)

	if got.Score >= 100 {
		t.Errorf("Score = %d, want < 100", got.Score)
	}
	if !hasIssueContaining(got.Issues, "example") {
		t.Errorf("Issues = %v, want a redundant-examples issue", got.Issues)
	}
}

func TestScoreStructuredDataDeduct(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/structured_data.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}
	report := Run(string(data), "")
	if len(report.StructuredData) == 0 {
		t.Fatal("fixture must have structured data")
	}

	got := Score(report)

	if got.Score >= 100 {
		t.Errorf("Score = %d, want < 100", got.Score)
	}
	if !hasIssueContaining(got.Issues, "structured data") {
		t.Errorf("Issues = %v, want a structured-data issue", got.Issues)
	}
}

func TestScoreLongHistoryDeduct(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/long_history.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}
	report := Run(string(data), "")

	got := Score(report)

	if got.Score != 100-penaltyLongHistory {
		t.Errorf("Score = %d, want %d", got.Score, 100-penaltyLongHistory)
	}
	if !hasIssueContaining(got.Issues, "history") {
		t.Errorf("Issues = %v, want a long-history issue", got.Issues)
	}
}

func TestScoreClampsAtZero(t *testing.T) {
	// Fabricate a report with enough findings to exceed 100 in penalties.
	report := Report{
		Duplicates:            make([]dedupe.Duplicate, 5),
		NearDuplicates:        make([]dedupe.Duplicate, 5),
		DuplicateExamples:     make([]dedupe.Duplicate, 5),
		NearDuplicateExamples: make([]dedupe.Duplicate, 5),
		StructuredData:        make([]StructuredBlock, 5),
		Suggestions:           []string{"Compress history"},
	}

	got := Score(report)

	if got.Score != 0 {
		t.Errorf("Score = %d, want 0 (clamped)", got.Score)
	}
	if len(got.Issues) == 0 {
		t.Fatal("expected issues even when clamped")
	}
}

func hasIssueContaining(issues []string, substr string) bool {
	for _, issue := range issues {
		if strings.Contains(strings.ToLower(issue), substr) {
			return true
		}
	}
	return false
}
