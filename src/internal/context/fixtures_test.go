package context

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSelectMiniMustIncludeExclude(t *testing.T) {
	ClearIndexCache()
	root := fixtureRepo(t, "mini")

	result, err := Select(Options{
		RepoRoot: root,
		SeedFile: "a/a.go",
		Question: "how does A call B?",
	})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}

	paths := selectedPaths(result)
	for _, want := range []string{"a/a.go", "b/b.go"} {
		if !slices.Contains(paths, want) {
			t.Errorf("mustInclude missing %s; got %v", want, paths)
		}
	}
	if slices.Contains(paths, "noise/noise.go") {
		t.Errorf("mustExclude noise/noise.go; got %v", paths)
	}
}

func TestSelectBudgetKeepsHighPriority(t *testing.T) {
	ClearIndexCache()
	root := fixtureRepo(t, "mini")

	result, err := Select(Options{
		RepoRoot:  root,
		SeedFile:  "a/a.go",
		Question:  "how does A call B?",
		MaxTokens: 50, // force truncation of low-priority files
	})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if result.ExcludedCount <= 0 {
		t.Fatalf("expected ExcludedCount > 0 with tight MaxTokens, got %d (files=%v)", result.ExcludedCount, selectedPaths(result))
	}
	paths := selectedPaths(result)
	if !slices.Contains(paths, "a/a.go") {
		t.Fatalf("seed a/a.go must be kept under budget; got %v", paths)
	}
	// High/medium always included: seed and direct import b.
	if !slices.Contains(paths, "b/b.go") {
		t.Fatalf("medium-priority b/b.go must be kept; got %v", paths)
	}
}

func TestSelectDeterministic(t *testing.T) {
	ClearIndexCache()
	root := fixtureRepo(t, "mini")
	opts := Options{
		RepoRoot: root,
		SeedFile: "a/a.go",
		Question: "how does A call B?",
	}

	a, err := Select(opts)
	if err != nil {
		t.Fatalf("Select a: %v", err)
	}
	b, err := Select(opts)
	if err != nil {
		t.Fatalf("Select b: %v", err)
	}
	if !slices.Equal(selectedPaths(a), selectedPaths(b)) {
		t.Fatalf("non-deterministic:\n first %v\nsecond %v", selectedPaths(a), selectedPaths(b))
	}
}

func TestFormatContextMarkdown(t *testing.T) {
	ClearIndexCache()
	root := fixtureRepo(t, "mini")

	result, err := Select(Options{
		RepoRoot: root,
		SeedFile: "a/a.go",
		Question: "how does A call B?",
		MaxFiles: 2,
	})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if len(result.Files) == 0 {
		t.Fatal("expected selected files")
	}

	md, err := FormatContext(result, root)
	if err != nil {
		t.Fatalf("FormatContext: %v", err)
	}
	for _, f := range result.Files {
		slash := filepath.ToSlash(f.Path)
		if !strings.Contains(md, slash) {
			t.Errorf("markdown missing path %s", slash)
		}
	}
	if !strings.Contains(md, "```go") {
		t.Error("markdown missing ```go fence")
	}
	if !strings.Contains(md, "package a") {
		t.Error("markdown missing file contents")
	}
}

func TestDogfoodDistillySnippet(t *testing.T) {
	ClearIndexCache()
	root := fixtureRepo(t, "distilly-snippet")

	// Seed cmd entrypoint: lint is direct import (medium), store is transitive (low).
	// MaxDepth=1 keeps store out while still including the lint package.
	result, err := Select(Options{
		RepoRoot: root,
		SeedFile: "cmd/lint/main.go",
		Question: "how does the lint command run?",
		MaxDepth: 1,
	})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}

	paths := selectedPaths(result)
	if !slices.Contains(paths, "cmd/lint/main.go") {
		t.Errorf("mustInclude cmd/lint/main.go; got %v", paths)
	}
	if !slices.Contains(paths, "internal/lint/lint.go") {
		t.Errorf("mustInclude co-package/import internal/lint/lint.go; got %v", paths)
	}
	if slices.Contains(paths, "internal/store/store.go") {
		t.Errorf("mustExclude store at MaxDepth=1; got %v", paths)
	}
}

func selectedPaths(r Result) []string {
	out := make([]string, len(r.Files))
	for i, f := range r.Files {
		out[i] = filepath.ToSlash(f.Path)
	}
	return out
}
