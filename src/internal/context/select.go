package context

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"distilly/internal/tokenizer"
)

const (
	defaultMaxDepth = 2
	defaultMaxFiles = 20

	tierHigh   = 0
	tierMedium = 1
	tierLow    = 2
)

type candidate struct {
	path    string
	tier    int
	reasons []InclusionReason
	tokens  int
}

// Select chooses relevant files for an LLM context window.
func Select(opts Options) (Result, error) {
	opts = applyDefaults(opts)
	if strings.TrimSpace(opts.RepoRoot) == "" {
		return Result{}, fmt.Errorf("context.Select: RepoRoot is required")
	}
	if strings.TrimSpace(opts.SeedFile) == "" {
		return Result{}, fmt.Errorf("context.Select: SeedFile is required")
	}

	idx, err := BuildIndex(opts.RepoRoot)
	if err != nil {
		return Result{}, err
	}

	seed := filepath.Clean(opts.SeedFile)
	if _, ok := idx.Files[seed]; !ok {
		// Also try slash-normalized lookup for callers using forward slashes on Windows.
		alt := filepath.FromSlash(filepath.ToSlash(seed))
		if _, ok := idx.Files[alt]; ok {
			seed = alt
		} else {
			return Result{}, fmt.Errorf("context.Select: seed file %q not in index", opts.SeedFile)
		}
	}

	graph, err := BuildGraph(idx)
	if err != nil {
		return Result{}, err
	}

	cands := map[string]*candidate{}
	addReason := func(path string, tier int, reason InclusionReason, allowTest bool) {
		if !opts.IncludeTests && !allowTest && isTestFile(path) {
			return
		}
		if _, ok := idx.Files[path]; !ok {
			return
		}
		c, ok := cands[path]
		if !ok {
			c = &candidate{path: path, tier: tier, reasons: nil}
			cands[path] = c
		}
		if tier < c.tier {
			c.tier = tier
		}
		for _, r := range c.reasons {
			if r.Kind == reason.Kind && r.Detail == reason.Detail {
				return
			}
		}
		c.reasons = append(c.reasons, reason)
	}

	// Seed is always eligible even if it is a *_test.go file.
	addReason(seed, tierHigh, InclusionReason{Kind: "seed", Detail: "seed file"}, true)

	siblings := samePackageSiblings(idx, seed)
	for _, sib := range siblings {
		addReason(sib, tierHigh, InclusionReason{Kind: "same_package", Detail: "same package as seed"}, false)
	}

	// Direct imports of seed (and of same-package siblings — over-include).
	highFiles := append([]string{seed}, siblings...)
	for _, hf := range highFiles {
		for _, impPath := range graph.Forward[hf] {
			detail := fmt.Sprintf("imported by %s", filepath.ToSlash(hf))
			addReason(impPath, tierMedium, InclusionReason{Kind: "import", Detail: detail}, false)
		}
	}

	// Reverse importers of seed (1 hop).
	for _, rev := range graph.Reverse[seed] {
		addReason(rev, tierMedium, InclusionReason{
			Kind:   "reverse_import",
			Detail: fmt.Sprintf("imports seed %s", filepath.ToSlash(seed)),
		}, false)
	}

	// Transitive imports via BFS (depth > 1 → low; depth 1 already medium).
	dist := graph.BFS(seed, opts.MaxDepth)
	for p, d := range dist {
		if d <= 1 {
			continue
		}
		addReason(p, tierLow, InclusionReason{
			Kind:   "import",
			Detail: fmt.Sprintf("transitive import depth %d", d),
		}, false)
	}

	// Symbol name matches against question tokens (low tier).
	qTokens := questionTokens(opts.Question)
	if len(qTokens) > 0 {
		symPaths := make([]string, 0, len(idx.Files))
		for p := range idx.Files {
			symPaths = append(symPaths, p)
		}
		slices.Sort(symPaths)
		for _, p := range symPaths {
			if p == seed {
				continue
			}
			fi := idx.Files[p]
			matched := matchingSymbols(fi.Symbols, qTokens)
			if len(matched) == 0 {
				continue
			}
			addReason(p, tierLow, InclusionReason{
				Kind:   "symbol_match",
				Detail: "symbols: " + strings.Join(matched, ", "),
			}, false)
		}
	}

	// Load contents + token counts (deterministic path order).
	paths := make([]string, 0, len(cands))
	for p := range cands {
		paths = append(paths, p)
	}
	slices.Sort(paths)

	var loadErrs []string
	for _, p := range paths {
		abs := filepath.Join(idx.RepoRoot, p)
		src, err := os.ReadFile(abs)
		if err != nil {
			loadErrs = append(loadErrs, fmt.Sprintf("read %s: %v", p, err))
			delete(cands, p)
			continue
		}
		content := string(src)
		cands[p].tokens = tokenizer.Count(content)
	}

	ordered := make([]*candidate, 0, len(cands))
	for _, p := range paths {
		if c, ok := cands[p]; ok {
			ordered = append(ordered, c)
		}
	}
	slices.SortFunc(ordered, cmpCandidate)

	selected, excluded := applyBudget(ordered, opts.MaxFiles, opts.MaxTokens)

	result := Result{
		Files:         make([]SelectedFile, 0, len(selected)),
		ExcludedCount: excluded,
		Warnings:      append(slices.Clone(idx.Warnings), loadErrs...),
	}
	slices.Sort(result.Warnings)
	for _, c := range selected {
		reasons := slices.Clone(c.reasons)
		slices.SortFunc(reasons, func(a, b InclusionReason) int {
			if a.Kind != b.Kind {
				return strings.Compare(a.Kind, b.Kind)
			}
			return strings.Compare(a.Detail, b.Detail)
		})
		result.Files = append(result.Files, SelectedFile{
			Path:    c.path,
			Tokens:  c.tokens,
			Reasons: reasons,
		})
		result.TotalTokens += c.tokens
	}
	return result, nil
}

