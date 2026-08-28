// Command bump increments Distilly's YYYYMMDD.N VERSION and syncs wails.json.
//
// Run from the Go module root (src/):
//
//	go run ./internal/version/cmd/bump
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"distilly/internal/version"
)

func main() {
	today := time.Now().Format("20060102")
	verPath := filepath.Join("internal", "version", "VERSION")
	wailsPath := "wails.json"

	next, err := version.ApplyBump(verPath, wailsPath, today)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bump: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(next)
}
