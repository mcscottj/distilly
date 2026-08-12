# distilly (src)

Go module and Wails project root. Prefer the [root README](../README.md) for
product status; this file covers working inside `src/`.

## Status

M1–M3 CLI/engine work is done. M4 desktop + store + proxy package are in
place; proxy lifecycle is not yet bound on `app.go`. See
[`../docs/architecture.md`](../docs/architecture.md) and
[`../memory-bank/progress.md`](../memory-bank/progress.md).

## Layout

```
cmd/lint/              CLI entrypoint (distilly-lint)
internal/lint/         Core engine: Run + Apply + score + examples + jsonify
internal/api/          Wails Analyze/Apply DTOs
internal/proxy/        OpenAI-compatible proxy (unwired to App lifecycle)
internal/store/        SQLite requests + settings
internal/tokenizer/    Token counting
internal/dedupe/       Exact + near-duplicate detection
internal/history/      History length flagger
internal/cost/         Token → $ estimation
internal/diff/         Before/after diff
internal/regression/   Constraint-survival harness
frontend/              React Lint / Dashboard / Settings
testdata/prompts/      Regression fixtures
main.go, app.go        Wails entry + bindings
```

## CLI

```bash
go run ./cmd/lint testdata/prompts/exact_duplicates.txt
go run ./cmd/lint -fix testdata/prompts/exact_duplicates.txt
go run ./cmd/lint -fix -approve-near-duplicates -approve-json-conversion \
  testdata/prompts/near_duplicates.txt
go test ./...
```

## Desktop

```bash
wails doctor   # CLI often at ~/go/bin/wails
wails dev
wails build    # → build/bin/distilly.app
```

Bound on `App` today: `Analyze`, `Apply`, `DiffForDuplicate`, `ListModels`,
dashboard/settings helpers, `LogRequest`. Not bound: proxy start/stop.
