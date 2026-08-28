package version

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ApplyBump writes NextVersion(today, current) to versionPath and syncs
// wails.json info.productName / info.productVersion. Returns the new version.
func ApplyBump(versionPath, wailsPath, today string) (string, error) {
	current := ""
	if b, err := os.ReadFile(versionPath); err == nil {
		current = string(b)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("read version: %w", err)
	}

	next := NextVersion(today, current)
	if err := os.WriteFile(versionPath, []byte(next+"\n"), 0o644); err != nil {
		return "", fmt.Errorf("write version: %w", err)
	}
	if err := syncWailsProductVersion(wailsPath, next); err != nil {
		return "", err
	}
	return next, nil
}

func syncWailsProductVersion(wailsPath, productVersion string) error {
	raw, err := os.ReadFile(wailsPath)
	if err != nil {
		return fmt.Errorf("read wails.json: %w", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("parse wails.json: %w", err)
	}
	info, _ := cfg["info"].(map[string]any)
	if info == nil {
		info = map[string]any{}
		cfg["info"] = info
	}
	info["productName"] = "Distilly"
	info["productVersion"] = strings.TrimSpace(productVersion)

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode wails.json: %w", err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(wailsPath, out, 0o644); err != nil {
		return fmt.Errorf("write wails.json: %w", err)
	}
	return nil
}
