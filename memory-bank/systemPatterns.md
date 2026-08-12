# System patterns

## Shared engine

All entry points call `internal/lint`:

- `Run(prompt, model)` → report (tokens, issues, savings, score)
- `Apply(prompt, options)` → optimized text

Desktop and proxy also use `internal/store` (SQLite).

## Message ↔ prompt bridge (proxy)

Chat `messages` map to a sectioned prompt for linting, then map back to
messages before upstream forward (unless passthrough). Streaming requests are
rejected.

## Confidence tiers

| Tier | Behavior |
|------|----------|
| Exact duplicate lines / example blocks | Auto-applied |
| Near-duplicates | Only with `ApproveNearDuplicates` |
| Structured prose → JSON | Only with `ApproveJSONConversion` |

## Persistence

SQLite at user config `…/distilly/distilly.db`:

- `settings` — upstream, API key, proxy port, toggles
- `requests` — logged savings (source e.g. `proxy`)

## Package roles

| Package | Role |
|---------|------|
| `internal/lint` | Orchestration, scoring, examples, jsonify |
| `internal/api` | Wails JSON DTOs |
| `internal/proxy` | `/v1/chat/completions` gateway |
| `internal/store` | SQLite |
| `tokenizer` / `dedupe` / `history` / `cost` / `diff` | Helpers |
| `internal/regression` | Constraint-survival harness on `Apply` |

## Desktop binding pattern

`src/app.go` exposes methods to React via Wails. Store opens on startup /
closes on shutdown. Proxy **package** is separate; lifecycle is not yet bound
on `App`.
