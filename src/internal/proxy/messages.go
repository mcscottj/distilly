package proxy

import (
	"strings"

	"distilly/internal/history"
	"distilly/internal/lint"
)

// ChatMessage is one OpenAI chat message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MessagesToPrompt reconstructs a sectioned prompt from chat messages.
//
// Mapping (v1 heuristic):
//   - role=system → System section
//   - preceding user/assistant pairs → History section (User:/Assistant: lines)
//   - final user message → Question section
//
// Non-system roles other than user/assistant are treated as user content.
func MessagesToPrompt(messages []ChatMessage) string {
	var systems []string
	var turns []ChatMessage

	for _, m := range messages {
		switch strings.ToLower(strings.TrimSpace(m.Role)) {
		case "system":
			if strings.TrimSpace(m.Content) != "" {
				systems = append(systems, m.Content)
			}
		default:
			turns = append(turns, m)
		}
	}

	var question string
	if n := len(turns); n > 0 && strings.EqualFold(turns[n-1].Role, "user") {
		question = turns[n-1].Content
		turns = turns[:n-1]
	}

	var b strings.Builder
	if len(systems) > 0 {
		b.WriteString("System:\n")
		b.WriteString(strings.Join(systems, "\n"))
		b.WriteByte('\n')
	}

	if len(turns) > 0 {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("History:\n")
		for _, t := range turns {
			role := normalizeHistoryRole(t.Role)
			b.WriteString(role)
			b.WriteString(": ")
			b.WriteString(singleLine(t.Content))
			b.WriteByte('\n')
		}
	}

	if question != "" {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("Question: ")
		b.WriteString(question)
		b.WriteByte('\n')
	}

	return strings.TrimRight(b.String(), "\n")
}

// PromptToMessages rebuilds chat messages from an optimized sectioned prompt.
// System / History / Question boundaries are preserved best-effort via
// lint.SplitSections and history.ParseTurns.
func PromptToMessages(prompt string) []ChatMessage {
	sections := lint.SplitSections(prompt)
	out := make([]ChatMessage, 0, 4)

	if sys := strings.TrimSpace(sections.System); sys != "" {
		out = append(out, ChatMessage{Role: "system", Content: sys})
	}

	for _, turn := range history.ParseTurns(sections.History) {
		role := strings.ToLower(turn.Role)
		if role != "user" && role != "assistant" && role != "system" {
			role = "user"
		}
		out = append(out, ChatMessage{Role: role, Content: turn.Content})
	}

	if q := strings.TrimSpace(sections.Question); q != "" {
		out = append(out, ChatMessage{Role: "user", Content: q})
	}

	return out
}

func normalizeHistoryRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "assistant":
		return "Assistant"
	case "system":
		return "System"
	default:
		return "User"
	}
}

func singleLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", " ")
}
