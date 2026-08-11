package api

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"distilly/internal/cost"
)

func TestListModelsReturnsCostTableKeys(t *testing.T) {
	models := ListModels()
	if len(models) != len(cost.Table) {
		t.Fatalf("ListModels() returned %d models, want %d", len(models), len(cost.Table))
	}
	for name := range cost.Table {
		if !slices.Contains(models, name) {
			t.Errorf("ListModels() missing %q", name)
		}
	}
	if !slices.IsSorted(models) {
		t.Errorf("ListModels() not sorted: %v", models)
	}
}

func TestAnalyzeMapsReportFields(t *testing.T) {
	prompt := readFixture(t, "exact_duplicates.txt")

	resp := Analyze(AnalyzeRequest{Prompt: prompt, Model: "gpt-4"})

	if resp.InputTokens <= 0 {
		t.Fatalf("InputTokens = %d, want > 0", resp.InputTokens)
	}
	if len(resp.Duplicates) == 0 {
		t.Fatal("expected at least one exact duplicate group")
	}
	for _, d := range resp.Duplicates {
		if d.Confidence != 1.0 {
			t.Errorf("exact duplicate Confidence = %f, want 1.0", d.Confidence)
		}
		if d.Keep == "" {
			t.Error("exact duplicate Keep is empty")
		}
		if len(d.Lines) < 2 {
			t.Errorf("exact duplicate Lines = %v, want >= 2", d.Lines)
		}
		if len(d.Diff) == 0 {
			t.Error("exact duplicate Diff is empty")
		}
		assertDiffMarkers(t, d.Diff)
	}
	if len(resp.Suggestions) == 0 {
		t.Fatal("expected suggestions")
	}
	if resp.PotentialSavings <= 0 {
		t.Fatalf("PotentialSavings = %f, want > 0", resp.PotentialSavings)
	}
	if !resp.CostKnown {
		t.Fatal("CostKnown = false for gpt-4")
	}
	if resp.EstimatedCostUSD <= 0 {
		t.Errorf("EstimatedCostUSD = %f, want > 0", resp.EstimatedCostUSD)
	}
	if resp.Model != "gpt-4" {
		t.Errorf("Model = %q, want gpt-4", resp.Model)
	}
	if resp.Sections.System == 0 && resp.Sections.Examples == 0 &&
		resp.Sections.History == 0 && resp.Sections.Question == 0 {
		t.Error("expected non-zero section breakdown")
	}
	if resp.Score >= 100 {
		t.Errorf("Score = %d, want < 100 for exact-duplicates fixture", resp.Score)
	}
	if len(resp.Issues) == 0 {
		t.Fatal("expected Issues for exact-duplicates fixture")
	}
}

func TestAnalyzeExposesPerfectScoreForCleanPrompt(t *testing.T) {
	resp := Analyze(AnalyzeRequest{
		Prompt: "System:\nYou are a helpful assistant.\n\nQuestion:\nWhat is 2+2?\n",
		Model:  "gpt-4",
	})
	if resp.Score != 100 {
		t.Errorf("Score = %d, want 100", resp.Score)
	}
	if len(resp.Issues) != 0 {
		t.Errorf("Issues = %v, want empty", resp.Issues)
	}
}

func TestAnalyzeIncludesNearDuplicatesAndStructuredData(t *testing.T) {
	near := Analyze(AnalyzeRequest{Prompt: readFixture(t, "near_duplicates.txt")})
	if len(near.NearDuplicates) == 0 {
		t.Fatal("expected near-duplicate groups")
	}
	for _, d := range near.NearDuplicates {
		if d.Confidence <= 0 || d.Confidence > 1 {
			t.Errorf("near-duplicate Confidence = %f, want in (0,1]", d.Confidence)
		}
		if len(d.Diff) == 0 {
			t.Error("near-duplicate Diff is empty")
		}
	}

	structured := Analyze(AnalyzeRequest{Prompt: readFixture(t, "structured_data.txt")})
	if len(structured.StructuredData) == 0 {
		t.Fatal("expected structured data blocks")
	}
	for _, b := range structured.StructuredData {
		if b.JSON == "" {
			t.Error("StructuredData.JSON is empty")
		}
		if len(b.Keys) == 0 || len(b.Raw) == 0 {
			t.Errorf("StructuredData incomplete: %+v", b)
		}
		if len(b.Diff) == 0 {
			t.Error("StructuredData.Diff is empty")
		}
		assertDiffMarkers(t, b.Diff)
	}
}

func TestApplyReturnsOptimizedAndStructuredDiff(t *testing.T) {
	prompt := readFixture(t, "exact_duplicates.txt")

	resp := Apply(ApplyRequest{Prompt: prompt})

	if resp.Optimized == "" {
		t.Fatal("Optimized is empty")
	}
	if resp.Optimized == prompt {
		t.Fatal("Optimized equals original; expected exact duplicates collapsed")
	}
	if strings.Count(resp.Optimized, "Always respond in JSON format.") >=
		strings.Count(prompt, "Always respond in JSON format.") {
		t.Fatal("expected fewer duplicate lines after Apply")
	}
	if len(resp.FullDiff) == 0 {
		t.Fatal("FullDiff is empty")
	}
	assertDiffMarkers(t, resp.FullDiff)

	hasMinus, hasPlus := false, false
	for _, line := range resp.FullDiff {
		switch line.Marker {
		case "-":
			hasMinus = true
		case "+":
			hasPlus = true
		}
	}
	if !hasMinus || !hasPlus {
		t.Fatalf("FullDiff missing +/- lines: %+v", resp.FullDiff)
	}
}

func TestApplyRespectsApprovalFlags(t *testing.T) {
	prompt := readFixture(t, "near_duplicates.txt")

	safe := Apply(ApplyRequest{Prompt: prompt})
	approved := Apply(ApplyRequest{
		Prompt:                prompt,
		ApproveNearDuplicates: true,
	})

	if approved.Optimized == safe.Optimized {
		// Near-dupe fixture should change when approval is on.
		t.Fatal("ApproveNearDuplicates=true produced same output as default Apply")
	}
}

func TestDiffForDuplicateReturnsStructuredLines(t *testing.T) {
	prompt := readFixture(t, "exact_duplicates.txt")
	resp := Analyze(AnalyzeRequest{Prompt: prompt})
	if len(resp.Duplicates) == 0 {
		t.Fatal("need a duplicate group")
	}

	lines := DiffForDuplicate(resp.Duplicates[0])
	if len(lines) == 0 {
		t.Fatal("DiffForDuplicate returned empty")
	}
	assertDiffMarkers(t, lines)
}

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "prompts", name))
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	return string(data)
}

func assertDiffMarkers(t *testing.T, lines []DiffLine) {
	t.Helper()
	for i, line := range lines {
		switch line.Marker {
		case " ", "-", "+":
			// ok
		default:
			t.Errorf("lines[%d].Marker = %q, want space/-/+", i, line.Marker)
		}
	}
}
