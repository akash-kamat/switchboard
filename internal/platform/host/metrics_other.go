//go:build !linux && !windows && !darwin

package host

import (
	"fmt"

	"github.com/akash-kamat/switchboard/internal/platform"
)

type SystemStats = platform.SystemStats

type systemMetrics struct{ diskPath string }

func newSystemMetrics(path string) *systemMetrics { return &systemMetrics{diskPath: path} }

// NewSystemCollector reports an explicit unsupported capability until a native
// collector is provided for this operating system.
func NewSystemCollector(path string) platform.SystemCollector { return newSystemMetrics(path) }

func readCPUStat() (uint64, uint64, error) { return 0, 0, fmt.Errorf("Linux /proc is required") }

func (m *systemMetrics) Stats() (SystemStats, error) {
	return SystemStats{}, fmt.Errorf("system metrics are supported on Linux")
}
