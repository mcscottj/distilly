// Package history flags and (eventually) compresses long conversation
// history sections within a prompt — e.g. collapsing 42 turns down to
// a bullet-point summary that preserves key context.
package history

import (
	"regexp"
	"strings"
)

// Turn is a single message in a conversation history.
type Turn struct {
	Role    string
	Content string
}

// Flag reports whether a history of this many turns is worth flagging
// to the user as a compression candidate. The threshold is a starting
// point, not tuned against real data yet.
func Flag(turns []Turn) bool {
	const threshold = 10
	return len(turns) > threshold
}

// turnRe matches a single "Role: content" history line, e.g.
// "User: how do I..." / "Assistant: you can...".
var turnRe = regexp.MustCompile(`(?i)^\s*(user|assistant|system)\s*:\s*(.*)$`)

// ParseTurns parses a raw History section into Turns, one per line in the
// form "Role: content". Lines that don't match this shape are ignored.
//
// This is a v1 heuristic: it doesn't support multi-line turn content.
func ParseTurns(section string) []Turn {
	var turns []Turn
	for _, line := range strings.Split(section, "\n") {
		m := turnRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		turns = append(turns, Turn{Role: capitalize(m[1]), Content: strings.TrimSpace(m[2])})
	}
	return turns
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// TODO(milestone 3): AI-assisted summarization that preserves key
// context (e.g. "User is debugging Docker. Using Go. Fixed auth issue.
// Current problem is networking.").
