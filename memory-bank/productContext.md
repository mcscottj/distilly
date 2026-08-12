# Product context

## Why it exists

LLM apps burn tokens on repeated system instructions, redundant few-shot
examples, and long conversation histories. Distilly finds that waste and
optionally removes it — without calling another model to “rewrite” the prompt.

## How users interact

| Surface | Job |
|---------|-----|
| **Lint workspace** | Paste a prompt, pick a model, see score/sections/suggestions, apply with opt-in toggles, inspect diff |
| **CLI** | `distilly-lint` on a prompt file; `-fix` + approval flags for rewrite |
| **Proxy** | Point the OpenAI SDK at Distilly; requests are optimized then forwarded upstream |
| **Dashboard** | Aggregate tokens/$ saved and recent requests (mainly from proxy logs) |
| **Settings** | Upstream URL, API key, proxy port, approval toggles, passthrough |

## Core product idea

Prompts are split into sections (`System` / `Examples` / `History` /
`Question`), checked for waste, optionally rewritten, and scored — whether
pasted as text or sent as chat `messages`.

## User trust model

- Exact duplicates: safe to auto-apply
- Near-duplicates and JSON conversion: require explicit approval (format/meaning risk)
- Regression harness guards `Apply` against dropping must-survive constraints
