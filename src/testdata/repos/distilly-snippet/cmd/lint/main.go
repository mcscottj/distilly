package main

import (
	"fmt"

	"distilly/internal/lint"
)

func main() {
	r := lint.Run("hello")
	fmt.Println(r.OK)
}
