// Package context selects a minimal relevant Go file set for LLM prompts.
package context

// Options configures context selection (Select).
type Options struct {
	RepoRoot     string
	SeedFile     string // repo-relative path
	Question     string
	MaxDepth     int  // default 2 — transitive import hops
	MaxFiles     int  // default 20
	MaxTokens    int  // default 32000; 0 = no cap
	IncludeTests bool // default false
}

// InclusionReason explains why a file was selected.
type InclusionReason struct {
	Kind   string // "seed" | "same_package" | "import" | "reverse_import" | "symbol_match"
	Detail string
}

// SelectedFile is one file included in a Select result.
type SelectedFile struct {
	Path    string
	Tokens  int
	Reasons []InclusionReason
}

// Result is the output of Select.
type Result struct {
	Files         []SelectedFile
	TotalTokens   int
	ExcludedCount int
	Warnings      []string
}
