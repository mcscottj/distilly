# Roadmap

## Guiding principle

Ship the deterministic, rule-based linter first. It's fast, testable, and
gives users immediate value. AI-powered semantic compression comes later
as an optional second pass, compared against the rule-based baseline.

## Milestone 1 — Prompt Linter (CLI, no AI)

Goal: `distilly-lint some_prompt.txt` prints a report like:

```
Input Tokens: 14,220

Repeated Sections
------------------
System Prompt   2,300
Examples        5,100
History         4,800
Question          220

Suggestions
-----------
✓ Remove duplicate instructions
✓ Compress history
✓ Remove unused examples

Potential Savings: 46%
```

Tasks:
- [x] internal/tokenizer: wrap a tiktoken-compatible library
- [x] internal/lint: section splitter (system / examples / history / question)
- [x] internal/dedupe: exact + near-duplicate instruction detection
- [x] internal/cost: per-model $/token tables + estimator
- [x] internal/diff: unified diff renderer for before/after
- [x] cmd/lint: CLI wiring, flags, output formatting
- [x] testdata/prompts: a handful of real-world-style prompts for regression tests

## Milestone 2 — Regression harness

Before adding any AI-assisted rewriting, build a test suite of
(original prompt, constraints that must survive optimization) pairs.
This guards against over-optimization — e.g. collapsing "Always answer
in JSON. Do not include markdown. Do not explain." down to "Return JSON."
which silently drops constraints.

Tasks:
- [x] internal/lint: `Apply` — deterministic optimizer that actually
      produces the optimized prompt text, currently limited to the
      high-confidence exact-duplicate tier (near-duplicates and any
      future semantic rewrites require review, per Milestone 3's
      confidence-tier design)
- [x] internal/regression: harness of (prompt, constraints-that-must-survive)
      pairs, run against `Apply`
- [x] regression cases covering: exact-duplicate collapse, near-duplicates
      left untouched, history/question preserved verbatim, textually
      distinct phrasing of the same instruction left untouched
- [ ] wire the regression suite into CI once one exists

## Milestone 3 — Semantic compression (optional AI pass)

- ~~Local model backends: Ollama, llama.cpp, local GGUF~~ (deferred — not
  being built right now)
- Rule-based fallback for users without a GPU
- Every optimization gets a confidence score
  - High confidence (exact duplicates) -> auto-applied
  - Low confidence (semantic rewrites) -> requires user approval
- Example deduplication via clustering
- Automatic JSON conversion for structured data

Tasks:
- [x] internal/lint: `ApplyOptions.ApproveNearDuplicates` — near-duplicate
      lines (low confidence) collapse only when explicitly approved;
      exact duplicates (high confidence) keep auto-applying as before
- [x] cmd/lint: `-fix` prints the optimized prompt; `-approve-near-duplicates`
      opts in to the low-confidence tier
- [x] internal/lint: example deduplication via clustering — `SplitExamples`
      extracts whole few-shot example blocks, `dedupe.FindExact`/`FindNear`
      (at `DefaultExampleNearThreshold`) cluster redundant ones, and `Apply`
      collapses exact whole-block duplicates automatically / near-duplicate
      blocks once approved. Along the way, fixed a real bug: line-level
      exact-duplicate detection used to reach into example blocks and could
      corrupt a legitimately-different example that happened to share one
      identical line with another (see `nonExampleLines` in
      `internal/lint/examples.go`)
- [ ] automatic JSON conversion for structured data

## Milestone 4 — Desktop app

- Wails-based desktop app wrapping the same Go engine
- Acts as a local proxy between apps and OpenAI/Claude/etc.
- React + Vite + Tailwind frontend
- Prompt scoring UI, side-by-side diff, cost estimator dashboard

## Milestone 5 — Code context optimizer

- Tree-sitter based dependency analysis
- Given a question about a file, select only relevant files from a repo
  instead of sending the whole thing
