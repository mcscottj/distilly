# Code context test fixtures

Fixtures live in [`src/testdata/repos/`](../src/testdata/repos/). They exercise the context engine (`Select` / `FormatContext`), the `distilly-context` CLI, desktop Context workspace, and proxy `@distilly:context` marker handling.

Run from `src/`:

```bash
go test ./internal/context/... -v
go run ./cmd/context -repo testdata/repos/mini -seed a/a.go -question "how does A call B?"
go run ./cmd/context -repo testdata/repos/mini -seed a/a.go -format json
go run ./cmd/context -repo testdata/repos/mini -seed a/a.go -format markdown
```

## `mini/`

A tiny three-package chain plus noise:

| Package | Path | Imports |
|---------|------|---------|
| `mini/a` | `a/a.go` | `mini/b` |
| `mini/b` | `b/b.go` | `mini/c` |
| `mini/c` | `c/c.go` | (stdlib only) |
| `mini/noise` | `noise/noise.go` | unrelated |

Also includes skipped paths for index tests: `vendor/fake/`, `build/out/`.

**Regression expectations** (seed `a/a.go`, question mentions A→B):

- **Must include:** `a/a.go`, `b/b.go` (seed + direct import)
- **Must exclude:** `noise/noise.go`
- **Budget test:** `-max-tokens 50` keeps high/medium tiers, drops low-priority files with `ExcludedCount > 0`

## `distilly-snippet/`

Subset mirroring real Distilly layout:

| Path | Role |
|------|------|
| `internal/lint/lint.go` | Lint engine stub |
| `internal/store/store.go` | Store stub |
| `cmd/lint/main.go` | CLI entry importing lint |

**Dogfood-style check:** seed `internal/lint/lint.go` — co-package lint files included; unrelated packages excluded unless imported.

## Suggested manual pass

1. **CLI report** — `mini` seed `a/a.go`, verify table lists `a` + `b`, excludes `noise`.
2. **CLI markdown** — same inputs; output contains `### a/a.go` and `### b/b.go` fenced blocks.
3. **Desktop** — Context workspace with repo root pointing at `testdata/repos/mini`; copy markdown → Open in Lint.
4. **Proxy** — enable code context in Settings, embed marker in system message:

   ```
   @distilly:context
   seed=a/a.go
   ```

   Forward through proxy; system message should contain selected files, not the marker.

## Confidence tiers (selection)

| Tier | Rule | Budget behavior |
|------|------|-----------------|
| High | Seed file, same-package siblings | Always included |
| Medium | Direct imports, reverse importers (1 hop) | Always included |
| Low | Transitive imports (depth > 1), symbol name matches | Included until file/token budget hit |

See [`architecture.md`](architecture.md) for the full data flow.
