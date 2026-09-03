<p align="center">
  <img src="docs/distilly-logo.png" alt="Distilly logo" width="512">
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"></a>
  <a href="https://wails.io/"><img src="https://img.shields.io/badge/Wails-v2-00ADD8?style=flat" alt="Wails"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue?style=flat" alt="Apache 2.0"></a>
</p>

# distilly

A local-first prompt linter and optimizer for developers who use LLM APIs
heavily. Distilly analyzes prompts and conversation history, flags waste
(duplicate instructions, repeated examples, bloated history), estimates
API cost, and suggests concrete token reductions — without changing the
meaning of your prompt. Run a prompt through it and see how you score —
a surprisingly honest way to test your prompting skills.

Think "ESLint for prompts."

## Features

- Token counter + section splitter (System / Examples / History / Question)
- Exact + near-duplicate detection (lines and few-shot example blocks)
- History length flagger, cost estimator, prompt score, before/after diff
- CLI: `distilly-lint` report and `-fix` with approval flags
- Regression harness (`go test ./...` under `src/`)
- Desktop app: Lint workspace, Context workspace, Dashboard, Settings
- Local OpenAI-compatible proxy (non-streaming) — start/stop from Settings
- Code context: Tree-sitter selection, `distilly-context` CLI, proxy
  `@distilly:context` marker injection

Optimizations are **rule-based Go** today — there is no AI rewrite backend.
See [`docs/roadmap.md`](docs/roadmap.md) for planned work and
[`docs/architecture.md`](docs/architecture.md) for how the pieces fit.

## Docs

- [`docs/user-guide.md`](docs/user-guide.md) — how to use the CLI, desktop app, and proxy
- [`docs/architecture.md`](docs/architecture.md) — engine, proxy, and package layout
- [`docs/roadmap.md`](docs/roadmap.md) — shipped scope and planned work
- [`docs/licensing.md`](docs/licensing.md) — Apache 2.0 and open-core model
- [`docs/prompt-fixtures.md`](docs/prompt-fixtures.md) /
  [`docs/code-context-fixtures.md`](docs/code-context-fixtures.md) — test fixtures

## Project layout

```
docs/                      User guide, architecture, roadmap, design history
scripts/                   Release builds, version bump, proxy smoke test
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

For Cursor (or similar) agents, prefer a **project rule or skill that invokes
these CLIs** (or the local proxy) rather than asking the model to compress
prompts itself — see
[Typical workflows → Wire Distilly into an AI agent](docs/user-guide.md#5-wire-distilly-into-an-ai-agent-rule-or-skill).

### Versioned release builds

Bump once and build desktop + both CLIs with the same `YYYYMMDD.N`:

```bash
./scripts/build-release.sh
```

Or step by step (no second bump between builds):

```bash
./scripts/bump-version.sh      # increment VERSION + wails.json
./scripts/build-desktop.sh     # wails build (no bump)
./scripts/build-cli.sh         # both CLIs (no bump)
```

Version lives in `src/internal/version/VERSION`. Dev runs (`wails dev`, `go run`)
show `…+dev` and do not bump. Desktop: **Distilly → About Distilly** shows the
same version string.

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

To exercise the local proxy end-to-end (not the ChatGPT app — that has no custom
base URL), configure an upstream API key and base URL in **Settings**, start the
proxy, then from the repo root:

```bash
./scripts/test-proxy.sh
```

That POSTs `stream: false` to `http://127.0.0.1:8787/v1/chat/completions`. Open
**Dashboard** (or click Refresh) to see a `proxy` row. Override the port with
`DISTILLY_PROXY_PORT` if you changed it in Settings.

## License

Licensed under the [Apache License, Version 2.0](LICENSE). See
[`docs/licensing.md`](docs/licensing.md) for the open-core model and trademark
notes.
