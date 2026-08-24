# Active context

Last updated: 2026-08-24

## Current focus

Milestone 5 (code context) is merged to `main`: Tree-sitter file selection,
`distilly-context` CLI, desktop Context workspace, proxy `@distilly:context`
marker.

## Merge policy — do not skip

**`milestone-3/local-model-backend` must NOT be merged to `main` without
explicit user verification**, even if the user asks to "merge everything" or
similar. That branch is intentionally held back pending review. When asked to
merge unmerged branches, list it separately and confirm before merging.

## Next engineering priorities

1. Manual Lint UI → `LogRequest` for Dashboard parity
2. Remove leftover `Greet` scaffold binding
3. GitHub Actions for `go test ./...` + regression
4. Semantic paraphrase detection (embeddings) — still out of scope
5. Revisit `milestone-3/local-model-backend` when user is ready to verify

## Recent decisions

- Shared canvases need Pro **plus team membership**; solo Pro → use repo docs
  instead of Publish
- Keep feature branches after merge unless user asks to delete
- Proxy binds `127.0.0.1` using `proxy_port` from SQLite; Settings saves before
  start so the port matches the UI
- M5 merged to `main`; local-model backend branch remains unmerged by choice
  (2026-08-24)

## Open questions / watchouts

- Should manual Analyze also `LogRequest` for Dashboard parity?
- When to add GitHub Actions for `go test ./...` + regression?
- When to review and merge `milestone-3/local-model-backend`?
