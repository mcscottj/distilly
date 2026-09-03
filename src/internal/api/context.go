package api

import (
	"distilly/internal/context"
)

const DefaultContextMaxTokens = 32000

// SelectContext runs context.Select and maps the result into SelectContextResponse.
func SelectContext(req SelectContextRequest) SelectContextResponse {
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = DefaultContextMaxTokens
	}

	result, err := context.Select(context.Options{
		RepoRoot:     req.RepoRoot,
		SeedFile:     req.SeedFile,
		Question:     req.Question,
		MaxDepth:     req.MaxDepth,
		MaxFiles:     req.MaxFiles,
		MaxTokens:    maxTokens,
		IncludeTests: req.IncludeTests,
	})
	if err != nil {
		return SelectContextResponse{Error: err.Error()}
	}

	md, err := context.FormatContext(result, req.RepoRoot)
	if err != nil {
		return SelectContextResponse{Error: err.Error()}
	}

	return SelectContextResponse{
		Files:         mapSelectedFiles(result.Files),
		TotalTokens:   result.TotalTokens,
		ExcludedCount: result.ExcludedCount,
		Warnings:      result.Warnings,
		Markdown:      md,
	}
}

func mapSelectedFiles(files []context.SelectedFile) []SelectedFileDTO {
	if len(files) == 0 {
		return nil
	}
	out := make([]SelectedFileDTO, len(files))
	for i, f := range files {
		reasons := make([]InclusionReasonDTO, len(f.Reasons))
		for j, r := range f.Reasons {
			reasons[j] = InclusionReasonDTO{Kind: r.Kind, Detail: r.Detail}
		}
		out[i] = SelectedFileDTO{
			Path:    f.Path,
			Tokens:  f.Tokens,
			Reasons: reasons,
		}
	}
	return out
}
