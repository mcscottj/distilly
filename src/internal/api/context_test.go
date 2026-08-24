package api

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"distilly/internal/context"
)

func TestSelectContextMiniFixture(t *testing.T) {
	context.ClearIndexCache()
	root := filepath.Join("..", "..", "testdata", "repos", "mini")

	resp := SelectContext(SelectContextRequest{
		RepoRoot: root,
		SeedFile: "a/a.go",
		Question: "how does A call B?",
	})
	if resp.Error != "" {
		t.Fatalf("SelectContext error: %s", resp.Error)
	}

	paths := selectedPaths(resp)
	for _, want := range []string{"a/a.go", "b/b.go"} {
		if !slices.Contains(paths, want) {
			t.Errorf("missing %s; got %v", want, paths)
		}
	}
	if slices.Contains(paths, "noise/noise.go") {
		t.Errorf("noise/noise.go should be excluded; got %v", paths)
	}
	if resp.Markdown == "" {
		t.Fatal("expected non-empty markdown")
	}
	if !strings.Contains(resp.Markdown, "### a/a.go") {
		t.Errorf("markdown missing seed file header:\n%s", resp.Markdown)
	}
}

func TestSelectContextReturnsErrorForMissingSeed(t *testing.T) {
	context.ClearIndexCache()
	root := filepath.Join("..", "..", "testdata", "repos", "mini")

	resp := SelectContext(SelectContextRequest{
		RepoRoot: root,
		SeedFile: "missing.go",
		Question: "test",
	})
	if resp.Error == "" {
		t.Fatal("expected error for missing seed")
	}
	if len(resp.Files) != 0 {
		t.Fatalf("expected no files on error, got %v", resp.Files)
	}
}

func selectedPaths(resp SelectContextResponse) []string {
	out := make([]string, len(resp.Files))
	for i, f := range resp.Files {
		out[i] = f.Path
	}
	return out
}

func TestSelectContextDefaultMaxTokens(t *testing.T) {
	context.ClearIndexCache()
	root := filepath.Join("..", "..", "testdata", "repos", "mini")

	resp := SelectContext(SelectContextRequest{
		RepoRoot:  root,
		SeedFile:  "a/a.go",
		Question:  "how does A call B?",
		MaxTokens: 0,
	})
	if resp.Error != "" {
		t.Fatalf("SelectContext error: %s", resp.Error)
	}
	if resp.TotalTokens <= 0 {
		t.Fatalf("TotalTokens = %d, want > 0", resp.TotalTokens)
	}
}

func TestSelectContextIncludeTests(t *testing.T) {
	context.ClearIndexCache()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Abs: %v", err)
	}
	seed := "internal/context/context_test.go"
	if _, err := os.Stat(filepath.Join(root, seed)); err != nil {
		t.Skip("distilly module root not available")
	}

	without := SelectContext(SelectContextRequest{
		RepoRoot: root,
		SeedFile: seed,
		Question: "context tests",
	})
	if without.Error != "" {
		t.Fatalf("without tests: %s", without.Error)
	}

	withTests := SelectContext(SelectContextRequest{
		RepoRoot:     root,
		SeedFile:     seed,
		Question:     "context tests",
		IncludeTests: true,
	})
	if withTests.Error != "" {
		t.Fatalf("with tests: %s", withTests.Error)
	}

	if len(withTests.Files) <= len(without.Files) {
		t.Fatalf("IncludeTests should add files: without=%d with=%d", len(without.Files), len(withTests.Files))
	}
}
