package main

import (
	"os"
	"path/filepath"
	"testing"

	"distilly/internal/api"
	"distilly/internal/store"
)

func TestAppDashboardBindings(t *testing.T) {
	app := NewApp()
	path := filepath.Join(t.TempDir(), "app.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	app.store = s

	if err := app.SetSetting("upstream_url", "https://example.com"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	got, err := app.GetSetting("upstream_url")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if got != "https://example.com" {
		t.Fatalf("GetSetting = %q", got)
	}

	if err := app.LogRequest(store.Request{
		Source:          store.SourceManual,
		Model:           "gpt-4",
		InputTokens:     100,
		OptimizedTokens: 80,
		SavingsPct:      20,
		CostUSD:         0.01,
		SavingsUSD:      0.002,
	}); err != nil {
		t.Fatalf("LogRequest: %v", err)
	}

	stats, err := app.GetDashboardStats()
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.RequestCount != 1 || stats.TokensSaved != 20 {
		t.Fatalf("stats = %+v", stats)
	}

	recent, err := app.GetRecentRequests(5)
	if err != nil {
		t.Fatalf("GetRecentRequests: %v", err)
	}
	if len(recent) != 1 || recent[0].Model != "gpt-4" {
		t.Fatalf("recent = %+v", recent)
	}
}

func TestAnalyzeLogsManualRequestForDashboard(t *testing.T) {
	app := NewApp()
	path := filepath.Join(t.TempDir(), "app.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	app.store = s

	prompt := readAppFixture(t, "exact_duplicates.txt")
	resp := app.Analyze(api.AnalyzeRequest{Prompt: prompt, Model: "gpt-4"})
	if resp.InputTokens <= 0 {
		t.Fatalf("Analyze InputTokens = %d, want > 0", resp.InputTokens)
	}
	if resp.PotentialSavings <= 0 {
		t.Fatalf("Analyze PotentialSavings = %f, want > 0", resp.PotentialSavings)
	}

	stats, err := app.GetDashboardStats()
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.RequestCount != 1 {
		t.Fatalf("RequestCount = %d, want 1 after Analyze", stats.RequestCount)
	}
	if stats.TokensSaved <= 0 {
		t.Fatalf("TokensSaved = %d, want > 0 after Analyze", stats.TokensSaved)
	}

	recent, err := app.GetRecentRequests(5)
	if err != nil {
		t.Fatalf("GetRecentRequests: %v", err)
	}
	if len(recent) != 1 {
		t.Fatalf("recent len = %d, want 1", len(recent))
	}
	r := recent[0]
	if r.Source != store.SourceManual {
		t.Fatalf("source = %q, want %q", r.Source, store.SourceManual)
	}
	if r.Model != "gpt-4" {
		t.Fatalf("model = %q, want gpt-4", r.Model)
	}
	if r.InputTokens != resp.InputTokens {
		t.Fatalf("logged InputTokens = %d, want %d", r.InputTokens, resp.InputTokens)
	}
	if r.OptimizedTokens >= r.InputTokens {
		t.Fatalf("expected savings: input=%d optimized=%d", r.InputTokens, r.OptimizedTokens)
	}
	if r.SavingsPct <= 0 {
		t.Fatalf("SavingsPct = %v, want > 0", r.SavingsPct)
	}
}

func TestAnalyzeLogsRequestWhenThereAreNoSavings(t *testing.T) {
	app := NewApp()
	path := filepath.Join(t.TempDir(), "app.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	app.store = s

	resp := app.Analyze(api.AnalyzeRequest{Prompt: "You are a helpful assistant.", Model: "gpt-4"})
	if resp.InputTokens <= 0 {
		t.Fatalf("Analyze InputTokens = %d, want > 0", resp.InputTokens)
	}

	recent, err := app.GetRecentRequests(5)
	if err != nil {
		t.Fatalf("GetRecentRequests: %v", err)
	}
	if len(recent) != 1 {
		t.Fatalf("recent len = %d, want 1", len(recent))
	}
	if recent[0].Source != store.SourceManual {
		t.Fatalf("source = %q, want %q", recent[0].Source, store.SourceManual)
	}
	if recent[0].OptimizedTokens != recent[0].InputTokens {
		t.Fatalf("expected no savings: input=%d optimized=%d", recent[0].InputTokens, recent[0].OptimizedTokens)
	}
}

func TestManualRequestFromAnalysisMapsSavings(t *testing.T) {
	got := manualRequestFromAnalysis(api.AnalyzeResponse{
		InputTokens:         100,
		PotentialSavings:    0.2,
		Model:               "gpt-4",
		EstimatedCostUSD:    0.01,
		EstimatedSavingsUSD: 0.002,
	})
	if got.Source != store.SourceManual {
		t.Fatalf("Source = %q, want %q", got.Source, store.SourceManual)
	}
	if got.Model != "gpt-4" {
		t.Fatalf("Model = %q, want gpt-4", got.Model)
	}
	if got.InputTokens != 100 {
		t.Fatalf("InputTokens = %d, want 100", got.InputTokens)
	}
	if got.OptimizedTokens != 80 {
		t.Fatalf("OptimizedTokens = %d, want 80", got.OptimizedTokens)
	}
	if got.SavingsPct != 20 {
		t.Fatalf("SavingsPct = %v, want 20", got.SavingsPct)
	}
	if got.CostUSD != 0.01 || got.SavingsUSD != 0.002 {
		t.Fatalf("cost fields = %+v", got)
	}
}

func TestAppBindingsWithoutStore(t *testing.T) {
	app := NewApp()

	stats, err := app.GetDashboardStats()
	if err == nil {
		t.Fatal("expected error when store is nil")
	}
	if stats.ByModel == nil {
		t.Fatal("ByModel should be empty slice even on error")
	}

	if _, err := app.GetRecentRequests(1); err == nil {
		t.Fatal("expected error when store is nil")
	}
	if _, err := app.GetSetting("x"); err == nil {
		t.Fatal("expected error when store is nil")
	}
	if err := app.SetSetting("x", "y"); err == nil {
		t.Fatal("expected error when store is nil")
	}
	if err := app.LogRequest(store.Request{Source: store.SourceManual}); err == nil {
		t.Fatal("expected error when store is nil")
	}

	resp := app.Analyze(api.AnalyzeRequest{Prompt: "hello", Model: "gpt-4"})
	if resp.InputTokens <= 0 {
		t.Fatalf("Analyze without store InputTokens = %d, want > 0", resp.InputTokens)
	}

	if err := app.StartProxy(); err == nil {
		t.Fatal("expected error when store is nil")
	}
	st := app.GetProxyStatus()
	if st.Running {
		t.Fatalf("status = %+v, want not running", st)
	}
}

func readAppFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "prompts", name))
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	return string(data)
}
