# App Versioning Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `YYYYMMDD.N` auto-bump versioning with desktop About Distilly and CLI `-version` flags.

**Architecture:** Single embedded file `src/internal/version/VERSION`; package exposes `Base`, `String` (appends `+dev` unless ldflag `release=1`), and `NextVersion`. Bump CLI + shell wrappers sync `wails.json` and build with release ldflags. Desktop uses Wails `Mac.About`; both CLIs print `version.String()`.

**Tech Stack:** Go 1.25, Wails v2, bash scripts, `//go:embed`

## Global Constraints

- Format exactly `YYYYMMDD.N` (local date)
- Dev default: `Base()+"+dev"`; release via `-X distilly/internal/version.release=1`
- Canonical file: `src/internal/version/VERSION`
- No bump on `wails dev` / `go run` / `go test`
- Branch: `feature/app-versioning`

---

### Task 1: Version package (`NextVersion`, `Base`, `String`)

**Files:**
- Create: `src/internal/version/VERSION`
- Create: `src/internal/version/version.go`
- Create: `src/internal/version/version_test.go`
- Create: `src/internal/version/bump.go` (or keep `NextVersion` in `version.go`)

**Interfaces:**
- Produces: `func NextVersion(today, current string) string`, `func Base() string`, `func String() string`, var `release string`

- [ ] **Step 1: Write failing tests for `NextVersion` and `String`**

```go
package version_test

func TestNextVersion_newDay(t *testing.T) {
	got := version.NextVersion("20260828", "20260827.3")
	if got != "20260828.1" { t.Fatalf("got %q", got) }
}
func TestNextVersion_sameDay(t *testing.T) {
	got := version.NextVersion("20260828", "20260828.1")
	if got != "20260828.2" { t.Fatalf("got %q", got) }
}
func TestNextVersion_emptyOrInvalid(t *testing.T) {
	for _, cur := range []string{"", "bogus", "20260828"} {
		got := version.NextVersion("20260828", cur)
		if got != "20260828.1" { t.Fatalf("cur=%q got %q", cur, got) }
	}
}
func TestString_devSuffixByDefault(t *testing.T) {
	// Base from embed; String must end with +dev when release unset
	s := version.String()
	if !strings.HasSuffix(s, "+dev") { t.Fatalf("got %q", s) }
}
```

- [ ] **Step 2: Run tests — expect FAIL** (`NextVersion` undefined)

Run: `cd src && go test ./internal/version/ -v`

- [ ] **Step 3: Implement package**

- Embed `VERSION` (initial content e.g. `20260828.1`)
- `NextVersion`: parse `YYYYMMDD.N`; same day → N+1; else today.1
- `Base()`: trim embed; empty → `"0"`
- `String()`: if `release == "1"` return Base(); else Base()+`+dev`

- [ ] **Step 4: Run tests — expect PASS**

- [ ] **Step 5: Commit** `feat(version): add YYYYMMDD.N package with NextVersion`

---

### Task 2: Bump command

**Files:**
- Create: `src/internal/version/cmd/bump/main.go`
- Modify: `src/wails.json` (add `info.productName` / `info.productVersion` if missing)

**Interfaces:**
- Consumes: `version.NextVersion`, `VERSION` path relative to module
- Produces: updated `VERSION` + `wails.json` productVersion

- [ ] **Step 1: Implement bump main** that reads VERSION, writes `NextVersion(today, current)`, updates `wails.json` via encoding/json map or regexp for `productVersion`

- [ ] **Step 2: Manual check** `cd src && go run ./internal/version/cmd/bump` twice; second bump increments N

- [ ] **Step 3: Commit** `feat(version): add bump command syncing VERSION and wails.json`

---

### Task 3: CLI `-version`

**Files:**
- Modify: `src/cmd/lint/main.go`
- Modify: `src/cmd/context/main.go`

- [ ] **Step 1: Add `-version` bool flag; if set, print `version.String()` and exit 0 before requiring args**

- [ ] **Step 2: Verify** `go run ./cmd/lint -version` prints `…+dev`

- [ ] **Step 3: Commit** `feat(cli): add -version to distilly-lint and distilly-context`

---

### Task 4: Desktop About + wails info

**Files:**
- Modify: `src/main.go`
- Modify: `src/wails.json`

- [ ] **Step 1: Wire `Mac.About` with Title Distilly, Message `Version `+version.String(), Icon from embed if `build/appicon.png` exists**

- [ ] **Step 2: Ensure `info.productVersion` present in wails.json**

- [ ] **Step 3: Commit** `feat(desktop): show version in About Distilly`

---

### Task 5: Build scripts + docs

**Files:**
- Create: `scripts/bump-version.sh`, `scripts/build-desktop.sh`, `scripts/build-cli.sh`, `scripts/build-release.sh`
- Modify: `README.md`, `src/README.md`, `docs/user-guide.md` (CLI `-version` one-liner)

- [ ] **Step 1: Scripts call bump then build with** `-ldflags "-X distilly/internal/version.release=1"` (wails: pass via `-ldflags` or env Wails supports)

- [ ] **Step 2: Document scripts, About, `-version`**

- [ ] **Step 3: `cd src && go test ./...`**

- [ ] **Step 4: Commit** `chore: add version build scripts and docs`

---

## Spec coverage

| Spec item | Task |
|-----------|------|
| `YYYYMMDD.N` + bump rules | 1, 2 |
| `+dev` / release ldflag | 1, 5 |
| About Distilly | 4 |
| CLI `-version` | 3 |
| Scripts + docs | 5 |
| Tests for NextVersion | 1 |
