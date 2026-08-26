# Distilly user guide

Distilly is a local-first prompt linter and optimizer. It finds waste in LLM prompts — duplicate instructions, redundant examples, bloated history, verbose structured text — estimates token and dollar impact, and can rewrite the prompt **without calling another model**.

Think of it as ESLint for prompts: deterministic rules, explicit opt-ins for riskier changes, everything stays on your machine.

This guide covers what the product does **today** (desktop app, CLI, and local proxy).

---

## Ways to use Distilly

| Surface | What it’s for |
|---------|----------------|
| **Desktop app** | Paste a prompt, see score and suggestions, preview a rewrite, select code context, manage proxy and settings |
| **CLI** (`distilly-lint`, `distilly-context`) | Lint or fix a prompt file; select relevant Go files for LLM context |
| **Local proxy** | Point an OpenAI-compatible client at Distilly; it optimizes chat requests (and optional code context) before forwarding upstream |

There is **no AI rewrite backend**. Near-duplicate detection and JSON conversion are rule-based heuristics that require your approval.

---

## Getting started

Commands run from the `src/` directory (Go module / Wails root). To install
`distilly-lint` and `distilly-context` on your `PATH` instead, see
[CLI → Install](#install).

```bash
# Desktop app
wails doctor   # once, to verify the toolchain
wails dev      # run the app

# CLI — report
go run ./cmd/lint testdata/prompts/exact_duplicates.txt

# CLI — rewrite (exact duplicates only by default)
go run ./cmd/lint -fix testdata/prompts/exact_duplicates.txt

# CLI — code context report
go run ./cmd/context -repo . -seed internal/lint/apply.go -question "why reject streaming?"
```

Settings and request metrics are stored in SQLite under your user config directory (typically `…/distilly/distilly.db`). Secrets and preferences never leave this machine unless you configure an upstream API for the proxy.

---

## How Distilly reads a prompt

Distilly splits text into four **sections**:

| Section | Meaning |
|---------|---------|
| **System** | Instructions, persona, constraints |
| **Examples** | Few-shot `Example N:` blocks |
| **History** | Prior `User:` / `Assistant:` turns |
| **Question** | The current ask |

### Labeled prompts (recommended)

Use explicit headers so section counts and optimizations are accurate:

```text
System:
You are a helpful assistant.
Always respond in JSON.

Example 1:
Q: Capital of France?
A: {"answer":"Paris"}

History:
User: Earlier context…
Assistant: Earlier reply…

Question:
What is the capital of Spain?
```

Headers are matched case-insensitively (`System:`, `System Prompt:`, `Example 1:`, `History:`, `Question:`).

### Unlabeled prompts

Text before the first header — or an entire prompt with **no** headers — is treated as **System**. A bare question with no `Question:` line will show up under System, not Question.

### Chat messages (proxy)

When traffic goes through the local proxy, Distilly maps OpenAI-style messages roughly as:

- `role: system` → System  
- Earlier `user` / `assistant` pairs → History  
- Final `user` message → Question  

---

## What Distilly looks for

| Finding | What it means | Applied automatically? |
|---------|----------------|-------------------------|
| **Exact duplicate instructions** | Same system line repeated | Yes (safe) |
| **Near-duplicate instructions** | Very similar system lines (cosmetic rewording) | Only if you approve |
| **Duplicate examples** | Identical few-shot blocks | Yes (safe) |
| **Near-duplicate examples** | Very similar few-shot blocks | Only if you approve |
| **Structured data** | Runs of `Key: value` lines in System that could be one JSON object | Only if you approve |
| **Long history** | More than 10 parsed User/Assistant turns | Flagged only — not rewritten yet |

**Potential savings %** in the report is based on high-confidence waste Distilly can quantify today (mainly exact duplicate lines/examples). Near-duplicates, JSON conversion, and history compression are suggested for review but may not move that percentage until you opt in or until history compression ships.

### Score (0–100)

Analyze starts at 100 and subtracts for each finding type (exact/near instructions, exact/near examples, structured data, long history). Higher is cleaner.

### Cost estimates

If you pick a **model** Distilly knows pricing for, the UI can show estimated input cost and potential dollar savings. Unknown models still get token counts; dollar fields stay hidden when cost isn’t known.

---

## Safety model (important)

Distilly refuses to silently “over-optimize” constraints.

- **Exact duplicates** — collapsing repeats keeps one copy of each distinct instruction. Safe by default.
- **Near-duplicates** — similarity is a guess. Approving can merge lines that look alike but mean different things (e.g. two different regions). Review the diff first.
- **JSON conversion** — changes format, not just length. Never automatic.
- **History** — Distilly can warn that a chat is long; it does **not** summarize or drop turns yet.

When in doubt: Analyze and Preview apply with toggles **off**, read the before/after, then enable approvals only for changes you’re comfortable with.

---

## Desktop app

The sidebar has four pages: **Lint**, **Context**, **Dashboard**, and **Settings**.

### Lint workspace

Paste a full prompt, analyze it, then preview an optimized version.

| Control | What it does |
|---------|----------------|
| **Model** | Used for token→$ estimates. Defaults from Settings when set. |
| **Prompt** editor | Paste labeled or unlabeled prompt text. Font size follows Settings → Editor text. |
| **Clear** | Empties the editor and clears analysis / diff results. |
| **Analyze** | Runs the linter. Shows score, section tokens, suggestions, and cost (when known). Does **not** rewrite the prompt. |
| **Preview apply** / **Refresh diff** | Runs the optimizer with your current approve toggles and shows before/after + unified diff. |
| **Approve near-duplicates** | Opt in to collapsing near-duplicate instructions and examples. |
| **Approve JSON conversion** | Opt in to rewriting `Key: value` runs into compact JSON. |
| **Use optimized in editor** | Replaces the editor contents with the optimized text and clears the current results so you can re-analyze. |

Toggles in the workspace are initialized from **Settings → Optimization defaults**, but changing them on this page only affects the current session’s Apply (and re-runs Apply if a preview is already open). Save Settings if you want those defaults next launch.

**Analyze vs Apply:** Analyze reports. Apply rewrites. You can Analyze without applying, or Preview apply without copying the result into the editor.

### Context workspace

Select relevant Go source files for LLM context from a repo — deterministic import-graph analysis (Tree-sitter + `go.mod` resolution), no embeddings.

| Control | What it does |
|---------|----------------|
| **Repo root** | Path to the Go module root. Defaults from Settings → Code context when set. |
| **Seed file** | Repo-relative `.go` file to start from (e.g. `internal/proxy/proxy.go`). |
| **Question** | Guides symbol-name matching for low-priority files. |
| **Max depth / Max tokens** | Import hop limit and token budget. Defaults from Settings. |
| **Include test files** | When checked, includes `*_test.go` siblings and imports. |
| **Select context** | Runs selection; shows a table of paths, token counts, and inclusion reasons. |
| **Copy markdown** | Copies formatted `### path` + fenced Go blocks ready to paste into a prompt. |
| **Open in Lint** | Sends the markdown into the Lint workspace editor for further analysis. |

High- and medium-priority files (seed, same package, direct imports, reverse importers) are always included. Low-priority transitive imports are dropped when the token or file budget is exceeded.

### Dashboard

Aggregate view of logged optimization activity.

| Element | What it shows |
|---------|----------------|
| **Requests analyzed** | Count of logged requests |
| **Tokens saved** | Sum of recorded token savings |
| **Estimated $ saved** | Sum of recorded dollar savings |
| **Per-model breakdown** | Same metrics grouped by model name |
| **Recent requests** | Latest rows: time, source, model, tokens before→after, % and $ saved |
| **Refresh** | Reloads stats from the local database |

Rows are written when you **Analyze** a prompt in Lint, and when traffic goes through the **local proxy**.

Clicking a recent row opens **Lint** with that row’s **model** selected. Prompt text is **not** stored in the request log.

### Settings

Everything here is local SQLite configuration for this machine.

#### Appearance

| Setting | What it does |
|---------|----------------|
| **Theme** | Light, Dark, or System |
| **High contrast** | Stronger text, borders, and focus rings |
| **Interface text** | Default / Large / Extra large for labels and chrome |
| **Editor text** | 12–18 px for the Lint editor, diffs, and monospace panels |

Appearance changes save immediately (no separate Save click).

#### Upstream

| Setting | What it does |
|---------|----------------|
| **Upstream base URL** | OpenAI-compatible API root the proxy forwards to (OpenAI, OpenRouter, Azure-style gateways, local servers, etc.) |
| **API key** | Sent as `Authorization: Bearer …` on proxied requests. Stored locally only; not logged. |
| **Default model** | Pre-selected in the Lint workspace model picker |

#### Local proxy

| Setting / control | What it does |
|-------------------|--------------|
| **Start / Stop proxy** | Binds an OpenAI-compatible server on `127.0.0.1` at the configured port. Start saves settings first. |
| **Proxy port** | Default `8787`. Change only while the proxy is stopped. |
| **Base URL** | Shown as `http://127.0.0.1:<port>/v1` — copy this into your client’s `baseURL`. |
| **Passthrough mode** | Still analyzes and logs potential savings, but forwards the **original** request body unchanged (no rewrite). |

**Streaming is not supported.** Clients must send `stream: false`. Streaming requests are rejected with 400, but Distilly still logs them on the Dashboard as `proxy-stream` (token counts only; no savings).

Point an SDK at Distilly like:

```text
base URL: http://127.0.0.1:8787/v1
API key:  (your upstream key, or whatever your client requires — Distilly uses the key from Settings when forwarding)
```

The ChatGPT desktop/web app cannot use this URL — it has no custom API endpoint
and it streams. For a smoke test from the repo root, start the proxy, then:

```bash
./test-proxy.sh
```

That sends the same non-streaming `POST /v1/chat/completions` as:

```bash
curl -sS http://127.0.0.1:8787/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gpt-4o","stream":false,"messages":[{"role":"user","content":"ping"}]}'
```

Then open **Dashboard** or click **Refresh**. Override the port with
`DISTILLY_PROXY_PORT` if it is not `8787`.

#### Optimization defaults

| Setting | What it does |
|---------|----------------|
| **Approve near-duplicates by default** | Pre-checks the Lint toggle; also used when the proxy applies optimizations |
| **Approve JSON conversion by default** | Same for JSON conversion |

Exact duplicates always apply. These two stay off unless you opt in — same policy as the CLI flags.

#### Code context

| Setting | What it does |
|---------|----------------|
| **Repo root** | Default Go module path for the Context workspace and proxy. |
| **Enable code context in proxy** | When on, requests with an `@distilly:context` marker get file selection injected (see below). |
| **Max import depth** | Transitive import hops (default `2`). |
| **Max context tokens** | Token budget for selected files (default `32000`). |

Click **Save settings** to persist Upstream, Local proxy, Optimization defaults, and Code context.

---

## CLI

### Install

Install once from the repo root so you can call the CLIs from any directory
(Go 1.20+):

```bash
go install -C src -o "$(go env GOPATH)/bin/distilly-lint" ./cmd/lint
go install -C src -o "$(go env GOPATH)/bin/distilly-context" ./cmd/context
```

Ensure `$(go env GOPATH)/bin` (typically `~/go/bin`) is on your `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"   # add to ~/.zshrc to persist
```

The `-o` flag sets the binary name. Without it, `go install ./cmd/lint` would
install `lint`, not `distilly-lint`.

Re-run the install commands after pulling changes that touch `cmd/lint` or
`cmd/context`.

Examples after install:

```bash
distilly-lint /path/to/prompt.txt
distilly-lint -fix -approve-near-duplicates /path/to/prompt.txt

distilly-context -repo /path/to/repo -seed internal/lint/apply.go -question "why reject streaming?"
distilly-context -repo /path/to/repo -seed internal/lint/apply.go -format markdown
```

When working inside the repo without installing, use `go run` from `src/` (see
below). For `distilly-context`, pass an absolute `-repo` path when your shell
is not already in that repository.

### `distilly-lint`

```bash
go run ./cmd/lint [flags] <prompt-file>
```

| Flag | Effect |
|------|--------|
| *(none)* | Print the lint report (tokens, findings, suggestions, potential savings) |
| `-model <name>` | Include USD estimates for a known model |
| `-fix` | Print the optimized prompt instead of the report |
| `-approve-near-duplicates` | With `-fix`, also collapse near-duplicates |
| `-approve-json-conversion` | With `-fix`, also convert structured `Key: value` runs to JSON |

Examples:

```bash
go run ./cmd/lint -model gpt-4 testdata/prompts/near_duplicates.txt

go run ./cmd/lint -fix -approve-near-duplicates -approve-json-conversion \
  testdata/prompts/structured_data.txt
```

Sample prompts for trying features live under `src/testdata/prompts/`.

### `distilly-context`

```bash
go run ./cmd/context -repo <path> -seed <repo-relative.go> [flags]
```

| Flag | Effect |
|------|--------|
| `-repo` | Repository root (default `.`) |
| `-seed` | **Required.** Repo-relative seed file |
| `-question` | Question for symbol matching |
| `-max-depth` | Import graph depth (default `2`) |
| `-max-files` | File cap (default `20`) |
| `-max-tokens` | Token budget (default `32000`; `0` = no cap) |
| `-include-tests` | Include `*_test.go` files |
| `-format` | `report` (default), `json`, or `markdown` |

Examples:

```bash
go run ./cmd/context -repo testdata/repos/mini -seed a/a.go -question "how does A call B?"
go run ./cmd/context -repo . -seed internal/proxy/proxy.go -format markdown
```

Miniature repo fixtures are documented in [code-context-fixtures.md](code-context-fixtures.md).

---

## Typical workflows

### 1. Clean up a prompt you’re about to send

1. Open **Lint**, paste the prompt, pick a model, click **Analyze**.  
2. Read score, section bars, and suggestions.  
3. Click **Preview apply** with both approvals off.  
4. If the after-text looks right, optionally enable near-duplicate / JSON toggles and refresh.  
5. **Use optimized in editor** (or copy from the After panel) and send that text to your model.

### 2. Sit in front of an app’s API calls

1. In **Settings**, set upstream URL + API key, confirm port, leave passthrough off (unless you only want metrics).  
2. Set optimization defaults you’re comfortable with.  
3. **Start proxy**, copy the base URL into your client, disable streaming.  
   Smoke test from the repo root with `./test-proxy.sh` (not the ChatGPT app).  
4. Use your app normally; check **Dashboard** for tokens/$ saved.  
5. Use **Passthrough** when you want measurement without changing what the model sees.

### 3. Scripted / CI-style checks

Use the CLI report on checked-in prompt files. Pair with fixtures under `testdata/prompts/` when you want known waste patterns.

### 4. Trim repo context for a coding question

1. Open **Context**, set repo root and seed file, enter your question.  
2. Click **Select context**, review the file table and token total.  
3. **Copy markdown** or **Open in Lint** to analyze token weight before sending upstream.  
4. For automated proxy injection, enable code context in Settings and add a marker to the system message (see below).

---

## Proxy code context marker

When **Enable code context in proxy** is on and **Repo root** is set, embed this in a `role: system` message:

```text
@distilly:context
seed=internal/proxy/proxy.go
```

Distilly replaces the entire marker block (through the next blank line) with markdown code blocks for selected files. The user's last message becomes the selection question. Without the marker, the proxy does not auto-select files — even when the setting is enabled.

Example system message:

```text
You are a senior Go engineer reviewing this codebase.

@distilly:context
seed=internal/proxy/proxy.go

Always cite file paths in your answer.
```

After optimization, the forwarded request contains the selected file contents instead of the marker.

---

## What Distilly does *not* do yet

- Summarize or rewrite long conversation history (it only flags it)  
- Stream chat completions through the proxy  
- Call a local or cloud model to “semantically compress” prompts  
- Speak Anthropic’s native Messages API (Claude works only via OpenAI-compatible gateways)

For internals and milestone status, see [architecture.md](architecture.md) and [roadmap.md](roadmap.md).
