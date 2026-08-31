# Publish desktop GitHub Release — Design

**Date:** 2026-08-31  
**Branch:** `feature/publish-desktop-release`  
**Status:** Approved

## Goal

Add a repeatable local script that bumps Distilly’s shared version, builds the
**desktop app only**, packages it as a macOS `.app` zip, commits the version
bump, pushes a git tag, and creates a **GitHub Release** whose only asset is
that zip.

## Non-goals

- Publishing CLI binaries (`distilly-lint`, `distilly-context`)
- GitHub Actions / CI-built releases
- `.dmg` packaging
- Apple code signing or notarization
- Windows / Linux desktop builds
- Changing the existing `YYYYMMDD.N` version scheme

## Background

Versioning already lives in `src/internal/version/VERSION` with scripts:

| Script | Role |
|--------|------|
| `scripts/bump-version.sh` | Bump `VERSION` + sync `wails.json` |
| `scripts/build-desktop.sh` | Wails release build (no bump) |
| `scripts/build-release.sh` | Bump once, then desktop + both CLIs |

The versioning design deferred GitHub release publishing. This work adds that
path for the desktop app only.

Built binaries under `src/build/bin` are gitignored; they must not be committed
to the repo. Distribution is via GitHub Release assets.

## Approach

**New script:** `scripts/publish-desktop.sh` (committed to the repo).

Keep `build-release.sh` as the local “build everything” path. Publishing is an
explicit, separate command so “ship desktop” is obvious and does not upload
CLIs by accident.

## Release asset

| Item | Value |
|------|--------|
| Source | `src/build/bin/distilly.app` (Wails output) |
| Package | Zip via `ditto -c -k --keepParent` (preserves `.app` bundle layout) |
| Zip path | `src/build/bin/Distilly-<version>-macos.zip` (under existing gitignore) |
| Release asset | That zip only |

Example: version `20260831.1` → `Distilly-20260831.1-macos.zip`.

## Git & GitHub

| Item | Value |
|------|--------|
| Commit files | `src/internal/version/VERSION`, `src/wails.json` only |
| Commit message | `chore(release): Distilly <version>` |
| Tag | `v<version>` (e.g. `v20260831.1`) |
| Push | `git push origin HEAD` then `git push origin v<version>` |
| Release | `gh release create` for that tag, published (not draft) |
| Title | `Distilly <version>` |
| Notes | Short fixed text: desktop app for macOS |

## Script flow

Preconditions (fail fast if unmet):

1. Inside the Distilly git repo with `origin` configured
2. `gh` installed and authenticated for the repo
3. Working tree clean before bump (no unrelated dirty files)
4. `wails` available (same requirement as `build-desktop.sh`)

Ordered steps:

1. Run `scripts/bump-version.sh`
2. Read the new version from `src/internal/version/VERSION`
3. Run `scripts/build-desktop.sh`
4. Create the zip next to the `.app` under `src/build/bin/`
5. Stage and commit only the version bump files
6. Create lightweight tag `v<version>`
7. Push branch and tag to `origin`
8. Create the GitHub Release and upload the zip

Use `set -euo pipefail`. Stop on first failure.

**No automatic rollback:** if push or `gh release create` fails after the local
commit/tag exist, leave them in place so the operator can fix auth/network and
retry or finish manually.

## Docs

Add a short pointer in `src/README.md` (Desktop / Version section) to
`./scripts/publish-desktop.sh` for shipping the macOS desktop app to GitHub
Releases.

Optionally list this spec in `docs/specs/README.md`.

## Success criteria

1. Running `./scripts/publish-desktop.sh` on a clean tree produces a new
   `YYYYMMDD.N`, a desktop release build, a committed bump, a pushed
   `vYYYYMMDD.N` tag, and a GitHub Release whose sole asset is
   `Distilly-<version>-macos.zip`.
2. The zip extracts to a runnable `distilly.app` on macOS.
3. CLIs are not built or uploaded by this script.
4. Binary/zip artifacts remain untracked (under `src/build/bin`).
5. `build-release.sh` behavior is unchanged.
