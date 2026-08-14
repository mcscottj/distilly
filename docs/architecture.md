# Distilly architecture overview

**Distilly** is a local-first “ESLint for prompts”: it analyzes LLM prompts, finds waste (duplicates, bloated history, verbose structured text), estimates token/cost impact, and can deterministically rewrite prompts without changing meaning. No AI rewrite backend in v1 — the engine is rule-based Go code.

You interact with it in three ways:

1. **Desktop app** (Wails: Go + React) — paste a prompt, analyze, apply fixes, see dashboard/settings
2. **CLI** (`distilly-lint`) — lint a prompt file from the terminal
3. **Local proxy** (OpenAI-compatible) — point your app at `localhost:8787`; Distilly optimizes chat requests before forwarding upstream

```mermaid
flowchart TB
  subgraph users [How you use it]
    Dev[Developer]
    AppClient[Your app via OpenAI SDK]
  end

  subgraph desktop [Wails desktop]
    UI[React UI]
    App[App bindings]
    Store[(SQLite)]
  end

  Engine[lint engine]
  Proxy[OpenAI-compatible proxy]
  Upstream[Upstream API]

  Dev --> UI
  Dev --> CLI[distilly-lint CLI]
  AppClient --> Proxy
  UI --> App
  App --> Engine
  App --> Store
  CLI --> Engine
  Proxy --> Engine
  Proxy --> Store
  Proxy --> Upstream
```

## Core idea

Prompts are split into **sections** (`System` / `Examples` / `History` / `Question`), checked for waste, optionally rewritten, and scored — whether you pasted text in the UI or sent chat `messages` through the proxy.

## Package map

| Area | Path | Role |
|------|------|------|
| Desktop entry | [`src/main.go`](../src/main.go), [`src/app.go`](../src/app.go) | Wails lifecycle; binds UI → Go |
| CLI | [`src/cmd/lint`](../src/cmd/lint) | Headless lint |
| Lint engine | [`src/internal/lint`](../src/internal/lint) | `Run` (report) + `Apply` (optimize) |
| UI DTOs | [`src/internal/api`](../src/internal/api) | JSON shapes for Analyze/Apply |
| Proxy | [`src/internal/proxy`](../src/internal/proxy) | `/v1/chat/completions` gateway |
| Persistence | [`src/internal/store`](../src/internal/store) | Settings + request metrics |
| Helpers | `tokenizer`, `dedupe`, `history`, `cost`, `diff` | Count, find dups, flag history, $, diffs |
| Frontend | [`src/frontend`](../src/frontend) | Lint / Dashboard / Settings pages |

## Flow 1 — Manual lint (UI or CLI)

```mermaid
sequenceDiagram
  participant User
  participant UI as React_or_CLI
  participant App as App_or_api
  participant Lint as lint_engine

  User->>UI: Paste or file prompt + model
  UI->>App: Analyze
  App->>Lint: Run(prompt, model)
  Lint-->>App: Report tokens issues savings score
  App-->>UI: Show score sections suggestions
  User->>UI: Apply with opt-in flags
  UI->>App: Apply
  App->>Lint: Apply(prompt, options)
  Lint-->>UI: Optimized text + diff
```

**In plain terms:** You give Distilly a prompt. It counts tokens, finds duplicates and other waste, estimates dollars saved, and (if you approve) rewrites the prompt. Exact duplicates apply automatically; near-duplicates and JSON conversion need an explicit opt-in.

## Flow 2 — Proxied chat completion

```mermaid
sequenceDiagram
  participant Client as Your_app
  participant Proxy as local_proxy
  participant Lint as lint_engine
  participant DB as SQLite
  participant API as Upstream_OpenAI

  Client->>Proxy: POST /v1/chat/completions
  Proxy->>Proxy: Reject if stream true
  Proxy->>DB: Load settings key upstream toggles
  Proxy->>Proxy: messages to sectioned prompt
  Proxy->>Lint: Run then Apply
  Proxy->>Proxy: prompt back to messages unless passthrough
  Proxy->>DB: Log source equals proxy
  Proxy->>API: Forward optimized request
  API-->>Client: Response via proxy
```

**In plain terms:** Your code talks to Distilly as if it were OpenAI. Distilly turns chat messages into a sectioned prompt, cleans it up, turns it back into messages, logs savings, and forwards the slimmed request to the real API. Streaming is rejected today. Native Anthropic Messages API is out of scope for M4 — Claude works only via OpenAI-compatible gateways.

Start and stop the proxy from **Settings** (`StartProxy` / `StopProxy` / `GetProxyStatus` on [`src/app.go`](../src/app.go)).

## Desktop UI surface

- **Lint workspace** — editor, model picker, score, sections, suggestions, apply + diff
- **Dashboard** — aggregate tokens/$ saved, recent requests (mainly from proxy logs)
- **Settings** — upstream URL, API key, proxy port, start/stop/status, approval toggles, passthrough

Data lives in SQLite under the user config dir (`…/distilly/distilly.db`).

## User documentation

End-user explanation of every shipped surface (Lint, Dashboard, Settings, CLI, proxy) is in [`user-guide.md`](user-guide.md).

## Milestone context

See also [`roadmap.md`](roadmap.md).

- **M1–M3:** CLI lint, regression harness, confidence-tier apply — largely done
- **M4:** Desktop app + SQLite + proxy lifecycle — done
- **M5 (future):** code-context / Tree-sitter file selection
