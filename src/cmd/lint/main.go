// Command distilly-lint reads a prompt file and prints a lint report:
// token counts by section, duplicate/repeated content, and estimated
// potential savings. This is the v1, fully deterministic entrypoint —
// no AI calls are made here.
package main

import (
	"flag"
	"fmt"
	"os"

	"distilly/internal/lint"
	"distilly/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	model := flag.String("model", "", "model name to estimate USD cost against (see internal/cost.Table)")
	fix := flag.Bool("fix", false, "print the optimized prompt instead of the lint report (exact duplicates only, unless -approve-near-duplicates/-approve-json-conversion are set)")
	approveNearDuplicates := flag.Bool("approve-near-duplicates", false, "with -fix, also collapse near-duplicate instructions/examples — review the report's near-duplicate diff first")
	approveJSONConversion := flag.Bool("approve-json-conversion", false, "with -fix, also convert detected structured-data lines to JSON — review the report's structured-data diff first")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: distilly-lint [-version] [-model gpt-4] [-fix [-approve-near-duplicates] [-approve-json-conversion]] <prompt-file>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		return
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	path := flag.Arg(0)
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		os.Exit(1)
	}

	if *fix {
		optimized := lint.Apply(string(data), lint.ApplyOptions{
			ApproveNearDuplicates: *approveNearDuplicates,
			ApproveJSONConversion: *approveJSONConversion,
		})
		fmt.Print(optimized)
		return
	}

	report := lint.Run(string(data), *model)
	report.Print(os.Stdout)
}
