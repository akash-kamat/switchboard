//go:build !windows

package main

import (
	"fmt"
	"io"
)

func runServiceCommand(_ []string, stderr io.Writer) int {
	fmt.Fprintln(stderr, "the service command is only used by Windows Service Control Manager")
	return 2
}
