# Active context

Last updated: 2026-08-12

## Current focus

Documenting architecture for agents and humans:

- Cursor canvas (local): architecture overview beside chat
- Repo: `docs/architecture.md` + this `memory-bank/`
- README refresh to match M1–M4 reality

Branch: `docs/architecture-overview` (from `main`).

## Next engineering priority (after docs)

Wire proxy lifecycle into the desktop app:

1. `StartProxy` / `StopProxy` / status on `App`
2. Settings start/stop/status controls
3. Confirm proxy logs appear on Dashboard

## Recent decisions

- Shared canvases need Pro **plus team membership**; solo Pro → use repo docs
  instead of Publish
- Keep feature branches after merge unless user asks to delete
- No AI rewrite in v1; proxy is OpenAI-compatible only for M4

## Open questions / watchouts

- Should manual Analyze also `LogRequest` for Dashboard parity?
- When to add GitHub Actions for `go test ./...` + regression?
