package lint

import (
	"regexp"
	"strings"
)

// Sections holds the raw text of each recognized prompt section, as
// produced by SplitSections.
type Sections struct {
	System   string
	Examples string
	History  string
	Question string
}

// SectionTokens holds a token count for each recognized section.
type SectionTokens struct {
	System   int
	Examples int
	History  int
	Question int
}

// sectionHeaderRe matches a line that declares a new section, e.g.
// "System:", "Examples:", "Example 1:", "History:", "Question: ...".
var sectionHeaderRe = regexp.MustCompile(`(?i)^\s*(system(?:\s+prompt)?|examples?|history|question)\s*\d*\s*:\s*(.*)$`)

// SplitSections splits a raw prompt into System/Examples/History/Question
// sections based on explicit headers. Text before the first header (or an
// entire prompt with no headers) is treated as the System section.
//
// This is a v1 heuristic tied to explicit header lines — it does not infer
// structure from unlabeled prompts. Smarter inference is deferred to a
// later milestone.
func SplitSections(prompt string) Sections {
	buffers := map[string][]string{
		"system":   nil,
		"examples": nil,
		"history":  nil,
		"question": nil,
	}
	current := "system"

	for _, line := range strings.Split(prompt, "\n") {
		if m := sectionHeaderRe.FindStringSubmatch(line); m != nil {
			current = normalizeSectionName(m[1])
			if rest := strings.TrimSpace(m[2]); rest != "" {
				buffers[current] = append(buffers[current], rest)
			}
			continue
		}
		buffers[current] = append(buffers[current], line)
	}

	return Sections{
		System:   strings.Join(buffers["system"], "\n"),
		Examples: strings.Join(buffers["examples"], "\n"),
		History:  strings.Join(buffers["history"], "\n"),
		Question: strings.Join(buffers["question"], "\n"),
	}
}

func normalizeSectionName(header string) string {
	switch h := strings.ToLower(strings.TrimSpace(header)); {
	case strings.HasPrefix(h, "system"):
		return "system"
	case strings.HasPrefix(h, "example"):
		return "examples"
	case h == "history":
		return "history"
	case h == "question":
		return "question"
	default:
		return "system"
	}
}
