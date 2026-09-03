#!/usr/bin/env bash
# Build desktop app only (no version bump). Uses current VERSION; release (no +dev).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/src"
LDFLAGS='-X distilly/internal/version.release=1'
exec wails build -ldflags "$LDFLAGS"
