# Distilly user guide

Distilly is a local-first prompt linter and optimizer. It finds waste in LLM prompts — duplicate instructions, redundant examples, bloated history, verbose structured text — estimates token and dollar impact, and can rewrite the prompt **without calling another model**.

Think of it as ESLint for prompts: deterministic rules, explicit opt-ins for riskier changes, everything stays on your machine.

This guide covers what the product does **today** (desktop app, CLI, and local proxy).

---

## Ways to use Distilly

| Surface | What it’s for |
|---------|----------------|
| **Desktop app** | Paste a prompt, see score and suggestions, preview a rewrite, manage proxy and settings |
| **CLI** (`distilly-lint`) | Lint or fix a prompt file from the terminal |
| **Local proxy** | Point an OpenAI-compatible client at Distilly; it optimizes chat requests before forwarding upstream |

There is **no AI rewrite backend**. Near-duplicate detection and JSON conversion are rule-based heuristics that require your approval.

---

## Getting started

Commands run from the `src/` directory (Go module / Wails root):

```bash
# Desktop app
wails doctor   # once, to verify the toolchain
wails dev      # run the app

# CLI — report
go run ./cmd/lint testdata/prompts/exact_duplicates.txt

# CLI — rewrite (exact duplicates only by default)
go run ./cmd/lint -fix testdata/prompts/exact_duplicates.txt
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

The sidebar has three pages: **Lint**, **Dashboard**, and **Settings**.

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

**Streaming is not supported.** Clients must send `stream: false`. Streaming requests are rejected.

Point an SDK at Distilly like:

```text
base URL: http://127.0.0.1:8787/v1
API key:  (your upstream key, or whatever your client requires — Distilly uses the key from Settings when forwarding)
```

#### Optimization defaults

| Setting | What it does |
|---------|----------------|
| **Approve near-duplicates by default** | Pre-checks the Lint toggle; also used when the proxy applies optimizations |
| **Approve JSON conversion by default** | Same for JSON conversion |

Exact duplicates always apply. These two stay off unless you opt in — same policy as the CLI flags.

Click **Save settings** to persist Upstream, Local proxy, and Optimization defaults.

---

## CLI

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
4. Use your app normally; check **Dashboard** for tokens/$ saved.  
5. Use **Passthrough** when you want measurement without changing what the model sees.

### 3. Scripted / CI-style checks

Use the CLI report on checked-in prompt files. Pair with fixtures under `testdata/prompts/` when you want known waste patterns.

---

## What Distilly does *not* do yet

- Summarize or rewrite long conversation history (it only flags it)  
- Stream chat completions through the proxy  
- Call a local or cloud model to “semantically compress” prompts  
- Speak Anthropic’s native Messages API (Claude works only via OpenAI-compatible gateways)  
- Automatically pick relevant code files from a repo (roadmap Milestone 5)

For internals and milestone status, see [architecture.md](architecture.md) and [roadmap.md](roadmap.md).
