package context

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// Index is a cached snapshot of Go files in a repository.
type Index struct {
	RepoRoot string
	Module   ModuleInfo
	Files    map[string]*FileInfo // repo-relative path -> info
	Warnings []string
}

var (
	indexCache   sync.Map // abs repo root -> *Index
	indexCacheMu sync.Mutex
)

var skipDirNames = map[string]bool{
	".git":         true,
	"vendor":       true,
	"node_modules": true,
	"build":        true,
}

// BuildIndex walks repoRoot, parses go.mod, and extracts metadata for each .go file.
// Results are cached per absolute repo root. Unparseable files add Warnings and are skipped.
func BuildIndex(repoRoot string) (*Index, error) {
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, err
	}
	if cached, ok := indexCache.Load(abs); ok {
		return cached.(*Index), nil
	}

	indexCacheMu.Lock()
	defer indexCacheMu.Unlock()
	if cached, ok := indexCache.Load(abs); ok {
		return cached.(*Index), nil
	}

	idx, err := buildIndexUncached(abs)
	if err != nil {
		return nil, err
	}
	indexCache.Store(abs, idx)
	return idx, nil
}

// ClearIndexCache drops all cached indexes (tests / forced rebuild).
func ClearIndexCache() {
	indexCache.Range(func(key, _ any) bool {
		indexCache.Delete(key)
		return true
	})
}

func buildIndexUncached(absRoot string) (*Index, error) {
	modPath := filepath.Join(absRoot, "go.mod")
	mod, err := ParseGoMod(modPath)
	if err != nil {
		return nil, fmt.Errorf("index: %w", err)
	}

	idx := &Index{
		RepoRoot: absRoot,
		Module:   mod,
		Files:    make(map[string]*FileInfo),
	}

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			idx.Warnings = append(idx.Warnings, fmt.Sprintf("walk %s: %v", path, walkErr))
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if skipDirNames[name] {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			idx.Warnings = append(idx.Warnings, fmt.Sprintf("rel %s: %v", path, err))
			return nil
		}
		rel = filepath.Clean(rel)

		src, err := os.ReadFile(path)
		if err != nil {
			idx.Warnings = append(idx.Warnings, fmt.Sprintf("read %s: %v", rel, err))
			return nil
		}
		info, err := ExtractGoFile(rel, src)
		if err != nil {
			idx.Warnings = append(idx.Warnings, fmt.Sprintf("%s: %v", rel, err))
			return nil
		}
		idx.Files[rel] = &info
		return nil
	})
	if err != nil {
		return nil, err
	}
	return idx, nil
}
