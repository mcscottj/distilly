# Publish Desktop GitHub Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `scripts/publish-desktop.sh` that bumps version, builds the macOS desktop app only, zips it, commits/tags/pushes, and creates a GitHub Release with that zip as the sole asset.

**Architecture:** Thin bash orchestrator over existing `bump-version.sh` and `build-desktop.sh`. Packaging uses macOS `ditto`. Git + `gh release create` handle distribution. Binaries stay under gitignored `src/build/bin`; only `VERSION` and `wails.json` are committed.

**Tech Stack:** bash, git, GitHub CLI (`gh`), Wails (`wails build` via existing script), macOS `ditto`

## Global Constraints

- Desktop only — do not call `build-cli.sh` or upload CLI binaries
- Version format remains `YYYYMMDD.N`; tag is `v<version>`
- Zip name: `Distilly-<version>-macos.zip` under `src/build/bin/`
- Commit message exactly: `chore(release): Distilly <version>`
- Release title: `Distilly <version>`; notes: `Desktop app for macOS.`
- No auto-rollback after commit/tag if push/release fails
- Branch: `feature/publish-desktop-release`
- Do not change `scripts/build-release.sh` behavior

## File map

| File | Responsibility |
|------|----------------|
| `scripts/publish-desktop.sh` | Preconditions, bump, build, zip, commit, tag, push, `gh release create` |
| `src/README.md` | Document how to publish the desktop app |

---

### Task 1: `scripts/publish-desktop.sh`

**Files:**
- Create: `scripts/publish-desktop.sh`

**Interfaces:**
- Consumes: `scripts/bump-version.sh`, `scripts/build-desktop.sh`, `src/internal/version/VERSION`, `src/wails.json`, `src/build/bin/distilly.app`
- Produces: `src/build/bin/Distilly-<version>-macos.zip`, git commit, tag `v<version>`, GitHub Release

- [ ] **Step 1: Create the script**

Create `scripts/publish-desktop.sh` with mode `0755` and this content:

```bash
#!/usr/bin/env bash
# Bump version, build desktop only, zip .app, commit/tag/push, GitHub Release.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

die() {
  echo "error: $*" >&2
  exit 1
}

command -v gh >/dev/null 2>&1 || die "gh is required (GitHub CLI)"
command -v wails >/dev/null 2>&1 || die "wails is required (often ~/go/bin/wails)"
command -v ditto >/dev/null 2>&1 || die "ditto is required (macOS)"
command -v git >/dev/null 2>&1 || die "git is required"

git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "not inside a git repository"
git remote get-url origin >/dev/null 2>&1 || die "git remote 'origin' is not configured"

if ! git diff --quiet || ! git diff --cached --quiet; then
  die "working tree is dirty; commit or stash before publishing"
fi

if [[ -n "$(git ls-files --others --exclude-standard)" ]]; then
  die "untracked files present; commit or stash before publishing"
fi

gh auth status >/dev/null 2>&1 || die "gh is not authenticated (run: gh auth login)"

"$ROOT/scripts/bump-version.sh"

VERSION="$(tr -d '[:space:]' < "$ROOT/src/internal/version/VERSION")"
[[ -n "$VERSION" ]] || die "VERSION file is empty after bump"
TAG="v${VERSION}"
ZIP_NAME="Distilly-${VERSION}-macos.zip"
APP_PATH="$ROOT/src/build/bin/distilly.app"
ZIP_PATH="$ROOT/src/build/bin/${ZIP_NAME}"

"$ROOT/scripts/build-desktop.sh"

[[ -d "$APP_PATH" ]] || die "expected app bundle missing: $APP_PATH"

rm -f "$ZIP_PATH"
ditto -c -k --keepParent "$APP_PATH" "$ZIP_PATH"
[[ -f "$ZIP_PATH" ]] || die "failed to create zip: $ZIP_PATH"

git add "$ROOT/src/internal/version/VERSION" "$ROOT/src/wails.json"
git commit -m "chore(release): Distilly ${VERSION}"

git tag "$TAG"

echo "Pushing branch and tag ${TAG}..."
git push origin HEAD
git push origin "$TAG"

echo "Creating GitHub Release ${TAG}..."
gh release create "$TAG" "$ZIP_PATH" \
  --title "Distilly ${VERSION}" \
  --notes "Desktop app for macOS."

echo "Published ${TAG} with asset ${ZIP_NAME}"
```

