# Tech context

## Stack

- **Backend:** Go (module root under `src/`)
- **Desktop:** Wails v2 (`main.go`, `app.go`, `wails.json`)
- **Frontend:** React + Vite + Tailwind (`src/frontend/`)
- **DB:** SQLite via `internal/store`
- **Tokenizer:** tiktoken-compatible wrapper in `internal/tokenizer`
- **Tests:** `go test ./...` from `src/` (12 `*_test.go` files)

## Commands (from `src/`)

```bash
go test ./...
go run ./cmd/lint testdata/prompts/exact_duplicates.txt
wails doctor    # CLI at ~/go/bin/wails (not Homebrew-global by default)
wails dev
wails build     # produces src/build/bin/distilly.app
```

Frontend: `npm run build` under `src/frontend/`.

## Layout

```
distilly/
├── docs/                 # roadmap, architecture
├── memory-bank/          # agent working memory (this folder)
└── src/                  # Go module + Wails root
    ├── main.go, app.go
    ├── cmd/lint/
    ├── internal/{api,lint,proxy,store,tokenizer,dedupe,cost,diff,history,regression}/
    ├── frontend/
    └── testdata/prompts/
```

## Constraints / notes

- No CI workflows yet (roadmap still open: wire regression into CI)
- Tree-sitter listed as future M5 tech — not integrated
- `.cursor/` and `local_docs/` are gitignored
- Proxy rejects `stream: true`; no native Anthropic Messages API
