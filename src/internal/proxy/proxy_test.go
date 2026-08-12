package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"distilly/internal/store"
)

func TestRejectStreaming(t *testing.T) {
	s := openStore(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("upstream should not be called for streaming requests")
	}))
	t.Cleanup(upstream.Close)
	mustSet(t, s, "upstream_url", upstream.URL)
	mustSet(t, s, "api_key", "test-key")

	srv := New(s)
	body := `{"model":"gpt-4","stream":true,"messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "streaming not supported") {
		t.Fatalf("body = %q, want streaming rejection message", rec.Body.String())
	}
}

func TestOptimizeForwardAndLogSavings(t *testing.T) {
	s := openStore(t)
	var gotAuth string
	var gotUpstreamBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("upstream path = %q, want /v1/chat/completions", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		var err error
		gotUpstreamBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-test","object":"chat.completion","choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	t.Cleanup(upstream.Close)
	mustSet(t, s, "upstream_url", upstream.URL)
	mustSet(t, s, "api_key", "sk-test")

	// Duplicate system instruction so lint.Apply collapses it.
	dupLine := "Always respond in JSON format."
	messages := []ChatMessage{
		{Role: "system", Content: dupLine + "\nBe concise.\n" + dupLine},
		{Role: "user", Content: "Ping"},
	}
	payload, err := json.Marshal(map[string]any{
		"model":    "gpt-4",
		"stream":   false,
		"messages": messages,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	srv := New(s)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if gotAuth != "Bearer sk-test" {
		t.Fatalf("Authorization = %q, want Bearer sk-test", gotAuth)
	}

	var forwarded chatRequest
	if err := json.Unmarshal(gotUpstreamBody, &forwarded); err != nil {
		t.Fatalf("unmarshal upstream body: %v\n%s", err, gotUpstreamBody)
	}
	if forwarded.Stream {
		t.Fatal("forwarded stream=true")
	}
	sys := findRole(forwarded.Messages, "system")
	if sys == "" {
		t.Fatalf("no system message in forwarded body: %+v", forwarded.Messages)
	}
	if strings.Count(sys, dupLine) != 1 {
		t.Fatalf("system still has duplicate lines:\n%s", sys)
	}

	recent, err := s.GetRecentRequests(1)
	if err != nil {
		t.Fatalf("GetRecentRequests: %v", err)
	}
	if len(recent) != 1 {
		t.Fatalf("logged = %d, want 1", len(recent))
	}
	r := recent[0]
	if r.Source != store.SourceProxy {
		t.Fatalf("source = %q, want proxy", r.Source)
	}
	if r.Model != "gpt-4" {
		t.Fatalf("model = %q, want gpt-4", r.Model)
	}
	if r.InputTokens <= 0 || r.OptimizedTokens <= 0 {
		t.Fatalf("tokens input=%d optimized=%d", r.InputTokens, r.OptimizedTokens)
	}
	if r.OptimizedTokens >= r.InputTokens {
		t.Fatalf("expected savings: input=%d optimized=%d", r.InputTokens, r.OptimizedTokens)
	}
	if r.SavingsPct <= 0 {
		t.Fatalf("SavingsPct = %v, want > 0", r.SavingsPct)
	}
	if r.SavingsUSD <= 0 {
		t.Fatalf("SavingsUSD = %v, want > 0", r.SavingsUSD)
	}
}

func TestPassthroughForwardsOriginalAndStillLogs(t *testing.T) {
	s := openStore(t)
	var gotUpstreamBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotUpstreamBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	t.Cleanup(upstream.Close)
	mustSet(t, s, "upstream_url", upstream.URL)
	mustSet(t, s, "api_key", "sk-test")
	mustSet(t, s, "passthrough", "true")

	dupLine := "Always respond in JSON format."
	messages := []ChatMessage{
		{Role: "system", Content: dupLine + "\n" + dupLine},
		{Role: "user", Content: "Hi"},
	}
	payload, err := json.Marshal(map[string]any{
		"model":    "gpt-4",
		"messages": messages,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

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
		t.Fatalf("unmarshal: %v", err)
	}
	sys := findRole(forwarded.Messages, "system")
	if strings.Count(sys, dupLine) != 2 {
		t.Fatalf("passthrough should keep duplicates:\n%s", sys)
	}

	recent, err := s.GetRecentRequests(1)
	if err != nil {
		t.Fatalf("GetRecentRequests: %v", err)
	}
	if len(recent) != 1 || recent[0].Source != store.SourceProxy {
		t.Fatalf("expected proxy log, got %+v", recent)
	}
	// Metrics still reflect potential optimization even in passthrough.
	if recent[0].InputTokens <= recent[0].OptimizedTokens {
		t.Fatalf("passthrough should still record potential savings: %+v", recent[0])
	}
}

func openStore(t *testing.T) *store.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "proxy.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func mustSet(t *testing.T, s *store.Store, key, value string) {
	t.Helper()
	if err := s.SetSetting(key, value); err != nil {
		t.Fatalf("SetSetting(%q): %v", key, err)
	}
}

func findRole(msgs []ChatMessage, role string) string {
	for _, m := range msgs {
		if m.Role == role {
			return m.Content
		}
	}
	return ""
}
