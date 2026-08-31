//go:build !linux

package main

import "fmt"

type systemMetrics struct{ diskPath string }

func newSystemMetrics(path string) *systemMetrics { return &systemMetrics{diskPath: path} }

func readCPUStat() (uint64, uint64, error) { return 0, 0, fmt.Errorf("Linux /proc is required") }

func (m *systemMetrics) Stats() (SystemStats, error) {
	return SystemStats{}, fmt.Errorf("system metrics are supported on Linux")
}
