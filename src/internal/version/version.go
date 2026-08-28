// Package version holds the Distilly build version (YYYYMMDD.N).
package version

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
)

//go:embed VERSION
var versionFile string

// release is set to "1" via -ldflags for release builds.
// Example: -X distilly/internal/version.release=1
var release string

// Base returns the embedded VERSION contents (trimmed). Empty embed yields "0".
func Base() string {
	s := strings.TrimSpace(versionFile)
	if s == "" {
		return "0"
	}
	return s
}

// String returns Base() for release builds, otherwise Base()+"+dev".
func String() string {
	b := Base()
	if release == "1" {
		return b
	}
	return b + "+dev"
}

// NextVersion returns the next YYYYMMDD.N for today given the current version string.
// If current is missing, invalid, or from a different day, returns today+".1".
func NextVersion(today, current string) string {
	today = strings.TrimSpace(today)
	current = strings.TrimSpace(current)
	if today == "" {
		return "0.1"
	}
	date, n, ok := parseVersion(current)
	if !ok || date != today {
		return today + ".1"
	}
	return fmt.Sprintf("%s.%d", today, n+1)
}

func parseVersion(s string) (date string, n int, ok bool) {
	i := strings.LastIndex(s, ".")
	if i <= 0 || i == len(s)-1 {
		return "", 0, false
	}
	date = s[:i]
	if len(date) != 8 {
		return "", 0, false
	}
	for _, c := range date {
		if c < '0' || c > '9' {
			return "", 0, false
		}
	}
	n64, err := strconv.ParseInt(s[i+1:], 10, 64)
	if err != nil || n64 < 1 {
		return "", 0, false
	}
	return date, int(n64), true
}
