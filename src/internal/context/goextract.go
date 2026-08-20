package context

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

//go:embed queries/go.scm
var goQuerySource string

var (
	goLangOnce sync.Once
	goLang     *tree_sitter.Language
	goQuery    *tree_sitter.Query
	goInitErr  error
)

func initGoTreeSitter() {
	goLangOnce.Do(func() {
		goLang = tree_sitter.NewLanguage(tree_sitter_go.Language())
		if goLang == nil {
			goInitErr = fmt.Errorf("tree-sitter-go: failed to load language")
			return
		}
		q, err := tree_sitter.NewQuery(goLang, goQuerySource)
		if err != nil {
			goInitErr = fmt.Errorf("tree-sitter go query: %w", err)
			return
		}
		goQuery = q
	})
}

// FileInfo is extracted metadata for one Go source file.
type FileInfo struct {
	Path    string   // repo-relative path
	Package string   // package clause name
	Imports []string // import paths (unquoted)
	Symbols []string // top-level function/type/var/const names
}

// ExtractGoFile parses Go source with Tree-sitter and returns package, imports, and symbols.
func ExtractGoFile(path string, src []byte) (FileInfo, error) {
	initGoTreeSitter()
	if goInitErr != nil {
		return FileInfo{}, goInitErr
	}

	parser := tree_sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(goLang); err != nil {
		return FileInfo{}, fmt.Errorf("set language: %w", err)
	}

	tree := parser.Parse(src, nil)
	if tree == nil {
		return FileInfo{}, fmt.Errorf("parse %s: nil tree", path)
	}
	defer tree.Close()

	root := tree.RootNode()
	if root == nil {
		return FileInfo{}, fmt.Errorf("parse %s: nil root", path)
	}
	if root.HasError() && !hasPackageClause(src) {
		return FileInfo{}, fmt.Errorf("parse %s: syntax errors and no package clause", path)
	}

	info := FileInfo{Path: path}
	seenImport := make(map[string]bool)
	seenSymbol := make(map[string]bool)

	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	matches := qc.Matches(goQuery, root, src)
	captureNames := goQuery.CaptureNames()

	for match := matches.Next(); match != nil; match = matches.Next() {
		for _, cap := range match.Captures {
			name := captureNames[cap.Index]
			text := strings.TrimSpace(cap.Node.Utf8Text(src))
			switch name {
			case "package":
				if info.Package == "" {
					info.Package = text
				}
			case "import":
				imp := strings.Trim(text, "`\"")
				if imp != "" && !seenImport[imp] {
					seenImport[imp] = true
					info.Imports = append(info.Imports, imp)
				}
			case "symbol":
				if text != "" && !seenSymbol[text] {
					seenSymbol[text] = true
					info.Symbols = append(info.Symbols, text)
				}
			}
		}
	}

	if info.Package == "" {
		return FileInfo{}, fmt.Errorf("parse %s: missing package clause", path)
	}
	return info, nil
}

func hasPackageClause(src []byte) bool {
	// Cheap pre-check used when the tree has errors.
	for _, line := range strings.Split(string(src), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			return true
		}
	}
	return false
}
