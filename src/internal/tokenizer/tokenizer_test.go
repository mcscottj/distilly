package tokenizer

import "testing"

func TestCountEmpty(t *testing.T) {
	if got := Count(""); got != 0 {
		t.Errorf("Count(\"\") = %d, want 0", got)
	}
	if got := Count("   \n\t"); got != 0 {
		t.Errorf("Count(whitespace) = %d, want 0", got)
	}
}

func TestCountKnownString(t *testing.T) {
	// "Hello, world!" is a well-known tiktoken example that encodes to
	// exactly 4 tokens under cl100k_base.
	if got := Count("Hello, world!"); got != 4 {
		t.Errorf(`Count("Hello, world!") = %d, want 4`, got)
	}
}

func TestCountForModelKnownVsUnknown(t *testing.T) {
	n, ok := CountForModel("gpt-4", "Hello, world!")
	if !ok {
		t.Fatal("expected gpt-4 to resolve to a known encoding")
	}
	if n != 4 {
		t.Errorf("CountForModel(gpt-4, ...) = %d, want 4", n)
	}

	_, ok = CountForModel("not-a-real-model", "Hello, world!")
	if ok {
		t.Error("expected unknown model to report ok=false")
	}
}
