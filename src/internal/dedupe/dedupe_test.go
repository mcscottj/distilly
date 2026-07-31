package dedupe

import "testing"

func TestFindExact(t *testing.T) {
	lines := []string{
		"You are a helpful assistant.",
		"Always respond in JSON format.",
		"Do not include markdown formatting in your response.",
		"",
		"Always respond in JSON format.",
		"Do not include markdown formatting in your response.",
		"Always respond in JSON format.",
	}

	dupes := FindExact(lines)

	if len(dupes) != 2 {
		t.Fatalf("expected 2 duplicate groups, got %d: %+v", len(dupes), dupes)
	}

	byKeep := map[string]int{}
	for _, d := range dupes {
		byKeep[d.Keep] = len(d.Lines)
	}

	if byKeep["Always respond in JSON format."] != 3 {
		t.Errorf("expected 'Always respond in JSON format.' to appear 3 times, got %d",
			byKeep["Always respond in JSON format."])
	}
	if byKeep["Do not include markdown formatting in your response."] != 2 {
		t.Errorf("expected markdown instruction to appear 2 times, got %d",
			byKeep["Do not include markdown formatting in your response."])
	}
}

func TestFindExactNoDuplicates(t *testing.T) {
	lines := []string{
		"Always be polite.",
		"Always answer professionally.",
		"Never be rude.",
	}

	dupes := FindExact(lines)
	if len(dupes) != 0 {
		t.Fatalf("expected 0 duplicate groups for all-unique lines, got %d: %+v", len(dupes), dupes)
	}
}
