# Roadmap

## Guiding principle

Ship the deterministic, rule-based linter first. It's fast, testable, and
gives users immediate value. AI-powered semantic compression comes later
as an optional second pass, compared against the rule-based baseline.

## Shipped

### Prompt linter (CLI)

`distilly-lint` reports token usage by section, flags duplicates and long
history, estimates cost, and can apply high-confidence optimizations
(exact duplicates auto-apply; near-duplicates and JSON conversion require
approval flags).

Includes: tokenizer, section splitter, exact/near dedupe (lines and
few-shot blocks), cost tables, unified diff, structured-data → JSON
conversion (opt-in), and prompt scoring.

### Regression harness

Constraint-survival tests guard `Apply` against dropping instructions
when collapsing duplicates. Run with `go test ./...` under `src/`.

### Desktop app

Wails desktop app with Lint workspace, Context workspace, Dashboard, and
Settings; SQLite for settings and request metrics; OpenAI-compatible
local proxy (`POST /v1/chat/completions`, non-streaming) startable from
Settings.

### Code context optimizer

Tree-sitter based selection of relevant Go files for a question (CLI
`distilly-context`, desktop Context workspace, and proxy
`@distilly:context` marker injection).

## Planned

- Wire the regression suite into CI
- History compression beyond flagging
- True semantic paraphrase detection (embeddings)
- Optional local-model semantic compression (deferred; exists on a held
  branch only until reviewed)

Streaming through the proxy and a native Anthropic Messages API are also
out of scope for now.
