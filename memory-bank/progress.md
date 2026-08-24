# Progress

Last updated: 2026-08-24

## Done

### Milestones 1–3 (CLI / engine)

- [x] Tokenizer, section splitter, cost tables, unified diff
- [x] Exact + near-duplicate detection (lines and example blocks)
- [x] History length flagger
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

### Milestone 5

- [x] Tree-sitter code-context selection (`internal/context`)
- [x] `distilly-context` CLI
- [x] Desktop Context workspace
- [x] Proxy `@distilly:context` marker injection

## In progress / gaps

- [ ] Manual Lint UI does not call `LogRequest` (dashboard mainly proxy-fed)
- [ ] Remove leftover `Greet` scaffold binding
- [ ] Wire regression suite into CI (no CI yet)
- [ ] True semantic paraphrase detection (embeddings)

## Held back (not on main)

- [ ] **`milestone-3/local-model-backend`** — optional local-model history
      compression (`internal/llm`, `lint.Optimize`, proxy wiring). Exists on
      branch only; **do not merge without explicit user verification** (see
      `activeContext.md` merge policy).

## Deferred

- In-process llama.cpp / GGUF loading (use llama-server HTTP instead)
- Native Anthropic Messages API
- Streaming through the proxy
