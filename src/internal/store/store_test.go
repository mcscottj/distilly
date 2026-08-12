package store

import (
	"path/filepath"
	"testing"
)

func TestOpenCreatesSchemaAndSettingsRoundTrip(t *testing.T) {
	s := openTemp(t)

	if err := s.SetSetting("upstream_url", "https://api.openai.com"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	got, err := s.GetSetting("upstream_url")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if got != "https://api.openai.com" {
		t.Fatalf("GetSetting = %q, want upstream URL", got)
	}

	missing, err := s.GetSetting("missing_key")
	if err != nil {
		t.Fatalf("GetSetting missing: %v", err)
	}
	if missing != "" {
		t.Fatalf("GetSetting missing = %q, want empty", missing)
	}
}

func TestInsertRequestAndRecentRequests(t *testing.T) {
	s := openTemp(t)

	id, err := s.InsertRequest(Request{
		Source:          SourceManual,
		Model:           "gpt-4",
		InputTokens:     1000,
		OptimizedTokens: 700,
		SavingsPct:      30,
		CostUSD:         0.03,
		SavingsUSD:      0.01,
	})
	if err != nil {
		t.Fatalf("InsertRequest: %v", err)
	}
	if id <= 0 {
		t.Fatalf("InsertRequest id = %d, want > 0", id)
	}

	_, err = s.InsertRequest(Request{
		Source:          SourceProxy,
		Model:           "gpt-4o",
		InputTokens:     500,
		OptimizedTokens: 400,
		SavingsPct:      20,
		CostUSD:         0.01,
		SavingsUSD:      0.002,
	})
	if err != nil {
		t.Fatalf("InsertRequest second: %v", err)
	}

	recent, err := s.GetRecentRequests(10)
	if err != nil {
		t.Fatalf("GetRecentRequests: %v", err)
	}
	if len(recent) != 2 {
		t.Fatalf("GetRecentRequests len = %d, want 2", len(recent))
	}
	// Newest first.
	if recent[0].Model != "gpt-4o" || recent[0].Source != SourceProxy {
		t.Fatalf("first recent = %+v, want gpt-4o proxy", recent[0])
	}
	if recent[1].Model != "gpt-4" || recent[1].Source != SourceManual {
		t.Fatalf("second recent = %+v, want gpt-4 manual", recent[1])
	}
	if recent[0].CreatedAt == "" || recent[1].CreatedAt == "" {
		t.Fatal("expected CreatedAt to be set")
	}

	limited, err := s.GetRecentRequests(1)
	if err != nil {
		t.Fatalf("GetRecentRequests limit: %v", err)
	}
	if len(limited) != 1 {
		t.Fatalf("GetRecentRequests(1) len = %d, want 1", len(limited))
	}
}

func TestGetDashboardStatsAggregates(t *testing.T) {
	s := openTemp(t)

	_, err := s.InsertRequest(Request{
		Source:          SourceManual,
		Model:           "gpt-4",
		InputTokens:     1000,
		OptimizedTokens: 700,
		SavingsPct:      30,
		CostUSD:         0.03,
		SavingsUSD:      0.01,
	})
	if err != nil {
		t.Fatalf("InsertRequest: %v", err)
	}
	_, err = s.InsertRequest(Request{
		Source:          SourceProxy,
		Model:           "gpt-4",
		InputTokens:     400,
		OptimizedTokens: 300,
		SavingsPct:      25,
		CostUSD:         0.012,
		SavingsUSD:      0.003,
	})
	if err != nil {
		t.Fatalf("InsertRequest: %v", err)
	}
	_, err = s.InsertRequest(Request{
		Source:          SourceProxy,
		Model:           "gpt-4o",
		InputTokens:     200,
		OptimizedTokens: 200,
		SavingsPct:      0,
		CostUSD:         0.005,
		SavingsUSD:      0,
	})
	if err != nil {
		t.Fatalf("InsertRequest: %v", err)
	}

	stats, err := s.GetDashboardStats()
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.RequestCount != 3 {
		t.Fatalf("RequestCount = %d, want 3", stats.RequestCount)
	}
	// (1000-700) + (400-300) + (200-200) = 400
	if stats.TokensSaved != 400 {
		t.Fatalf("TokensSaved = %d, want 400", stats.TokensSaved)
	}
	if stats.SavingsUSD < 0.0129 || stats.SavingsUSD > 0.0131 {
		t.Fatalf("SavingsUSD = %f, want ~0.013", stats.SavingsUSD)
	}
	if len(stats.ByModel) != 2 {
		t.Fatalf("ByModel len = %d, want 2", len(stats.ByModel))
	}

	byName := map[string]ModelStats{}
	for _, m := range stats.ByModel {
		byName[m.Model] = m
	}
	gpt4 := byName["gpt-4"]
	if gpt4.RequestCount != 2 || gpt4.TokensSaved != 400 || gpt4.SavingsUSD < 0.0129 {
		t.Fatalf("gpt-4 stats = %+v", gpt4)
	}
	gpt4o := byName["gpt-4o"]
	if gpt4o.RequestCount != 1 || gpt4o.TokensSaved != 0 {
		t.Fatalf("gpt-4o stats = %+v", gpt4o)
	}
}

func TestGetDashboardStatsEmpty(t *testing.T) {
	s := openTemp(t)
	stats, err := s.GetDashboardStats()
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.RequestCount != 0 || stats.TokensSaved != 0 || stats.SavingsUSD != 0 {
		t.Fatalf("empty stats = %+v", stats)
	}
	if stats.ByModel == nil {
		t.Fatal("ByModel should be non-nil empty slice for JSON")
	}
	if len(stats.ByModel) != 0 {
		t.Fatalf("ByModel len = %d, want 0", len(stats.ByModel))
	}
}

func TestSetSettingOverwrites(t *testing.T) {
	s := openTemp(t)
	if err := s.SetSetting("proxy_port", "8787"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if err := s.SetSetting("proxy_port", "9000"); err != nil {
		t.Fatalf("SetSetting overwrite: %v", err)
	}
	got, err := s.GetSetting("proxy_port")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if got != "9000" {
		t.Fatalf("GetSetting = %q, want 9000", got)
	}
}

func openTemp(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open(%q): %v", path, err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})
	return s
}
