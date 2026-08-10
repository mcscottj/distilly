package main

import (
	"context"
	"fmt"

	"distilly/internal/api"
)

// App is the Wails-bound application struct.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved so we can
// call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
func (a *App) Analyze(req api.AnalyzeRequest) api.AnalyzeResponse {
	return api.Analyze(req)
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
