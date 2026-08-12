# Active context

Last updated: 2026-08-12

## Current focus

Milestone 4 complete: proxy lifecycle wired into the desktop app
(`StartProxy` / `StopProxy` / `GetProxyStatus` + Settings UI).

## Next engineering priorities

1. Manual Lint UI → `LogRequest` for Dashboard parity
2. Remove leftover `Greet` scaffold binding
3. GitHub Actions for `go test ./...` + regression
4. History compression (beyond flagging)
5. Milestone 5: Tree-sitter code-context file selection

## Recent decisions

- Shared canvases need Pro **plus team membership**; solo Pro → use repo docs
  instead of Publish
- Keep feature branches after merge unless user asks to delete
- No AI rewrite in v1; proxy is OpenAI-compatible only for M4
- Proxy binds `127.0.0.1` using `proxy_port` from SQLite; Settings saves before
  start so the port matches the UI

## Open questions / watchouts

- Should manual Analyze also `LogRequest` for Dashboard parity?
- When to add GitHub Actions for `go test ./...` + regression?
