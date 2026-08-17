package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"distilly/internal/api"
	"distilly/internal/proxy"
	"distilly/internal/store"
)

// App is the Wails-bound application struct.
type App struct {
	ctx   context.Context
	store *store.Store

	proxyMu sync.Mutex
	proxy   *proxy.Server
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved so we can
// call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	path, err := defaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "distilly: resolve db path: %v\n", err)
		return
	}
	s, err := store.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "distilly: open store %s: %v\n", path, err)
		return
	}
	a.store = s
}

// shutdown stops the proxy and closes the SQLite store.
func (a *App) shutdown(ctx context.Context) {
	if err := a.StopProxy(); err != nil {
		fmt.Fprintf(os.Stderr, "distilly: stop proxy: %v\n", err)
	}
	if a.store != nil {
		if err := a.store.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "distilly: close store: %v\n", err)
		}
		a.store = nil
	}
}

func defaultDBPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "distilly")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "distilly.db"), nil
}

func (a *App) requireStore() (*store.Store, error) {
	if a == nil || a.store == nil {
		return nil, fmt.Errorf("store not available")
	}
	return a.store, nil
}

// Greet returns a greeting for the given name. Placeholder binding for the
// Wails scaffold; replaced by lint/dashboard/proxy methods in later slices.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, welcome to Distilly!", name)
}

// ListModels returns known model names for the cost/selector UI.
func (a *App) ListModels() []string {
	return api.ListModels()
}

// Analyze lints a prompt and returns a serializable report for the UI.
// Successful analyses are appended to the request log for the Dashboard.
func (a *App) Analyze(req api.AnalyzeRequest) api.AnalyzeResponse {
	resp := api.Analyze(req)
	if err := a.logManualAnalysis(resp); err != nil {
		fmt.Fprintf(os.Stderr, "distilly: log analyze: %v\n", err)
	}
	return resp
}

func (a *App) logManualAnalysis(resp api.AnalyzeResponse) error {
	if a == nil || a.store == nil {
		return nil
	}
	_, err := a.store.InsertRequest(manualRequestFromAnalysis(resp))
	return err
}

func manualRequestFromAnalysis(resp api.AnalyzeResponse) store.Request {
	saved := 0
	if resp.PotentialSavings > 0 && resp.InputTokens > 0 {
		saved = int(math.Round(float64(resp.InputTokens) * resp.PotentialSavings))
	}
	if saved > resp.InputTokens {
		saved = resp.InputTokens
	}
	savingsPct := 0.0
	if resp.InputTokens > 0 {
		savingsPct = float64(saved) / float64(resp.InputTokens) * 100
	}
	return store.Request{
		Source:          store.SourceManual,
		Model:           resp.Model,
		InputTokens:     resp.InputTokens,
		OptimizedTokens: resp.InputTokens - saved,
		SavingsPct:      savingsPct,
		CostUSD:         resp.EstimatedCostUSD,
		SavingsUSD:      resp.EstimatedSavingsUSD,
	}
}

// Apply optimizes a prompt under the given approval flags and returns a
// structured before/after diff.
func (a *App) Apply(req api.ApplyRequest) api.ApplyResponse {
	return api.Apply(req)
}

// DiffForDuplicate returns a structured diff for one duplicate group.
func (a *App) DiffForDuplicate(d api.DuplicateGroup) []api.DiffLine {
	return api.DiffForDuplicate(d)
}

// GetDashboardStats returns aggregate savings for the cost dashboard.
func (a *App) GetDashboardStats() (store.DashboardStats, error) {
	s, err := a.requireStore()
	if err != nil {
		return store.DashboardStats{ByModel: []store.ModelStats{}}, err
	}
	return s.GetDashboardStats()
}

// GetRecentRequests returns the newest logged requests for the dashboard.
func (a *App) GetRecentRequests(limit int) ([]store.Request, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.GetRecentRequests(limit)
}

// GetSetting returns a settings value, or "" if unset.
func (a *App) GetSetting(key string) (string, error) {
	s, err := a.requireStore()
	if err != nil {
		return "", err
	}
	return s.GetSetting(key)
}

// SetSetting upserts a settings key/value pair.
func (a *App) SetSetting(key, value string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	return s.SetSetting(key, value)
}

// LogRequest appends a request row (manual analyze or proxy).
func (a *App) LogRequest(r store.Request) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	_, err = s.InsertRequest(r)
	return err
}
