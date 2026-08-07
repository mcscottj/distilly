package main

import (
	"context"
	"fmt"
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
