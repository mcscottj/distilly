# Project brief

Distilly is a local-first prompt linter and optimizer — “ESLint for prompts.”
It analyzes LLM prompts and chat history, flags waste (duplicates, bloated
history, verbose structured text), estimates token/cost impact, and can
deterministically rewrite prompts without changing meaning.

v1 has **no AI rewrite backend**. The engine is rule-based Go code shared by:

1. CLI (`distilly-lint`)
2. Wails desktop app (React UI)
3. OpenAI-compatible local proxy (`localhost:8787`)

## Goals

- Fast, testable, deterministic lint + optimize
- Clear confidence tiers (auto-apply exact; opt-in for near-dupes / JSON)
- Local persistence of settings and savings metrics (SQLite)
- Drop-in proxy so existing OpenAI SDK clients benefit without code changes

## Non-goals (current)

- Semantic AI compression / local model backends (deferred)
- Native Anthropic Messages API (Claude via OpenAI-compatible gateways only)
- Streaming chat completions through the proxy
- Repo-wide code-context selection (Milestone 5 / Tree-sitter)
