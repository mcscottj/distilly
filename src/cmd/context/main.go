// Command distilly-context selects relevant Go files for LLM context from a
// repo seed file and question. Output formats: report (default), json, markdown.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"distilly/internal/context"
	"distilly/internal/version"
)

const defaultMaxTokens = 32000

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	repo := flag.String("repo", ".", "repository root")
	seed := flag.String("seed", "", "repo-relative seed file path (required)")
	question := flag.String("question", "", "question guiding file selection")
	maxDepth := flag.Int("max-depth", 0, "max import graph depth (default 2)")
	maxFiles := flag.Int("max-files", 0, "max files to include (default 20)")
	maxTokens := flag.Int("max-tokens", defaultMaxTokens, "token budget (0 = no cap)")
	includeTests := flag.Bool("include-tests", false, "include *_test.go files")
	format := flag.String("format", "report", "output format: report, json, or markdown")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: distilly-context [-version] -seed <path> [flags]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  distilly-context -version")
		fmt.Fprintln(os.Stderr, "  distilly-context -repo . -seed internal/lint/apply.go -question \"why reject streaming?\"")
		fmt.Fprintln(os.Stderr, "  distilly-context -repo . -seed internal/lint/apply.go -format json")
		fmt.Fprintln(os.Stderr, "  distilly-context -repo . -seed internal/lint/apply.go -format markdown")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		return
	}

	if strings.TrimSpace(*seed) == "" {
		flag.Usage()
		os.Exit(1)
	}

	repoRoot, err := filepath.Abs(*repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving repo path: %v\n", err)
		os.Exit(1)
	}

	opts := context.Options{
		RepoRoot:     repoRoot,
		SeedFile:     filepath.ToSlash(*seed),
		Question:     *question,
		MaxDepth:     *maxDepth,
		MaxFiles:     *maxFiles,
		MaxTokens:    *maxTokens,
		IncludeTests: *includeTests,
	}

	result, err := context.Select(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch *format {
	case "report":
		result.Print(os.Stdout)
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "error encoding json: %v\n", err)
			os.Exit(1)
		}
	case "markdown":
		md, err := context.FormatContext(result, repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting markdown: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(md)
	default:
		fmt.Fprintf(os.Stderr, "unknown format %q (want report, json, or markdown)\n", *format)
		os.Exit(1)
	}
}
