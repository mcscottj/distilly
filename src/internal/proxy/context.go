package proxy

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"distilly/internal/context"
	"distilly/internal/lint"
)

const contextMarker = "@distilly:context"

// Code context setting keys (must match frontend SettingKey / store values).
const (
	SettingRepoRoot           = "repo_root"
	SettingCodeContextEnabled = "enable_code_context"
	SettingContextMaxDepth    = "context_max_depth"
	SettingContextMaxTokens   = "context_max_tokens"
	DefaultContextMaxDepth    = 2
	DefaultContextMaxTokens   = 32000
)

type contextTrimConfig struct {
	RepoRoot  string
	MaxDepth  int
	MaxTokens int
}

// TrimCodeContext replaces an @distilly:context block in the prompt with
// selected file contents when a seed= directive is present.
func TrimCodeContext(prompt string, cfg contextTrimConfig) (string, error) {
	sections := lint.SplitSections(prompt)
	newSystem, seed, found := extractContextBlock(sections.System)
	if !found || seed == "" {
		return prompt, nil
	}

	result, err := context.Select(context.Options{
		RepoRoot:  cfg.RepoRoot,
		SeedFile:  seed,
		Question:  strings.TrimSpace(sections.Question),
		MaxDepth:  cfg.MaxDepth,
		MaxTokens: cfg.MaxTokens,
	})
	if err != nil {
		return prompt, fmt.Errorf("context.Select: %w", err)
	}

	formatted, err := context.FormatContext(result, cfg.RepoRoot)
	if err != nil {
		return prompt, fmt.Errorf("context.FormatContext: %w", err)
	}

	rebuilt := rebuildPrompt(newSystem, sections.Examples, sections.History, sections.Question, formatted)
	return rebuilt, nil
}

// extractContextBlock removes the @distilly:context directive block from system
// text and returns the remaining system text plus the seed path.
func extractContextBlock(system string) (rest string, seed string, found bool) {
	idx := strings.Index(system, contextMarker)
	if idx < 0 {
		return system, "", false
	}

	prefix := strings.TrimRight(system[:idx], "\n")
	suffix := system[idx:]
	lines := strings.Split(suffix, "\n")

	blockEnd := len(lines)
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			blockEnd = i
			break
		}
	}

	for _, line := range lines[:blockEnd] {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "seed=") {
			seed = strings.TrimSpace(strings.TrimPrefix(trimmed, "seed="))
		}
	}

	var afterParts []string
	if blockEnd+1 < len(lines) {
		afterParts = lines[blockEnd+1:]
	}
	after := strings.TrimLeft(strings.Join(afterParts, "\n"), "\n")

	var b strings.Builder
	if prefix != "" {
		b.WriteString(prefix)
	}
	if after != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(after)
	}
	return b.String(), seed, true
}

func rebuildPrompt(system, examples, history, question, insertion string) string {
	if insertion != "" {
		if strings.TrimSpace(system) != "" {
			system = strings.TrimRight(system, "\n") + "\n\n" + insertion
		} else {
			system = insertion
		}
	}

	var b strings.Builder
	if strings.TrimSpace(system) != "" {
		b.WriteString("System:\n")
		b.WriteString(strings.TrimRight(system, "\n"))
		b.WriteByte('\n')
	}
	if strings.TrimSpace(examples) != "" {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("Examples:\n")
		b.WriteString(strings.TrimRight(examples, "\n"))
		b.WriteByte('\n')
	}
	if strings.TrimSpace(history) != "" {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("History:\n")
		b.WriteString(strings.TrimRight(history, "\n"))
		b.WriteByte('\n')
	}
	if strings.TrimSpace(question) != "" {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("Question: ")
		b.WriteString(strings.TrimSpace(question))
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func parseContextIntSetting(v string, defaultVal int) int {
	v = strings.TrimSpace(v)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}

func applyCodeContext(prompt string, cfg runtimeConfig) string {
	if !cfg.CodeContextEnabled || strings.TrimSpace(cfg.RepoRoot) == "" {
		return prompt
	}
	trimmed, err := TrimCodeContext(prompt, contextTrimConfig{
		RepoRoot:  cfg.RepoRoot,
		MaxDepth:  cfg.ContextMaxDepth,
		MaxTokens: cfg.ContextMaxTokens,
	})
	if err != nil {
		log.Printf("distilly proxy: code context trim failed: %v", err)
		return prompt
	}
	return trimmed
}
