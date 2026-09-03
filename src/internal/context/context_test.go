package context

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func fixtureRepo(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "repos", name)
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return abs
}

func TestParseGoModModuleAndReplace(t *testing.T) {
	root := fixtureRepo(t, "mini")
	mod, err := ParseGoMod(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("ParseGoMod: %v", err)
	}
	if mod.Path != "example.com/mini" {
		t.Fatalf("module path: got %q want %q", mod.Path, "example.com/mini")
	}
	if got, ok := mod.Replaces["example.com/other"]; !ok || got != "../other" {
		t.Fatalf("replace example.com/other: got %q ok=%v want ../other", got, ok)
	}
	if got, ok := mod.Replaces["golang.org/x/sys"]; !ok || got != "github.com/golang/sys" {
		t.Fatalf("replace golang.org/x/sys: got %q ok=%v want github.com/golang/sys", got, ok)
	}
}

func TestBuildIndexWalksGoFilesAndSkipsDirs(t *testing.T) {
	root := fixtureRepo(t, "mini")
	idx, err := BuildIndex(root)
	if err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	paths := make([]string, 0, len(idx.Files))
	for p := range idx.Files {
		paths = append(paths, filepath.ToSlash(p))
	}
	slices.Sort(paths)

	want := []string{"a/a.go", "b/b.go", "c/c.go", "noise/noise.go"}
	if !slices.Equal(paths, want) {
		t.Fatalf("indexed paths:\n got %v\nwant %v", paths, want)
	}

	for _, skip := range []string{
		"vendor/fake/fake.go",
		"build/out/generated.go",
		"node_modules/pkg/index.go",
		".git/objects/dummy.go",
	} {
		if _, ok := idx.Files[filepath.FromSlash(skip)]; ok {
			t.Errorf("expected %s to be skipped", skip)
		}
	}

	if idx.Module.Path != "example.com/mini" {
		t.Fatalf("index module path: got %q", idx.Module.Path)
	}
}

func TestExtractGoFilePackageImportsSymbols(t *testing.T) {
	root := fixtureRepo(t, "mini")
	src, err := os.ReadFile(filepath.Join(root, "a", "a.go"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	info, err := ExtractGoFile("a/a.go", src)
	if err != nil {
		t.Fatalf("ExtractGoFile: %v", err)
	}
	if info.Package != "a" {
		t.Fatalf("package: got %q want a", info.Package)
	}
	if !slices.Contains(info.Imports, "example.com/mini/b") {
		t.Fatalf("imports missing example.com/mini/b: %v", info.Imports)
	}
	for _, sym := range []string{"CallB", "Helper", "Version", "Shared"} {
		if !slices.Contains(info.Symbols, sym) {
			t.Errorf("symbols missing %q: %v", sym, info.Symbols)
		}
	}
}

func TestResolveImportPathToFiles(t *testing.T) {
	root := fixtureRepo(t, "mini")
	idx, err := BuildIndex(root)
	if err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	files, err := idx.ResolveImport("example.com/mini/b")
	if err != nil {
		t.Fatalf("ResolveImport: %v", err)
	}
	got := make([]string, len(files))
	for i, f := range files {
		got[i] = filepath.ToSlash(f)
	}
	if !slices.IsSorted(got) {
		t.Fatalf("ResolveImport result not sorted: %v", got)
	}
	want := []string{"b/b.go"}
	if !slices.Equal(got, want) {
		t.Fatalf("ResolveImport files: got %v want %v", got, want)
	}

	files, err = idx.ResolveImport("example.com/mini/noise")
	if err != nil {
		t.Fatalf("ResolveImport noise: %v", err)
	}
	got = make([]string, len(files))
	for i, f := range files {
		got[i] = filepath.ToSlash(f)
	}
	if !slices.IsSorted(got) {
		t.Fatalf("ResolveImport noise not sorted: %v", got)
	}
	want = []string{"noise/noise.go"}
	if !slices.Equal(got, want) {
		t.Fatalf("ResolveImport noise: got %v want %v", got, want)
	}
}

func TestBuildIndexDistillySnippet(t *testing.T) {
	root := fixtureRepo(t, "distilly-snippet")
	idx, err := BuildIndex(root)
	if err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}
	if idx.Module.Path != "distilly" {
		t.Fatalf("module: got %q want distilly", idx.Module.Path)
	}

	files, err := idx.ResolveImport("distilly/internal/lint")
	if err != nil {
		t.Fatalf("ResolveImport lint: %v", err)
	}
	found := false
	for _, f := range files {
		if filepath.ToSlash(f) == "internal/lint/lint.go" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected internal/lint/lint.go in %v", files)
	}

	lintInfo, ok := idx.Files[filepath.FromSlash("internal/lint/lint.go")]
	if !ok {
		t.Fatal("missing internal/lint/lint.go in index")
	}
	if !slices.Contains(lintInfo.Imports, "distilly/internal/store") {
		t.Fatalf("lint imports: %v", lintInfo.Imports)
	}
}

func TestBuildIndexSkipsUnparseableWithWarning(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/bad\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ok.go"), []byte("package ok\n\nfunc Fine() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Truncated / garbage that tree-sitter may still parse, but make it empty package-less.
	if err := os.WriteFile(filepath.Join(dir, "broken.go"), []byte("this is not go {{{{\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	idx, err := BuildIndex(dir)
	if err != nil {
		t.Fatalf("BuildIndex should not fail whole index: %v", err)
	}
	if _, ok := idx.Files["ok.go"]; !ok {
		t.Fatal("expected ok.go indexed")
	}
	if len(idx.Warnings) == 0 {
		t.Fatal("expected warning for unparseable file")
	}
	joined := strings.Join(idx.Warnings, "\n")
	if !strings.Contains(joined, "broken.go") {
		t.Fatalf("warnings should mention broken.go: %v", idx.Warnings)
	}
}

func TestSelectDefaultsAndSamePackage(t *testing.T) {
	ClearIndexCache()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/pkg\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "seed.go"), []byte("package pkg\n\nfunc Seed() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sibling.go"), []byte("package pkg\n\nfunc Sibling() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "seed_test.go"), []byte("package pkg\n\nfunc TestSeed(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Select(Options{RepoRoot: dir, SeedFile: "seed.go", Question: "seed"})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	paths := selectedPaths(result)
	if !slices.Contains(paths, "seed.go") || !slices.Contains(paths, "sibling.go") {
		t.Fatalf("expected seed + same-package sibling; got %v", paths)
	}
	if slices.Contains(paths, "seed_test.go") {
		t.Fatalf("IncludeTests=false should exclude *_test.go; got %v", paths)
	}

	result, err = Select(Options{RepoRoot: dir, SeedFile: "seed.go", Question: "seed", IncludeTests: true})
	if err != nil {
		t.Fatalf("Select IncludeTests: %v", err)
	}
	if !slices.Contains(selectedPaths(result), "seed_test.go") {
		t.Fatalf("IncludeTests=true should include seed_test.go; got %v", selectedPaths(result))
	}
}
