// Command distilly-lint reads a prompt file and prints a lint report:
// token counts by section, duplicate/repeated content, and estimated
// potential savings. This is the v1, fully deterministic entrypoint —
// no AI calls are made here.
package main

import (
	"fmt"
	"os"

	"github.com/smcguire/distilly/internal/lint"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: distilly-lint <prompt-file>")
		os.Exit(1)
	}

	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		os.Exit(1)
	}

	report := lint.Run(string(data))
	report.Print(os.Stdout)
}
