// Package diff renders a simple line-based before/after diff, used to
// show the user exactly what an optimization changed (e.g. three
// redundant instructions collapsing into one).
package diff

import "fmt"

// Line is a single rendered diff line with its marker: " " (unchanged),
// "-" (removed), or "+" (added).
type Line struct {
	Marker  string
	Content string
}

// Lines returns a naive line-based diff between before and after.
// This is intentionally simple (no LCS/Myers algorithm) — good enough
// for short instruction blocks; revisit if diffing full prompts.
func Lines(before, after []string) []Line {
	var out []Line
	for _, b := range before {
		out = append(out, Line{Marker: "-", Content: b})
	}
	for _, a := range after {
		out = append(out, Line{Marker: "+", Content: a})
	}
	return out
}

// Render formats diff lines the way the CLI/report prints them.
func Render(lines []Line) string {
	s := ""
	for _, l := range lines {
		s += fmt.Sprintf("%s %s\n", l.Marker, l.Content)
	}
	return s
}
