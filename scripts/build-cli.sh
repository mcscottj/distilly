#!/usr/bin/env bash
# Build distilly-lint and distilly-context only (no version bump). Release (no +dev).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/src"
OUT="${DISTILLY_CLI_OUT:-$ROOT/src/build/bin}"
mkdir -p "$OUT"
LDFLAGS='-X distilly/internal/version.release=1'
go build -ldflags "$LDFLAGS" -o "$OUT/distilly-lint" ./cmd/lint
go build -ldflags "$LDFLAGS" -o "$OUT/distilly-context" ./cmd/context
echo "Built $OUT/distilly-lint and $OUT/distilly-context"
"$OUT/distilly-lint" -version
