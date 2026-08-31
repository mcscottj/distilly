#!/usr/bin/env bash
# Bump version, build desktop only, zip .app, commit/tag/push, GitHub Release.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

die() {
  echo "error: $*" >&2
  exit 1
}

command -v git >/dev/null 2>&1 || die "git is required"

git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "not inside a git repository"
git remote get-url origin >/dev/null 2>&1 || die "git remote 'origin' is not configured"

if ! git diff --quiet || ! git diff --cached --quiet; then
  die "working tree is dirty; commit or stash before publishing"
fi

if [[ -n "$(git ls-files --others --exclude-standard)" ]]; then
  die "untracked files present; commit or stash before publishing"
fi

command -v gh >/dev/null 2>&1 || die "gh is required (GitHub CLI)"
command -v wails >/dev/null 2>&1 || die "wails is required (often ~/go/bin/wails)"
command -v ditto >/dev/null 2>&1 || die "ditto is required (macOS)"

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
