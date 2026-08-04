package lint

import (
	"os"
	"strings"
	"testing"
)

func TestFindStructuredDataDetectsRunOfKeyValueLines(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/structured_data.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	sections := SplitSections(string(data))
	blocks := FindStructuredData(strings.Split(sections.System, "\n"))

	if len(blocks) != 1 {
		t.Fatalf("expected 1 structured block, got %d: %+v", len(blocks), blocks)
	}

	b := blocks[0]
	wantKeys := []string{"Name", "Age", "City", "Occupation"}
	if len(b.Keys) != len(wantKeys) {
		t.Fatalf("expected keys %v, got %v", wantKeys, b.Keys)
	}
	for i, k := range wantKeys {
		if b.Keys[i] != k {
			t.Errorf("key %d: expected %q, got %q", i, k, b.Keys[i])
		}
	}
}

func TestFindStructuredDataIgnoresShortRuns(t *testing.T) {
	// Only two consecutive key/value lines — below MinStructuredRun,
	// ordinary prose (e.g. a couple of labeled asides) shouldn't be
	// flagged as a data block worth converting.
	lines := []string{
		"You are a helpful assistant.",
		"Name: John Smith",
		"Age: 30",
		"Please be concise.",
	}

	if blocks := FindStructuredData(lines); len(blocks) != 0 {
		t.Errorf("expected no structured blocks below MinStructuredRun, got %+v", blocks)
	}
}

func TestStructuredBlockJSONPreservesKeyOrderAndEscapes(t *testing.T) {
	b := StructuredBlock{
		Keys:   []string{"Name", "Quote"},
		Values: []string{"John Smith", `He said "hi".`},
	}

	got := b.JSON()
	want := `{"Name": "John Smith", "Quote": "He said \"hi\"."}`
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

func TestRunFindsStructuredData(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/structured_data.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	report := Run(string(data), "")

	if len(report.StructuredData) != 1 {
		t.Fatalf("expected 1 structured data block, got %d: %+v", len(report.StructuredData), report.StructuredData)
	}

	found := false
	for _, s := range report.Suggestions {
		if s == "Convert structured data to JSON" {
			found = true
		}
	}
	if !found {
		t.Errorf(`expected "Convert structured data to JSON" suggestion, got %+v`, report.Suggestions)
	}
}

func TestApplyLeavesStructuredDataUntouchedByDefault(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/structured_data.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	optimized := Apply(string(data), ApplyOptions{})

	if optimized != string(data) {
		t.Errorf("expected structured data to be left for user review, but Apply changed the prompt:\n%s", optimized)
	}
}

func TestApplyConvertsStructuredDataWhenApproved(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prompts/structured_data.txt")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	optimized := Apply(string(data), ApplyOptions{ApproveJSONConversion: true})

	if strings.Contains(optimized, "Name: John Smith") {
		t.Errorf("expected the prose key/value lines to be gone, got:\n%s", optimized)
	}
	if !strings.Contains(optimized, `{"Name": "John Smith", "Age": "30", "City": "New York", "Occupation": "Software Engineer"}`) {
		t.Errorf("expected the collapsed JSON object to be present, got:\n%s", optimized)
	}
	// Non-System content must survive untouched.
	if !strings.Contains(optimized, "Summarize this customer's profile in one sentence.") {
		t.Errorf("expected the Question section to survive untouched, got:\n%s", optimized)
	}
}
