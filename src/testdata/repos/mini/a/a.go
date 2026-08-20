package a

import (
	"example.com/mini/b"
)

// CallB forwards to package b.
func CallB() string {
	return b.CallC()
}

type Helper struct {
	Name string
}

const Version = 1

var Shared = "a"
