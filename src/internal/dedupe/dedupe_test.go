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

func TestFindNearCatchesPunctuationAndCasingVariants(t *testing.T) {
	lines := []string{
		"Do not explain your reasoning, just answer.",
		"Do not explain your reasoning; just answer",
	}

	dupes := FindNear(lines, DefaultNearThreshold)

	if len(dupes) != 1 {
		t.Fatalf("expected 1 near-duplicate group, got %d: %+v", len(dupes), dupes)
	}
	if dupes[0].Confidence != 1.0 {
		t.Errorf("expected confidence 1.0 for punctuation-only variants, got %f", dupes[0].Confidence)
	}
}

func TestFindNearCatchesMinorRewording(t *testing.T) {
	lines := []string{
		"Always respond in JSON format.",
		"Please always respond in JSON format.",
	}

	dupes := FindNear(lines, DefaultNearThreshold)

	if len(dupes) != 1 {
		t.Fatalf("expected 1 near-duplicate group, got %d: %+v", len(dupes), dupes)
	}
	if dupes[0].Confidence <= 0 || dupes[0].Confidence >= 1.0 {
		t.Errorf("expected confidence strictly between 0 and 1 for a partial rewording, got %f", dupes[0].Confidence)
	}
}

func TestFindNearNoFalsePositivesOnSemanticallySimilarLines(t *testing.T) {
	// These express the same intent but share little lexical overlap.
	// Catching this requires embeddings/semantic matching, deferred to
	// Milestone 3 — FindNear must not flag them.
	lines := []string{
		"Use markdown.",
		"Respond in markdown.",
		"Format as markdown.",
	}

	dupes := FindNear(lines, DefaultNearThreshold)
	if len(dupes) != 0 {
		t.Fatalf("expected 0 near-duplicate groups for semantically-similar-but-lexically-distinct lines, got %d: %+v", len(dupes), dupes)
	}
}

func TestFindNearExcludesLinesAlreadyExactDuplicates(t *testing.T) {
	lines := []string{
		"Always respond in JSON format.",
		"Always respond in JSON format.",
		"Always respond in JSON format.",
	}

	dupes := FindNear(lines, DefaultNearThreshold)
	if len(dupes) != 0 {
		t.Fatalf("expected FindNear to leave exact-duplicate lines to FindExact, got %d: %+v", len(dupes), dupes)
	}
}
