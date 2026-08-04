package regression

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"distilly/internal/lint"
)

func TestOptimizationPreservesConstraints(t *testing.T) {
	for _, c := range Cases {
		t.Run(c.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("../../testdata/prompts", c.PromptFile))
			if err != nil {
				t.Fatalf("reading %s: %v", c.PromptFile, err)
			}

			optimized := lint.Apply(string(data))

			for _, constraint := range c.MustSurvive {
				if !strings.Contains(optimized, constraint) {
					t.Errorf("optimization dropped constraint %q\n\noptimized output:\n%s", constraint, optimized)
				}
			}
		})
	}
}

// TestApplyIsIdempotent guards the harness itself: a deterministic
// optimizer that's already collapsed everything it's confident about
// should produce no further changes on a second pass. A violation here
// usually means Apply is order-dependent or comparing on the wrong key.
func TestApplyIsIdempotent(t *testing.T) {
	for _, c := range Cases {
		t.Run(c.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("../../testdata/prompts", c.PromptFile))
			if err != nil {
				t.Fatalf("reading %s: %v", c.PromptFile, err)
			}

			once := lint.Apply(string(data))
			twice := lint.Apply(once)

			if once != twice {
				t.Errorf("Apply is not idempotent:\n--- first pass ---\n%s\n--- second pass ---\n%s", once, twice)
			}
		})
	}
}
