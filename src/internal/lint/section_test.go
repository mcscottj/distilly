package lint

import "testing"

func TestSplitSectionsWithExplicitHeaders(t *testing.T) {
	prompt := `System:
You are a helpful assistant.

Example 1:
Q: What is the capital of France?
A: Paris

History:
User: hi
Assistant: hello

Question: What's the weather like?`

	sections := SplitSections(prompt)

	if sections.System != "You are a helpful assistant.\n" {
		t.Errorf("System = %q", sections.System)
	}
	if sections.Examples == "" {
		t.Errorf("expected non-empty Examples section")
	}
	if sections.History == "" {
		t.Errorf("expected non-empty History section")
	}
	if sections.Question != "What's the weather like?" {
		t.Errorf("Question = %q", sections.Question)
	}
}

func TestSplitSectionsWithNoHeadersFallsBackToSystem(t *testing.T) {
	prompt := "You are a helpful assistant.\nAlways respond in JSON."

	sections := SplitSections(prompt)

	if sections.System != prompt {
		t.Errorf("System = %q, want everything in System when no headers are present", sections.System)
	}
	if sections.Examples != "" || sections.History != "" || sections.Question != "" {
		t.Errorf("expected other sections to be empty, got Examples=%q History=%q Question=%q",
			sections.Examples, sections.History, sections.Question)
	}
}
