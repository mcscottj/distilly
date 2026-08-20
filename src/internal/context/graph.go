package context

import (
	"path"
	"path/filepath"
	"slices"
)

// Graph is a file-level import adjacency list over an Index.
type Graph struct {
	// Forward maps a file to the repo-relative files it imports (in-module only).
	Forward map[string][]string
	// Reverse maps a file to the repo-relative files that import it.
	Reverse map[string][]string
}

// BuildGraph constructs forward and reverse edges from idx using ResolveImport.
func BuildGraph(idx *Index) (*Graph, error) {
	g := &Graph{
		Forward: make(map[string][]string, len(idx.Files)),
		Reverse: make(map[string][]string, len(idx.Files)),
	}
	for filePath, fi := range idx.Files {
		if fi == nil {
			continue
		}
		seen := make(map[string]bool)
		var targets []string
		for _, imp := range fi.Imports {
			resolved, err := idx.ResolveImport(imp)
			if err != nil {
				return nil, err
			}
			for _, t := range resolved {
				if t == filePath || seen[t] {
					continue
				}
				seen[t] = true
				targets = append(targets, t)
			}
		}
		slices.Sort(targets)
		g.Forward[filePath] = targets
		for _, t := range targets {
			g.Reverse[t] = append(g.Reverse[t], filePath)
		}
	}
	for k := range g.Reverse {
		slices.Sort(g.Reverse[k])
		g.Reverse[k] = slices.Compact(g.Reverse[k])
	}
	return g, nil
}

// BFS returns the shortest import-hop distance from seed to each reachable file,
// limited to maxDepth hops. Seed itself is depth 0.
func (g *Graph) BFS(seed string, maxDepth int) map[string]int {
	dist := map[string]int{seed: 0}
	if maxDepth < 0 {
		return dist
	}
	queue := []string{seed}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		d := dist[cur]
		if d >= maxDepth {
			continue
		}
		for _, next := range g.Forward[cur] {
			if _, ok := dist[next]; ok {
				continue
			}
			dist[next] = d + 1
			queue = append(queue, next)
		}
	}
	return dist
}

// samePackageSiblings returns other indexed files in the same directory as seed
// (Go package co-files), excluding seed itself.
func samePackageSiblings(idx *Index, seed string) []string {
	seedDir := path.Clean(filepath.ToSlash(filepath.Dir(seed)))
	var out []string
	for p := range idx.Files {
		if p == seed {
			continue
		}
		if path.Clean(filepath.ToSlash(filepath.Dir(p))) == seedDir {
			out = append(out, p)
		}
	}
	slices.Sort(out)
	return out
}
