package context

import (
	"fmt"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// ResolveImport maps an import path to repo-relative .go files in this module.
// External modules (not under Module.Path and not covered by a local replace) return nil, nil.
func (idx *Index) ResolveImport(importPath string) ([]string, error) {
	if idx == nil {
		return nil, fmt.Errorf("ResolveImport: nil index")
	}
	importPath = strings.TrimSpace(importPath)
	if importPath == "" {
		return nil, fmt.Errorf("ResolveImport: empty import path")
	}

	dirRel, ok := idx.importDir(importPath)
	if !ok {
		return nil, nil
	}

	wantDir := path.Clean(filepath.ToSlash(dirRel))
	var out []string
	for p, fi := range idx.Files {
		if fi == nil {
			continue
		}
		dir := path.Clean(filepath.ToSlash(filepath.Dir(p)))
		if dir == wantDir {
			out = append(out, p)
		}
	}
	slices.Sort(out)
	return out, nil
}

// importDir returns the repo-relative directory for an in-module (or local-replaced) import path.
func (idx *Index) importDir(importPath string) (string, bool) {
	mod := idx.Module.Path

	for _, old := range slices.Sorted(maps.Keys(idx.Module.Replaces)) {
		repl := idx.Module.Replaces[old]
		if importPath != old && !strings.HasPrefix(importPath, old+"/") {
			continue
		}
		if !isFilesystemReplace(repl) {
			return "", false
		}
		suffix := strings.TrimPrefix(importPath, old)
		suffix = strings.TrimPrefix(suffix, "/")
		joined := path.Clean(filepath.ToSlash(filepath.Join(repl, suffix)))
		return joined, true
	}

	if importPath == mod {
		return ".", true
	}
	if strings.HasPrefix(importPath, mod+"/") {
		return path.Clean(strings.TrimPrefix(importPath, mod+"/")), true
	}
	return "", false
}

func isFilesystemReplace(repl string) bool {
	return strings.HasPrefix(repl, ".") || strings.HasPrefix(repl, "/") || filepath.VolumeName(repl) != ""
}
