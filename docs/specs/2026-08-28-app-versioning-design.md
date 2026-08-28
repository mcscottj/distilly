# App versioning — Design

**Date:** 2026-08-28  
**Branch:** `feature/app-versioning`

## Goal

Give Distilly a single, auto-bumped version string shared by the desktop app and CLIs. Every intentional release rebuild gets a new version. Local/dev runs show the same base version with a `+dev` suffix. The desktop app surfaces the version in the native **Distilly → About Distilly** menu; CLIs expose `-version`.

## Non-goals

- Semver (`0.1.0`) or git-hash-based versioning
- GitHub Actions / automated release publishing
- Custom in-app About React page (native Wails/macOS About is enough)
- Auto-bumping on every `go test`, `go run`, or `wails dev`

## Version format

```
YYYYMMDD.N
```

Examples: `20260828.1`, `20260828.2`, `20260829.1`.

- **Date:** local calendar date when the bump runs (`YYYYMMDD`).
- **N:** integer build counter for that day, starting at `1`.

### Bump rules

On each release-oriented build (via project scripts):

1. Read `src/internal/version/VERSION` (single line, no quotes).
2. If the stored date equals today → set `N` to `N+1`.
3. If the stored date differs from today (or file missing/invalid) → write `YYYYMMDD.1`.
4. Write the new value back, sync `wails.json` `info.productVersion`, then build.

### Dev vs release display

| Context | Display |
|---------|---------|
| Release build (build scripts + ldflags) | Exact file contents, e.g. `20260828.1` |
| Dev (`wails dev`, `go run`, tests) | Same base + `+dev`, e.g. `20260828.1+dev` |

**Mechanism:** package var `release` defaults empty. Release scripts pass:

```text
-ldflags "-X distilly/internal/version.release=1"
```

When `release == "1"`, `String()` returns `Base()` only; otherwise `Base()+"+dev"`.

## Architecture

### Single source of truth

- **File:** `src/internal/version/VERSION` — committed, one line, format `YYYYMMDD.N`.
- **Package:** `src/internal/version` — `//go:embed VERSION`.
- **Exports:** `Base()` (trimmed embed), `String()` (dev/release), `NextVersion(today, current string) string` (pure bump logic).

Desktop app and both CLIs import this package only; no duplicated version literals.

### Bump + build

- `NextVersion` lives in `internal/version` and is unit-tested.
- `go run ./internal/version/cmd/bump` (from `src/`) writes the new VERSION and updates `wails.json` `info.productName` / `info.productVersion`.
- Repo-root scripts:
  - `scripts/bump-version.sh` — runs the bump command.
  - `scripts/build-desktop.sh` — bump, then `wails build` from `src/` with release ldflags.
  - `scripts/build-cli.sh` — bump once, then `go build` both CLIs with release ldflags.

README / `src/README.md`: use these scripts when a rebuild should mint a new version; do not bump on casual `wails dev`.

### Desktop: About Distilly

In `src/main.go`, configure Wails:

```go
Mac: &mac.Options{
    About: &mac.AboutInfo{
        Title:   "Distilly",
        Message: "Version " + version.String(),
        Icon:    icon, // embed app icon if present under build/
    },
},
```

`wails.json` `info.productVersion` stays in sync via bump so Info.plist / Finder Get Info match About.

### CLIs: `-version`

- `distilly-lint -version` and `distilly-context -version` print `version.String()` to stdout and exit `0`.
- No positional args required when `-version` is set.
- Usage / help text mentions `-version`.

## Error handling

- Invalid or empty VERSION: bump treats as missing and writes `YYYYMMDD.1`.
- Empty embed at runtime: `Base()` returns `"0"` → `"0+dev"` in dev (should not happen once the file is committed).
- Bump is safe to run twice in a row (second run increments `N`).

## Testing

- Unit tests for `NextVersion`: new day, same-day increment, empty/invalid input.
- CLI: `-version` exits 0 and prints a non-empty `YYYYMMDD.N` or `YYYYMMDD.N+dev`.
- No automated click-test of the native About panel.

## Docs

- Root README / `src/README.md`: version file path, build scripts, `-version`, About menu.
- User-guide CLI section: one-liner for `-version`.

## Success criteria

1. After a release build script, `src/internal/version/VERSION` is a new `YYYYMMDD.N`.
2. **Distilly → About Distilly** shows that version (without `+dev` for release builds).
3. `distilly-lint -version` and `distilly-context -version` print the same scheme.
4. `wails dev` / `go run` show `…+dev` and do not auto-bump.
