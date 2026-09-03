package version_test

import (
	"strings"
	"testing"

	"distilly/internal/version"
)

func TestNextVersion_newDay(t *testing.T) {
	got := version.NextVersion("20260828", "20260827.3")
	if got != "20260828.1" {
		t.Fatalf("got %q want %q", got, "20260828.1")
	}
}

func TestNextVersion_sameDay(t *testing.T) {
	got := version.NextVersion("20260828", "20260828.1")
	if got != "20260828.2" {
		t.Fatalf("got %q want %q", got, "20260828.2")
	}
}

func TestNextVersion_emptyOrInvalid(t *testing.T) {
	for _, cur := range []string{"", "bogus", "20260828", "20260828.", "20260828.x"} {
		got := version.NextVersion("20260828", cur)
		if got != "20260828.1" {
			t.Fatalf("cur=%q got %q want %q", cur, got, "20260828.1")
		}
	}
}

func TestString_devSuffixByDefault(t *testing.T) {
	s := version.String()
	if !strings.HasSuffix(s, "+dev") {
		t.Fatalf("String() = %q, want +dev suffix", s)
	}
	base := version.Base()
	if base == "" || strings.Contains(base, "+") {
		t.Fatalf("Base() = %q, want non-empty without +", base)
	}
	if s != base+"+dev" {
		t.Fatalf("String() = %q, want Base()+dev = %q", s, base+"+dev")
	}
}
