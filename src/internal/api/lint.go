package api

import (
	"sort"
	"strings"

	"distilly/internal/cost"
	"distilly/internal/dedupe"
	"distilly/internal/diff"
	"distilly/internal/lint"
)

// ListModels returns the model names from cost.Table, sorted alphabetically.
func ListModels() []string {
	models := make([]string, 0, len(cost.Table))
	for name := range cost.Table {
		models = append(models, name)
	}
	sort.Strings(models)
	return models
}

// Analyze runs the lint engine and maps the report into AnalyzeResponse.
func Analyze(req AnalyzeRequest) AnalyzeResponse {
	r := lint.Run(req.Prompt, req.Model)
	scored := lint.Score(r)
	return AnalyzeResponse{
		InputTokens:           r.InputTokens,
		Sections:              SectionBreakdown(r.Sections),
		Duplicates:            mapDuplicateGroups(r.Duplicates),
		NearDuplicates:        mapDuplicateGroups(r.NearDuplicates),
		DuplicateExamples:     mapDuplicateGroups(r.DuplicateExamples),
		NearDuplicateExamples: mapDuplicateGroups(r.NearDuplicateExamples),
		StructuredData:        mapStructuredData(r.StructuredData),
		Suggestions:           r.Suggestions,
		PotentialSavings:      r.PotentialSavings,
		Model:                 r.Model,
		CostKnown:             r.CostKnown,
		EstimatedCostUSD:      r.EstimatedCostUSD,
		EstimatedSavingsUSD:   r.EstimatedSavingsUSD,
		Score:                 scored.Score,
		Issues:                scored.Issues,
	}
}

// Apply runs lint.Apply and returns the optimized prompt with a
// structured line diff of the full before/after text.
func Apply(req ApplyRequest) ApplyResponse {
	optimized := lint.Apply(req.Prompt, lint.ApplyOptions{
		ApproveNearDuplicates: req.ApproveNearDuplicates,
		ApproveJSONConversion: req.ApproveJSONConversion,
	})
	before := strings.Split(req.Prompt, "\n")
	after := strings.Split(optimized, "\n")
	return ApplyResponse{
		Optimized: optimized,
		FullDiff:  mapDiffLines(diff.Lines(before, after)),
	}
}

// DiffForDuplicate returns a structured before/after diff for one
// duplicate group (all Lines collapsing to Keep).
func DiffForDuplicate(d DuplicateGroup) []DiffLine {
	return mapDiffLines(diff.Lines(d.Lines, []string{d.Keep}))
}

func mapDuplicateGroups(dupes []dedupe.Duplicate) []DuplicateGroup {
	if len(dupes) == 0 {
		return nil
	}
	out := make([]DuplicateGroup, len(dupes))
	for i, d := range dupes {
		out[i] = DuplicateGroup{
			Lines:      d.Lines,
			Keep:       d.Keep,
			Confidence: d.Confidence,
			Diff:       mapDiffLines(diff.Lines(d.Lines, []string{d.Keep})),
		}
	}
	return out
}

func mapStructuredData(blocks []lint.StructuredBlock) []StructuredDataBlock {
	if len(blocks) == 0 {
		return nil
	}
	out := make([]StructuredDataBlock, len(blocks))
	for i, b := range blocks {
		jsonLine := b.JSON()
		out[i] = StructuredDataBlock{
			Keys:   b.Keys,
			Values: b.Values,
			Raw:    b.Raw,
			JSON:   jsonLine,
			Diff:   mapDiffLines(diff.Lines(b.Raw, []string{jsonLine})),
		}
	}
	return out
}

func mapDiffLines(lines []diff.Line) []DiffLine {
	if len(lines) == 0 {
		return nil
	}
	out := make([]DiffLine, len(lines))
	for i, l := range lines {
		out[i] = DiffLine{Marker: l.Marker, Content: l.Content}
	}
	return out
}
