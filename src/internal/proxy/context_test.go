package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"distilly/internal/context"
	"distilly/internal/tokenizer"
)

func TestTrimCodeContextReplacesMarkerBlock(t *testing.T) {
	context.ClearIndexCache()
	root := filepath.Join("..", "..", "testdata", "repos", "mini")

	prompt := "System:\nYou are helpful.\n\n@distilly:context\nseed=a/a.go\n\nQuestion: how does A call B?\n"
	cfg := contextTrimConfig{
		RepoRoot:  root,
		MaxDepth:  2,
		MaxTokens: 32000,
	}

	out, err := TrimCodeContext(prompt, cfg)
	if err != nil {
		t.Fatalf("TrimCodeContext: %v", err)
	}
	if strings.Contains(out, "@distilly:context") {
		t.Fatalf("marker still present:\n%s", out)
	}
	if !strings.Contains(out, "### a/a.go") {
		t.Fatalf("expected formatted seed file in output:\n%s", out)
	}
	if !strings.Contains(out, "You are helpful.") {
		t.Fatal("expected prefix system text preserved")
	}
	if !strings.Contains(out, "how does A call B?") {
		t.Fatal("expected question preserved")
	}
}

func TestTrimCodeContextNoOpWithoutMarker(t *testing.T) {
	prompt := "System:\nPlain instructions.\n\nQuestion: hello\n"
	out, err := TrimCodeContext(prompt, contextTrimConfig{
		RepoRoot: filepath.Join("..", "..", "testdata", "repos", "mini"),
	})
	if err != nil {
		t.Fatalf("TrimCodeContext: %v", err)
	}
	if out != prompt {
		t.Fatalf("expected unchanged prompt, got:\n%s", out)
	}
}

func TestProxyContextMarkerForward(t *testing.T) {
	context.ClearIndexCache()
	repoRoot := filepath.Join("..", "..", "testdata", "repos", "mini")

	s := openStore(t)
	mustSet(t, s, "upstream_url", "")
	mustSet(t, s, "api_key", "sk-test")
	mustSet(t, s, SettingRepoRoot, repoRoot)
	mustSet(t, s, SettingCodeContextEnabled, "true")
	mustSet(t, s, SettingContextMaxDepth, "2")
	mustSet(t, s, SettingContextMaxTokens, "32000")

	// Simulated full-repo paste (old workflow) for token comparison.
	fullDump := strings.Repeat("// filler line with tokens\n", 500) + "### noise/noise.go\n```go\npackage noise\n```\n"

	var gotUpstreamBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotUpstreamBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"ok","choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	t.Cleanup(upstream.Close)
	mustSet(t, s, "upstream_url", upstream.URL)

	system := "You are helpful.\n\n@distilly:context\nseed=a/a.go\n"
	messages := []ChatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: "how does A call B?"},
	}
	payload, err := json.Marshal(map[string]any{
		"model":    "gpt-4",
		"stream":   false,
		"messages": messages,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	fullPasteMessages := []ChatMessage{
		{Role: "system", Content: "You are helpful.\n\n" + fullDump},
		{Role: "user", Content: "how does A call B?"},
	}
	beforeTokens := tokenizer.Count(MessagesToPrompt(fullPasteMessages))

	srv := New(s)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}

	var forwarded chatRequest
	if err := json.Unmarshal(gotUpstreamBody, &forwarded); err != nil {
		t.Fatalf("unmarshal upstream: %v\n%s", err, gotUpstreamBody)
	}
	sys := findRole(forwarded.Messages, "system")
	if sys == "" {
		t.Fatal("no system message forwarded")
	}
	if strings.Contains(sys, "@distilly:context") {
		t.Fatalf("marker not replaced:\n%s", sys)
	}
	if strings.Contains(sys, "noise/noise.go") {
		t.Fatalf("selection should exclude noise/noise.go:\n%s", sys)
	}
	if !strings.Contains(sys, "### a/a.go") {
		t.Fatalf("expected selected a/a.go in forwarded system:\n%s", sys)
	}

	afterTokens := tokenizer.Count(MessagesToPrompt(forwarded.Messages))
	if afterTokens >= beforeTokens {
		t.Fatalf("expected token decrease vs full-repo paste: before=%d after=%d", beforeTokens, afterTokens)
	}
}

func TestExtractContextBlock(t *testing.T) {
	rest, seed, ok := extractContextBlock("prefix\n\n@distilly:context\nseed=foo.go\n\nsuffix")
	if !ok {
		t.Fatal("expected found")
	}
	if seed != "foo.go" {
		t.Fatalf("seed = %q, want foo.go", seed)
	}
	if !strings.Contains(rest, "prefix") || !strings.Contains(rest, "suffix") {
		t.Fatalf("rest = %q, want prefix and suffix", rest)
	}
}

func TestProxyContextDisabledNoOp(t *testing.T) {
	s := openStore(t)
	mustSet(t, s, "upstream_url", "")
	mustSet(t, s, SettingCodeContextEnabled, "false")

	var gotUpstreamBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUpstreamBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	t.Cleanup(upstream.Close)
	mustSet(t, s, "upstream_url", upstream.URL)

	marker := "@distilly:context\nseed=a/a.go"
	messages := []ChatMessage{
		{Role: "system", Content: marker},
		{Role: "user", Content: "hi"},
	}
	payload, _ := json.Marshal(map[string]any{"model": "gpt-4", "stream": false, "messages": messages})

	srv := New(s)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	var forwarded chatRequest
	_ = json.Unmarshal(gotUpstreamBody, &forwarded)
	sys := findRole(forwarded.Messages, "system")
	if !strings.Contains(sys, "@distilly:context") {
		t.Fatalf("disabled context should leave marker intact:\n%s", sys)
	}
}

func TestApplyCodeContextIntegrationUsesRepoFromStore(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", "testdata", "repos", "mini"))
	if err != nil {
		t.Fatalf("Abs: %v", err)
	}
	if _, err := os.Stat(repoRoot); err != nil {
		t.Fatalf("fixture repo: %v", err)
	}

	prompt := "System:\n@distilly:context\nseed=a/a.go\n\nQuestion: test\n"
	cfg := runtimeConfig{
		CodeContextEnabled: true,
		RepoRoot:           repoRoot,
		ContextMaxDepth:    2,
		ContextMaxTokens:   32000,
	}
	out := applyCodeContext(prompt, cfg)
	if strings.Contains(out, "@distilly:context") {
		t.Fatalf("marker not replaced:\n%s", out)
	}
}
