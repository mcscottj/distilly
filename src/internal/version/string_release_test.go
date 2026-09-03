package version

import (
	"strings"
	"testing"
)

func TestString_releaseOmitsDevSuffix(t *testing.T) {
	prev := release
	release = "1"
	t.Cleanup(func() { release = prev })

	s := String()
	if s != Base() {
		t.Fatalf("String() = %q want Base() %q", s, Base())
	}
	if strings.HasSuffix(s, "+dev") {
		t.Fatalf("release String() should not end with +dev: %q", s)
	}
}
