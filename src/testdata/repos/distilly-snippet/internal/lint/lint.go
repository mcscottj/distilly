package lint

import "distilly/internal/store"

type Report struct {
	OK bool
}

func Run(prompt string) Report {
	_ = store.Open()
	return Report{OK: true}
}
