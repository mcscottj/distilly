package version_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"distilly/internal/version"
)

func TestApplyBump_writesNextAndWails(t *testing.T) {
	dir := t.TempDir()
	verPath := filepath.Join(dir, "VERSION")
	wailsPath := filepath.Join(dir, "wails.json")
	if err := os.WriteFile(verPath, []byte("20260828.1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(wailsPath, []byte(`{"name":"distilly","info":{"productName":"Old"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	next, err := version.ApplyBump(verPath, wailsPath, "20260828")
	if err != nil {
		t.Fatal(err)
	}
	if next != "20260828.2" {
		t.Fatalf("next=%q", next)
	}

	raw, err := os.ReadFile(verPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "20260828.2\n" {
		t.Fatalf("VERSION=%q", raw)
	}

	var cfg map[string]any
	if err := json.Unmarshal(mustRead(t, wailsPath), &cfg); err != nil {
		t.Fatal(err)
	}
	info, _ := cfg["info"].(map[string]any)
	if info["productVersion"] != "20260828.2" || info["productName"] != "Distilly" {
		t.Fatalf("info=%v", info)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
