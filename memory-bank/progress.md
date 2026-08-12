# Progress

Last updated: 2026-08-12

## Done

### Milestones 1–3 (CLI / engine)

- [x] Tokenizer, section splitter, cost tables, unified diff
- [x] Exact + near-duplicate detection (lines and example blocks)
- [x] History length flagger (no compression yet)
- [x] `Run` / `Apply` with confidence-tier options
- [x] Structured data → JSON conversion (opt-in)
- [x] Prompt scoring
- [x] CLI `distilly-lint` (`-model`, `-fix`, approval flags)
- [x] Regression harness (`internal/regression` + `testdata/prompts`)

### Milestone 4

- [x] Wails scaffold + React Lint / Dashboard / Settings
- [x] `internal/api` DTOs; App bindings: Analyze, Apply, DiffForDuplicate,
      ListModels, dashboard, settings, LogRequest
- [x] SQLite store (requests + settings)
- [x] OpenAI-compatible proxy package + unit tests (httptest upstream,
      stream rejection, savings logging)
- [x] Wire `StartProxy` / `StopProxy` / `GetProxyStatus` on `app.go` and
      Settings UI start/stop controls

## In progress / gaps

- [ ] Manual Lint UI does not call `LogRequest` (dashboard mainly proxy-fed)
- [ ] Remove leftover `Greet` scaffold binding
- [ ] Wire regression suite into CI (no CI yet)
- [ ] History compression (flag only today)
- [ ] Milestone 5: Tree-sitter code-context file selection

## Deferred

- Local model backends (Ollama / llama.cpp / GGUF) and GPU fallbacks
- Native Anthropic Messages API
- Streaming through the proxy
