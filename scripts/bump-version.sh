#!/usr/bin/env bash
# Bump Distilly VERSION (YYYYMMDD.N) and sync wails.json productVersion.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/src"
go run ./internal/version/cmd/bump