// FormatContext renders a Select result as markdown code blocks labeled by path.
func FormatContext(result Result, repoRoot string) (string, error) {
	if strings.TrimSpace(repoRoot) == "" {
		return "", fmt.Errorf("context.FormatContext: repoRoot is required")
	}
	var b strings.Builder
	for i, f := range result.Files {
		abs := filepath.Join(repoRoot, f.Path)
		src, err := os.ReadFile(abs)
		if err != nil {
			return "", fmt.Errorf("context.FormatContext: read %s: %w", f.Path, err)
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		slash := filepath.ToSlash(f.Path)
		fmt.Fprintf(&b, "### %s\n\n```go\n%s", slash, src)
		if len(src) > 0 && src[len(src)-1] != '\n' {
			b.WriteByte('\n')
		}
		b.WriteString("```\n")
	}
	return b.String(), nil
}

func applyDefaults(opts Options) Options {
	if opts.MaxDepth == 0 {
		opts.MaxDepth = defaultMaxDepth
	}
	if opts.MaxFiles == 0 {
		opts.MaxFiles = defaultMaxFiles
	}
	// MaxTokens: 0 means no token cap (Options comment). The documented 32000
	// default is a caller convention — Select does not rewrite 0 → 32000.
	return opts
}

func isTestFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, "_test.go")
}

func cmpCandidate(a, b *candidate) int {
	if a.tier != b.tier {
		return a.tier - b.tier
	}
	return strings.Compare(filepath.ToSlash(a.path), filepath.ToSlash(b.path))
}

func applyBudget(ordered []*candidate, maxFiles, maxTokens int) (selected []*candidate, excluded int) {
	var totalTokens int
	for _, c := range ordered {
		// High and medium are always included (over-include over miss).
		if c.tier <= tierMedium {
			selected = append(selected, c)
			totalTokens += c.tokens
			continue
		}
		// Low tier: include until file or token budget would be exceeded.
		if maxFiles > 0 && len(selected) >= maxFiles {
			excluded++
			continue
		}
		if maxTokens > 0 && totalTokens+c.tokens > maxTokens {
			excluded++
			continue
		}
		selected = append(selected, c)
		totalTokens += c.tokens
	}
	return selected, excluded
}

var nonIdent = regexp.MustCompile(`[^A-Za-z0-9_]+`)

func questionTokens(q string) []string {
	parts := nonIdent.Split(strings.TrimSpace(q), -1)
	seen := make(map[string]bool)
	var out []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		// Skip very short / common English glue words.
		if len(p) < 2 || isStopWord(p) {
			continue
		}
		key := strings.ToLower(p)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
	}
	return out
}

func isStopWord(s string) bool {
	switch strings.ToLower(s) {
	case "how", "does", "the", "a", "an", "is", "are", "to", "of", "in", "on",
		"for", "and", "or", "with", "what", "why", "when", "where", "do", "call",
		"this", "that", "it", "be", "by", "from", "as", "at", "into", "about":
		return true
	default:
		return false
	}
}

func matchingSymbols(symbols, qTokens []string) []string {
	var matched []string
	for _, sym := range symbols {
		for _, tok := range qTokens {
			if symbolMatchesToken(sym, tok) {
				matched = append(matched, sym)
				break
			}
		}
	}
	slices.Sort(matched)
	return matched
}

func symbolMatchesToken(sym, tok string) bool {
	if strings.EqualFold(sym, tok) {
		return true
	}
	// Token appears as a CamelCase / snake segment of the symbol.
	symParts := splitIdent(sym)
	for _, part := range symParts {
		if strings.EqualFold(part, tok) {
			return true
		}
	}
	return false
}

func splitIdent(s string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	var cur strings.Builder
	runes := []rune(s)
	flush := func() {
		if cur.Len() > 0 {
			parts = append(parts, cur.String())
			cur.Reset()
		}
	}
	for i, r := range runes {
		if r == '_' {
			flush()
			continue
		}
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || nextLower {
				flush()
			}
		}
		cur.WriteRune(r)
	}
	flush()
	return parts
}
