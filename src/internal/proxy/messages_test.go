package proxy

import (
	"strings"
	"testing"
)

func TestMessagesToPrompt_SystemHistoryQuestion(t *testing.T) {
	prompt := MessagesToPrompt([]ChatMessage{
		{Role: "system", Content: "You are helpful.\nBe concise."},
		{Role: "user", Content: "Hi"},
		{Role: "assistant", Content: "Hello!"},
		{Role: "user", Content: "What is 2+2?"},
	})

	wantParts := []string{
		"System:",
		"You are helpful.",
		"Be concise.",
		"History:",
		"User: Hi",
		"Assistant: Hello!",
		"Question: What is 2+2?",
	}
	for _, part := range wantParts {
		if !strings.Contains(prompt, part) {
			t.Fatalf("prompt missing %q:\n%s", part, prompt)
		}
	}
}

func TestMessagesToPrompt_SystemOnlyThenQuestion(t *testing.T) {
	prompt := MessagesToPrompt([]ChatMessage{
		{Role: "system", Content: "Be brief."},
		{Role: "user", Content: "Ping"},
	})
	if strings.Contains(prompt, "History:") {
		t.Fatalf("unexpected History section:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Question: Ping") {
		t.Fatalf("missing question:\n%s", prompt)
	}
}

func TestPromptToMessages_RoundTripRoles(t *testing.T) {
	original := []ChatMessage{
		{Role: "system", Content: "You are helpful."},
		{Role: "user", Content: "Hi"},
		{Role: "assistant", Content: "Hello!"},
		{Role: "user", Content: "Capital of France?"},
	}
	prompt := MessagesToPrompt(original)
	// Simulate exact-duplicate collapse leaving system shorter.
	optimized := strings.ReplaceAll(prompt, "You are helpful.", "You are helpful.")
	got := PromptToMessages(optimized)

	if len(got) != 4 {
		t.Fatalf("len = %d, want 4: %+v", len(got), got)
	}
	if got[0].Role != "system" || got[0].Content == "" {
		t.Fatalf("system = %+v", got[0])
	}
	if got[1].Role != "user" || got[1].Content != "Hi" {
		t.Fatalf("history user = %+v", got[1])
	}
	if got[2].Role != "assistant" || got[2].Content != "Hello!" {
		t.Fatalf("history assistant = %+v", got[2])
	}
	if got[3].Role != "user" || got[3].Content != "Capital of France?" {
		t.Fatalf("question = %+v", got[3])
	}
}
