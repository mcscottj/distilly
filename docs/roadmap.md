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
- [ ] internal/diff: unified diff renderer for before/after
- [ ] cmd/lint: CLI wiring, flags, output formatting
- [ ] testdata/prompts: a handful of real-world-style prompts for regression tests

## Milestone 2 — Regression harness

Before adding any AI-assisted rewriting, build a test suite of
(original prompt, constraints that must survive optimization) pairs.
This guards against over-optimization — e.g. collapsing "Always answer
in JSON. Do not include markdown. Do not explain." down to "Return JSON."
which silently drops constraints.

## Milestone 3 — Semantic compression (optional AI pass)

- Local model backends: Ollama, llama.cpp, local GGUF
- Rule-based fallback for users without a GPU
- Every optimization gets a confidence score
  - High confidence (exact duplicates) -> auto-applied
  - Low confidence (semantic rewrites) -> requires user approval
- Example deduplication via clustering
- Automatic JSON conversion for structured data

## Milestone 4 — Desktop app

- Wails-based desktop app wrapping the same Go engine
- Acts as a local proxy between apps and OpenAI/Claude/etc.
- React + Vite + Tailwind frontend
- Prompt scoring UI, side-by-side diff, cost estimator dashboard

## Milestone 5 — Code context optimizer

- Tree-sitter based dependency analysis
- Given a question about a file, select only relevant files from a repo
  instead of sending the whole thing
