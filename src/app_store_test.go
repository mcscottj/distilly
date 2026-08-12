package main

import (
	"path/filepath"
	"testing"

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
	if err := app.StartProxy(); err == nil {
		t.Fatal("expected error when store is nil")
	}
	st := app.GetProxyStatus()
	if st.Running {
		t.Fatalf("status = %+v, want not running", st)
	}
}
