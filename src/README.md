# distilly

A local-first prompt linter and optimizer for developers who use LLM APIs
heavily. Distilly analyzes prompts and conversation history, flags waste
(duplicate instructions, repeated examples, bloated history), estimates
API cost, and suggests concrete token reductions — without changing the
meaning of your prompt.

Think "ESLint for prompts."

## Status

v1 (deterministic linter) and the regression harness are done. Now building
out Milestone 3's confidence-scored optimizations — see `docs/roadmap.md`.
Still no AI calls anywhere; everything is rule-based.

## Project layout

```
cmd/lint/          CLI entrypoint (distilly-lint)
internal/tokenizer/ Token counting (tiktoken-compatible)
internal/lint/      Core lint engine: orchestrates checks, produces a report
internal/dedupe/    Duplicate instruction / repeated example detection
internal/history/   Conversation history compression
internal/cost/      Token -> $ cost estimation per model
internal/diff/      Side-by-side before/after diff rendering
internal/regression/ (Prompt, constraints-that-must-survive) pairs guarding the optimizer
frontend/           React + Vite + Tailwind UI (used later by the Wails desktop app)
testdata/prompts/   Sample prompts used for regression testing heuristics
docs/               Design notes, roadmap
```

## v1 goal: Prompt Linter (no AI required)

- [x] Token counter
- [x] Duplicate instruction detector
- [x] Repeated example detector
- [x] History length flagger
- [x] Cost estimator
- [x] Before/after diff view
- [x] CLI: `distilly-lint <file>` prints a lint report

## Later: v2+

- Semantic compression via local models (Ollama / llama.cpp / GGUF) — deferred for now
- Code context optimizer (relevant-file selection for repo-aware prompts)
- Desktop app (Wails) acting as a local proxy between apps and OpenAI/Claude/etc.

## Tech stack

- Backend: Go, SQLite, tiktoken-compatible tokenizer, Tree-sitter (code analysis)
- Frontend: React, Vite, Tailwind CSS
- Desktop: Wails (Go + React)

## Getting started

```bash
# example.txt has semantically-similar-but-textually-distinct lines
# (e.g. "Use markdown." / "Respond in markdown." / "Format as markdown.").
# v1's exact-match dedupe correctly reports 0% savings on it — catching
# these requires near-duplicate detection (Milestone 3).
go run ./cmd/lint testdata/prompts/example.txt

# exact_duplicates.txt has real byte-identical repeated instructions,
# the kind that show up when a prompt template re-includes the system
# instructions before every few-shot example. v1 catches these.
go run ./cmd/lint testdata/prompts/exact_duplicates.txt

# -fix prints the optimized prompt instead of the report. By default it
# only touches the high-confidence tier (exact duplicate lines/examples).
go run ./cmd/lint -fix testdata/prompts/exact_duplicates.txt

# Lower-confidence tiers (near-duplicates, JSON conversion) need explicit
# approval — review the report's diff first, then opt in:
go run ./cmd/lint -fix -approve-near-duplicates -approve-json-conversion testdata/prompts/near_duplicates.txt
```

Run the test suite with:

```bash
go test ./...
```
