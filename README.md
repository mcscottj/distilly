<img src="docs/distilly-logo.png" alt="Distilly logo" width="512">

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

**Milestone 4 is done**: Wails desktop app (Lint workspace, Dashboard,
Settings), SQLite persistence, OpenAI-compatible proxy package, and
`StartProxy` / `StopProxy` / status wired into the App and Settings UI.

**Milestone 5 is done**: Tree-sitter code-context selection (`internal/context`),
`distilly-context` CLI, desktop Context workspace, and proxy `@distilly:context`
marker injection.

There is still **no AI rewrite backend** — everything is rule-based Go.
No CI yet.

See [`docs/roadmap.md`](docs/roadmap.md), [`docs/architecture.md`](docs/architecture.md),
[`docs/user-guide.md`](docs/user-guide.md), [`docs/prompt-fixtures.md`](docs/prompt-fixtures.md),
[`docs/code-context-fixtures.md`](docs/code-context-fixtures.md),
and [`memory-bank/`](memory-bank/) for detail.

## Project layout

```
docs/                      Roadmap + architecture overview
memory-bank/               Agent working memory (progress, patterns, context)
test-proxy.sh              Curl a non-streaming chat completion through the local proxy
src/                       Go module + Wails project root
  main.go, app.go          Desktop entry + UI bindings
  cmd/lint/                CLI (distilly-lint)
  cmd/context/             CLI (distilly-context)
  internal/lint/           Core engine: Run (report) + Apply (optimize)
  internal/context/        Code context: Select + FormatContext
  internal/api/            JSON DTOs for Analyze / Apply / SelectContext
  internal/proxy/          OpenAI-compatible /v1/chat/completions gateway
  internal/store/          SQLite settings + request metrics
  internal/tokenizer/      Token counting (tiktoken-compatible)
  internal/dedupe/         Exact + near-duplicate detection
  internal/history/        History length flagger
  internal/cost/           Token → $ estimates per model
  internal/diff/           Before/after diff rendering
  internal/regression/     Constraint-survival harness for Apply
  frontend/                React + Vite + Tailwind (Lint / Context / Dashboard / Settings)
  testdata/prompts/        Regression fixtures
  testdata/repos/          Code-context fixture modules
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
- [x] Proxy package (non-streaming; tested) — start/stop from Settings
- [x] Code context: CLI, desktop Context workspace, proxy `@distilly:context` marker

## Later

- History compression (beyond flagging)
- CI for the regression suite
- Optional semantic compression via local models (deferred)

## Tech stack

- Backend: Go, SQLite, tiktoken-compatible tokenizer
- Frontend: React, Vite, Tailwind CSS
- Desktop: Wails (Go + React)

## Getting started

### Install the CLI (optional)

To run `distilly-lint` and `distilly-context` from any directory, install once
from the repo root (Go 1.20+):

```bash
go install -C src -o "$(go env GOPATH)/bin/distilly-lint" ./cmd/lint
go install -C src -o "$(go env GOPATH)/bin/distilly-context" ./cmd/context
```

Ensure `$(go env GOPATH)/bin` (usually `~/go/bin`) is on your `PATH` — e.g. add
`export PATH="$(go env GOPATH)/bin:$PATH"` to `~/.zshrc`. Re-run the install
commands after pulling CLI changes.

Then, from anywhere:

```bash
distilly-lint /path/to/prompt.txt
distilly-context -repo /path/to/repo -seed internal/lint/apply.go -question "..."
```

The `-o` names are required; a plain `go install ./cmd/lint` would produce
`lint`, not `distilly-lint`. See [`docs/user-guide.md`](docs/user-guide.md#cli)
for flags and examples.

`distilly-lint -version` / `distilly-context -version` print the build version
(`YYYYMMDD.N`, or `…+dev` for local `go run` / `go install` without release
ldflags).

### Versioned release builds

Each intentional rebuild that should mint a new version:

```bash
./scripts/build-desktop.sh   # bumps VERSION, then wails build (no +dev)
./scripts/build-cli.sh       # bumps VERSION, then builds both CLIs
./scripts/bump-version.sh    # bump only
```

Version lives in `src/internal/version/VERSION` (`YYYYMMDD.N`). Dev runs
(`wails dev`, `go run`) show `…+dev` and do not bump. Desktop: **Distilly →
About Distilly** shows the same version string.

### Development commands

Commands below run from `src/` (Go module root):

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

To exercise the local proxy (not the ChatGPT app — that has no custom base URL),
start Distilly, **Settings → Start proxy**, then from the repo root:

```bash
./test-proxy.sh
```

That POSTs `stream: false` to `http://127.0.0.1:8787/v1/chat/completions`. Open
**Dashboard** (or click Refresh) to see a `proxy` row. Override the port with
`DISTILLY_PROXY_PORT` if you changed it in Settings.
