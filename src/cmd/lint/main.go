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
)

func main() {
	model := flag.String("model", "", "model name to estimate USD cost against (see internal/cost.Table)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: distilly-lint [-model gpt-4] <prompt-file>")
		flag.PrintDefaults()
	}
	flag.Parse()

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

	report := lint.Run(string(data), *model)
	report.Print(os.Stdout)
}
