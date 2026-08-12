![Distilly logo](docs/distilly-logo.png)

# distilly

A local-first prompt linter and optimizer for developers who use LLM APIs
heavily. Distilly analyzes prompts and conversation history, flags waste
(duplicate instructions, repeated examples, bloated history), estimates
API cost, and suggests concrete token reductions — without changing the
meaning of your prompt.

Think "ESLint for prompts."

## Status

Milestones **1–3 are done**: deterministic CLI linter, confidence-tier
`Apply` (exact auto-apply; near-duplicates and JSON conversion opt-in),
prompt scoring, and a regression harness that guards optimizations against
dropping constraints.

**Milestone 4 is mostly built**: Wails desktop app (Lint workspace,
Dashboard, Settings), SQLite persistence, and a unit-tested
OpenAI-compatible proxy package. The remaining M4 gap is wiring
`StartProxy` / `StopProxy` into the App and Settings UI so the proxy is
startable from the desktop app.

There is still **no AI rewrite backend** — everything is rule-based Go.
No CI yet. Milestone 5 (Tree-sitter code-context selection) is not started.

See [`docs/roadmap.md`](docs/roadmap.md), [`docs/architecture.md`](docs/architecture.md),
and [`memory-bank/`](memory-bank/) for detail.

## Project layout

```
docs/                      Roadmap + architecture overview
memory-bank/               Agent working memory (progress, patterns, context)
src/                       Go module + Wails project root
  main.go, app.go          Desktop entry + UI bindings
  cmd/lint/                CLI (distilly-lint)
  internal/lint/           Core engine: Run (report) + Apply (optimize)
  internal/api/            JSON DTOs for Analyze / Apply
  internal/proxy/          OpenAI-compatible /v1/chat/completions gateway
  internal/store/          SQLite settings + request metrics
  internal/tokenizer/      Token counting (tiktoken-compatible)
  internal/dedupe/         Exact + near-duplicate detection
  internal/history/        History length flagger
  internal/cost/           Token → $ estimates per model
  internal/diff/           Before/after diff rendering
  internal/regression/     Constraint-survival harness for Apply
  frontend/                React + Vite + Tailwind (Lint / Dashboard / Settings)
  testdata/prompts/        Regression fixtures
```

## What works today

- [x] Token counter + section splitter (System / Examples / History / Question)
- [x] Exact + near-duplicate detection (lines and few-shot example blocks)
- [x] History length flagger
- [x] Cost estimator + prompt score
- [x] Before/after diff
- [x] CLI: `distilly-lint` report and `-fix` with approval flags
- [x] Regression harness (`go test ./...` under `src/`)
- [x] Desktop Lint workspace (analyze / apply / diff / toggles)
- [x] Dashboard + Settings (SQLite-backed)
- [x] Proxy package (non-streaming; tested) — **not yet startable from UI**

## Later

- Wire proxy start/stop into the desktop app (finish M4)
- History compression (beyond flagging)
- CI for the regression suite
- Code context optimizer (Milestone 5 / Tree-sitter)
- Optional semantic compression via local models (deferred)

## Tech stack

- Backend: Go, SQLite, tiktoken-compatible tokenizer
- Frontend: React, Vite, Tailwind CSS
- Desktop: Wails (Go + React)

## Getting started

Commands run from `src/` (Go module root):

```bash
# Lint a prompt file (report)
go run ./cmd/lint testdata/prompts/exact_duplicates.txt

# Optimize — exact duplicates auto-apply; lower-confidence tiers need flags
go run ./cmd/lint -fix testdata/prompts/exact_duplicates.txt
go run ./cmd/lint -fix -approve-near-duplicates -approve-json-conversion \
  testdata/prompts/near_duplicates.txt

# Tests
go test ./...

# Desktop app (Wails CLI typically at ~/go/bin/wails)
wails doctor
wails dev
```
