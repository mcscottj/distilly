package context

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ModuleInfo holds parsed go.mod module path and replace directives.
type ModuleInfo struct {
	Path     string
	Replaces map[string]string // old module path -> replacement path (module or filesystem)
}

// ParseGoMod reads a go.mod file and returns its module path and replaces.
func ParseGoMod(path string) (ModuleInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return ModuleInfo{}, err
	}
	defer f.Close()

	info := ModuleInfo{Replaces: make(map[string]string)}
	sc := bufio.NewScanner(f)
	inReplaceBlock := false

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}

		if inReplaceBlock {
			if line == ")" {
				inReplaceBlock = false
				continue
			}
			oldPath, newPath, ok := parseReplaceFields(line)
			if !ok {
				return ModuleInfo{}, fmt.Errorf("parse go.mod replace: %q", line)
			}
			info.Replaces[oldPath] = newPath
			continue
		}

		switch {
		case strings.HasPrefix(line, "module "):
			info.Path = strings.TrimSpace(strings.TrimPrefix(line, "module "))
			info.Path = strings.Trim(info.Path, `"`)
		case line == "replace (":
			inReplaceBlock = true
		case strings.HasPrefix(line, "replace "):
			rest := strings.TrimSpace(strings.TrimPrefix(line, "replace "))
			oldPath, newPath, ok := parseReplaceFields(rest)
			if !ok {
				return ModuleInfo{}, fmt.Errorf("parse go.mod replace: %q", line)
			}
			info.Replaces[oldPath] = newPath
		}
	}
	if err := sc.Err(); err != nil {
		return ModuleInfo{}, err
	}
	if info.Path == "" {
		return ModuleInfo{}, fmt.Errorf("go.mod missing module directive: %s", path)
	}
	return info, nil
}

// parseReplaceFields parses "old [version] => new [version]" and returns old/new paths.
func parseReplaceFields(line string) (oldPath, newPath string, ok bool) {
	parts := strings.Split(line, "=>")
	if len(parts) != 2 {
		return "", "", false
	}
	oldFields := strings.Fields(strings.TrimSpace(parts[0]))
	newFields := strings.Fields(strings.TrimSpace(parts[1]))
	if len(oldFields) == 0 || len(newFields) == 0 {
		return "", "", false
	}
	return oldFields[0], newFields[0], true
}