- [ ] **Step 2: Syntax-check the script**

Run: `bash -n scripts/publish-desktop.sh`

Expected: no output, exit 0

- [ ] **Step 3: Verify dirty-tree precondition fails**

Run from repo root (creates a temporary dirty file, then cleans up):

```bash
echo dirty > /tmp/distilly-publish-precondition-test.txt
cp /tmp/distilly-publish-precondition-test.txt ./_publish_precondition_test.txt
set +e
./scripts/publish-desktop.sh
status=$?
set -e
rm -f ./_publish_precondition_test.txt
test "$status" -ne 0
```

Expected: script exits non-zero with a message about dirty/untracked files; no bump commit created.

- [ ] **Step 4: Commit**

```bash
git add scripts/publish-desktop.sh
git commit -m "feat: add publish-desktop script for GitHub Releases"
```

---

### Task 2: Document publish path in `src/README.md`

**Files:**
- Modify: `src/README.md` (Desktop / Version sections around lines 54–73)

**Interfaces:**
- Consumes: Task 1 script path `./scripts/publish-desktop.sh`

- [ ] **Step 1: Update Desktop and Version docs**

In the Desktop section, after the `build-release.sh` block, add this prose and command block:

````markdown
To bump, build desktop only, and publish a macOS GitHub Release (zip of
`distilly.app`):

```bash
./scripts/publish-desktop.sh
```
````

In the Version section, after the sentence about `build-release.sh`, add:

```text
Ship the desktop app to GitHub Releases with `./scripts/publish-desktop.sh`.
```

Keep existing wording for CLI/`build-release.sh` behavior unchanged.

- [ ] **Step 2: Skim the rendered section**

Confirm `src/README.md` still documents:

1. `wails dev` (no bump)
2. `./scripts/build-release.sh` (desktop + CLIs)
3. `./scripts/publish-desktop.sh` (desktop GitHub Release)

- [ ] **Step 3: Commit**

```bash
git add src/README.md
git commit -m "docs: document publish-desktop GitHub Release script"
```

---

### Task 3: End-to-end publish (manual / operator)

**Files:** none (uses Task 1 script against real `origin`)

**Interfaces:**
- Consumes: clean working tree on `feature/publish-desktop-release` (or `main` after merge), authenticated `gh`, `wails`

- [ ] **Step 1: Preconditions**

Confirm:

```bash
git status          # clean
gh auth status      # logged in
command -v wails
```

- [ ] **Step 2: Run publish (creates a real release)**

```bash
./scripts/publish-desktop.sh
```

Expected:

1. New `YYYYMMDD.N` in `src/internal/version/VERSION` and matching `wails.json`
2. Commit `chore(release): Distilly <version>`
3. Tag `v<version>` pushed
4. GitHub Release with sole asset `Distilly-<version>-macos.zip`
5. No CLI binaries attached to the release

- [ ] **Step 3: Verify the zip**

```bash
VERSION=$(tr -d '[:space:]' < src/internal/version/VERSION)
unzip -l "src/build/bin/Distilly-${VERSION}-macos.zip" | head
```

Expected: listing includes `distilly.app/` (bundle contents).

- [ ] **Step 4: Confirm release on GitHub**

```bash
gh release view "v${VERSION}"
```

Expected: one asset, the macOS zip; title `Distilly <version>`.

Do **not** amend or delete the release as part of this task unless the operator explicitly asks to clean up a test publish.

---

## Spec coverage checklist

| Spec requirement | Task |
|------------------|------|
| New `publish-desktop.sh` | 1 |
| Bump via existing bump script | 1 |
| Desktop-only build | 1 |
| `ditto` zip under `src/build/bin/` | 1 |
| Commit VERSION + wails.json only | 1 |
| Tag `v<version>`, push, `gh release create` | 1 |
| Dirty-tree / tool preconditions | 1 |
| No auto-rollback | 1 (by omission of rollback logic) |
| `src/README.md` pointer | 2 |
| Success criteria / live verify | 3 |
| Spec listed in `docs/specs/README.md` | Already done with the design commit |
