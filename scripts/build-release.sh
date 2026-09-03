#!/usr/bin/env bash
# Bump VERSION once, then build desktop + both CLIs with the same version.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
"$ROOT/scripts/bump-version.sh"
"$ROOT/scripts/build-desktop.sh"
"$ROOT/scripts/build-cli.sh"
