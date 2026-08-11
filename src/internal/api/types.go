// Package api exposes JSON-serializable DTOs and thin adapters over the
// lint engine for Wails UI bindings (and other callers that need stable
// wire types instead of internal lint/dedupe structs).
package api

// AnalyzeRequest is the input to Analyze.
type AnalyzeRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

// AnalyzeResponse mirrors lint.Report as stable, serializable fields for
// the desktop UI, plus Score/Issues from lint.Score.
type AnalyzeResponse struct {
	InputTokens           int                   `json:"inputTokens"`
	Sections              SectionBreakdown      `json:"sections"`
	Duplicates            []DuplicateGroup      `json:"duplicates"`
	NearDuplicates        []DuplicateGroup      `json:"nearDuplicates"`
	DuplicateExamples     []DuplicateGroup      `json:"duplicateExamples"`
	NearDuplicateExamples []DuplicateGroup      `json:"nearDuplicateExamples"`
	StructuredData        []StructuredDataBlock `json:"structuredData"`
	Suggestions           []string              `json:"suggestions"`
	PotentialSavings      float64               `json:"potentialSavings"`
	Model                 string                `json:"model"`
	CostKnown             bool                  `json:"costKnown"`
	EstimatedCostUSD      float64               `json:"estimatedCostUsd"`
	EstimatedSavingsUSD   float64               `json:"estimatedSavingsUsd"`

	// Score is 0–100 prompt quality; Issues lists human-readable
	// deductions from lint.Score.
	Score  int      `json:"score"`
	Issues []string `json:"issues"`
}

// ApplyRequest is the input to Apply.
type ApplyRequest struct {
	Prompt                string `json:"prompt"`
	ApproveNearDuplicates bool   `json:"approveNearDuplicates"`
	ApproveJSONConversion bool   `json:"approveJsonConversion"`
}

// ApplyResponse is the optimized prompt plus a structured full-prompt diff.
type ApplyResponse struct {
	Optimized string     `json:"optimized"`
	FullDiff  []DiffLine `json:"fullDiff"`
}

// DiffLine is one line of a before/after diff (wraps internal/diff.Line).
type DiffLine struct {
	Marker  string `json:"marker"`  // " ", "-", or "+"
	Content string `json:"content"`
}

// SectionBreakdown is per-section token counts.
type SectionBreakdown struct {
	System   int `json:"system"`
	Examples int `json:"examples"`
	History  int `json:"history"`
	Question int `json:"question"`
}

// DuplicateGroup is a serializable duplicate/near-duplicate cluster with
// an optional per-group structured diff (collapse Lines → Keep).
type DuplicateGroup struct {
	Lines      []string   `json:"lines"`
	Keep       string     `json:"keep"`
	Confidence float64    `json:"confidence"`
	Diff       []DiffLine `json:"diff"`
}

// StructuredDataBlock is a serializable structured-data opportunity with
// suggested JSON and a before/after diff.
type StructuredDataBlock struct {
	Keys   []string   `json:"keys"`
	Values []string   `json:"values"`
	Raw    []string   `json:"raw"`
	JSON   string     `json:"json"`
	Diff   []DiffLine `json:"diff"`
}
