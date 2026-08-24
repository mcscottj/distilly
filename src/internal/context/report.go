package context

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Print writes a human-readable selection report to w.
func (r Result) Print(w io.Writer) {
	fmt.Fprintf(w, "Selected Files (%d)\n", len(r.Files))
	fmt.Fprintln(w, "------------------")
	for _, f := range r.Files {
		reason := formatReasons(f.Reasons)
		fmt.Fprintf(w, "%-40s %10s tokens   %s\n", f.Path, formatInt(f.Tokens), reason)
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Total: %s tokens", formatInt(r.TotalTokens))
	if r.ExcludedCount > 0 {
		fmt.Fprintf(w, " (excluded %d files due to budget)", r.ExcludedCount)
	}
	fmt.Fprintln(w)
	if len(r.Warnings) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Warnings")
		fmt.Fprintln(w, "--------")
		for _, warn := range r.Warnings {
			fmt.Fprintf(w, "- %s\n", warn)
		}
	}
}

func formatReasons(reasons []InclusionReason) string {
	if len(reasons) == 0 {
		return ""
	}
	parts := make([]string, len(reasons))
	for i, r := range reasons {
		switch r.Kind {
		case "import", "reverse_import", "symbol_match":
			if r.Detail != "" {
				parts[i] = r.Kind + ": " + r.Detail
			} else {
				parts[i] = r.Kind
			}
		default:
			if r.Detail != "" {
				parts[i] = r.Detail
			} else {
				parts[i] = r.Kind
			}
		}
	}
	return strings.Join(parts, "; ")
}

func formatInt(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	start := len(s) % 3
	if start > 0 {
		b.WriteString(s[:start])
		if len(s) > start {
			b.WriteByte(',')
		}
	}
	for i := start; i < len(s); i += 3 {
		if i > start {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
