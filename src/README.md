# distilly (src)

Go module and Wails project root. Prefer the [root README](../README.md) for
product status; this file covers working inside `src/`.

## Status

M1–M3 CLI/engine work is done. M4 desktop + store + proxy package and
lifecycle bindings are in place. M5 code-context optimizer (Tree-sitter
Select, CLI, desktop Context workspace, proxy marker) is in place. See
[`../docs/architecture.md`](../docs/architecture.md) and
[`../memory-bank/progress.md`](../memory-bank/progress.md).

## Layout

```
cmd/lint/              CLI entrypoint (distilly-lint)
cmd/context/           CLI entrypoint (distilly-context)
internal/lint/         Core engine: Run + Apply + score + examples + jsonify
internal/context/      Code context: Select + FormatContext (Tree-sitter)
internal/api/          Wails Analyze/Apply/SelectContext DTOs
internal/proxy/        OpenAI-compatible proxy (Start/Stop bound on App)
internal/store/        SQLite requests + settings
internal/tokenizer/    Token counting
internal/dedupe/       Exact + near-duplicate detection
internal/history/      History length flagger
internal/cost/         Token → $ estimation
internal/diff/         Before/after diff
internal/regression/   Constraint-survival harness
internal/version/      YYYYMMDD.N + bump helpers
frontend/              React Lint / Context / Dashboard / Settings
testdata/prompts/      Regression fixtures
testdata/repos/        Code-context fixture modules
main.go, app.go        Wails entry + bindings
```

## CLI

### Install (run from anywhere)

From the repo root:

```bash
go install -C src -o "$(go env GOPATH)/bin/distilly-lint" ./cmd/lint
go install -C src -o "$(go env GOPATH)/bin/distilly-context" ./cmd/context
```

Add `$(go env GOPATH)/bin` to your `PATH` if needed. Re-run after CLI changes.

### Development (`go run` from `src/`)

```bash
go run ./cmd/lint -version
go run ./cmd/lint testdata/prompts/exact_duplicates.txt
go run ./cmd/lint -fix testdata/prompts/exact_duplicates.txt
go run ./cmd/lint -fix -approve-near-duplicates -approve-json-conversion \
  testdata/prompts/near_duplicates.txt
go run ./cmd/context -repo . -seed internal/lint/apply.go -question "Apply options"
go test ./...
```

## Desktop

```bash
wails doctor   # CLI often at ~/go/bin/wails
wails dev      # About shows VERSION+dev; does not bump
```

For a versioned release build from the repo root:

```bash
./scripts/build-desktop.sh   # bumps YYYYMMDD.N, then wails build
```

**Distilly → About Distilly** shows `Version <string>`.

## Version

Source of truth: `internal/version/VERSION`. CLIs: `-version`. Release builds
use `./scripts/build-cli.sh` / `./scripts/build-desktop.sh` (auto-bump + no
`+dev` suffix).

Bound on `App` today: `Analyze`, `Apply`, `SelectContext`, `DiffForDuplicate`, `ListModels`,
dashboard/settings helpers, `LogRequest`, `StartProxy`, `StopProxy`,
`GetProxyStatus`.
