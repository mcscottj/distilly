#!/usr/bin/env bash
# Send a non-streaming chat completion through the Distilly local proxy.
#
# Prerequisites:
#   1. Distilly desktop app running
#   2. Settings → Upstream: API key (and base URL, default OpenAI)
#   3. Settings → Local proxy → Start proxy
#
# Then open Dashboard (or click Refresh) to see a `proxy` row.
#
# Usage:
#   ./scripts/test-proxy.sh
#   DISTILLY_PROXY_PORT=8787 ./scripts/test-proxy.sh

set -euo pipefail

PORT="${DISTILLY_PROXY_PORT:-8787}"
URL="http://127.0.0.1:${PORT}/v1/chat/completions"

echo "POST ${URL}" >&2

if ! curl -sS "$URL" \
  -H 'Content-Type: application/json' \
  -d '{"model":"gpt-4o","stream":false,"messages":[{"role":"user","content":"ping"}]}'; then
  echo >&2
  echo "Could not reach the Distilly proxy. Start it from Settings → Start proxy." >&2
  exit 1
fi

echo
echo "If that succeeded, open Dashboard (or Refresh) in Distilly." >&2
