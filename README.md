![Distilly logo](docs/distilly-logo.png)

# distilly

A local-first prompt linter and optimizer for developers who use LLM APIs
heavily. Distilly analyzes prompts and conversation history, flags waste
(duplicate instructions, repeated examples, bloated history), estimates
API cost, and suggests concrete token reductions — without changing the
meaning of your prompt.

Think "ESLint for prompts."

## Status

Early scaffolding. Building the v1 CLI linter first (see `docs/roadmap.md`).
No AI-powered features yet — v1 is fully deterministic and rule-based.

## Project layout

```
cmd/lint/          CLI entrypoint (distilly-lint)
internal/tokenizer/ Token counting (tiktoken-compatible)
internal/lint/      Core lint engine: orchestrates checks, produces a report
internal/dedupe/    Duplicate instruction / repeated example detection
internal/history/   Conversation history compression
internal/cost/      Token -> $ cost estimation per model
internal/diff/      Side-by-side before/after diff rendering
frontend/           React + Vite + Tailwind UI (used later by the Wails desktop app)
testdata/prompts/   Sample prompts used for regression testing heuristics
docs/               Design notes, roadmap
```

## v1 goal: Prompt Linter (no AI required)

- [x] Token counter
- [ ] Duplicate instruction detector
- [ ] Repeated example detector
- [x] History length flagger
- [x] Cost estimator
- [x] Before/after diff view
- [ ] CLI: `distilly-lint <file>` prints a lint report

## Later: v2+

- Semantic compression via local models (Ollama / llama.cpp / GGUF)
- Rule-based fallback for users without a GPU
- Code context optimizer (relevant-file selection for repo-aware prompts)
- Automatic JSON conversion for structured data
- Confidence-scored optimizations (auto-apply high confidence, prompt for review on low confidence)
- Desktop app (Wails) acting as a local proxy between apps and OpenAI/Claude/etc.

## Tech stack

- Backend: Go, SQLite, tiktoken-compatible tokenizer, Tree-sitter (code analysis)
- Frontend: React, Vite, Tailwind CSS
- Desktop: Wails (Go + React)

## Getting started

```bash
go run ./cmd/lint testdata/prompts/example.txt
```
