#!/usr/bin/env bash
# Auto-bump version, then build the desktop app (release, no +dev).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
"$ROOT/scripts/bump-version.sh"
cd "$ROOT/src"
LDFLAGS='-X distilly/internal/version.release=1'
exec wails build -ldflags "$LDFLAGS"
