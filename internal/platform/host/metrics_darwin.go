//go:build darwin

package host

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akash-kamat/switchboard/internal/platform"
	"golang.org/x/sys/unix"
)

type SystemStats = platform.SystemStats

type systemMetrics struct {
	diskPath          string
	cpuMu             sync.Mutex
	cpuTotal, cpuIdle uint64
}

func NewSystemCollector(path string) platform.SystemCollector {
	total, idle, _ := darwinCPUTimes()
	return &systemMetrics{diskPath: path, cpuTotal: total, cpuIdle: idle}
}

func commandOutput(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func sysctlValue(name string) (string, error) { return commandOutput("/usr/sbin/sysctl", "-n", name) }

func darwinCPUTimes() (uint64, uint64, error) {
	value, err := sysctlValue("kern.cp_time")
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(value)
	if len(fields) < 4 {
		return 0, 0, fmt.Errorf("unexpected kern.cp_time value")
	}
	var total uint64
	for _, field := range fields {
		v, parseErr := strconv.ParseUint(field, 10, 64)
		if parseErr != nil {
			return 0, 0, parseErr
		}
		total += v
	}
	idle, err := strconv.ParseUint(fields[3], 10, 64)
	return total, idle, err
}

func (m *systemMetrics) cpuPercent() (float64, error) {
	total, idle, err := darwinCPUTimes()
	if err != nil {
		return 0, err
	}
	m.cpuMu.Lock()
	oldTotal, oldIdle := m.cpuTotal, m.cpuIdle
	m.cpuTotal, m.cpuIdle = total, idle
	m.cpuMu.Unlock()
	if total <= oldTotal || idle < oldIdle {
		return 0, nil
	}
	delta, idleDelta := total-oldTotal, idle-oldIdle
	if delta == 0 || idleDelta > delta {
		return 0, nil
	}
	return float64(delta-idleDelta) / float64(delta) * 100, nil
}

func uintSysctl(name string) uint64 {
	value, _ := sysctlValue(name)
	parsed, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return parsed
}

func darwinAvailableMemory() uint64 {
	out, err := commandOutput("/usr/bin/vm_stat")
	if err != nil {
		return 0
	}
	pageSize := uint64(4096)
	if start := strings.Index(out, "page size of "); start >= 0 {
		part := out[start+len("page size of "):]
		if end := strings.Index(part, " bytes"); end >= 0 {
			pageSize, _ = strconv.ParseUint(part[:end], 10, 64)
		}
	}
	var pages uint64
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Pages free:") || strings.HasPrefix(line, "Pages inactive:") || strings.HasPrefix(line, "Pages speculative:") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				value, _ := strconv.ParseUint(strings.TrimSuffix(fields[len(fields)-1], "."), 10, 64)
				pages += value
			}
		}
	}
	return pages * pageSize
}

func darwinUptime() uint64 {
	value, err := sysctlValue("kern.boottime")
	if err != nil {
		return 0
	}
	start := strings.Index(value, "sec = ")
	if start < 0 {
		return 0
	}
	rest := value[start+6:]
	end := strings.IndexAny(rest, ", ")
	if end >= 0 {
		rest = rest[:end]
	}
	boot, _ := strconv.ParseInt(rest, 10, 64)
	if boot <= 0 {
		return 0
	}
	return uint64(time.Now().Unix() - boot)
}

func (m *systemMetrics) Stats() (SystemStats, error) {
	cpu, err := m.cpuPercent()
	if err != nil {
		return SystemStats{}, fmt.Errorf("CPU metrics: %w", err)
	}
	var disk unix.Statfs_t
	if err := unix.Statfs(m.diskPath, &disk); err != nil {
		return SystemStats{}, fmt.Errorf("storage metrics: %w", err)
	}
	totalMemory := uintSysctl("hw.memsize")
	freeMemory := darwinAvailableMemory()
	if freeMemory > totalMemory {
		freeMemory = totalMemory
	}
	diskTotal := uint64(disk.Blocks) * uint64(disk.Bsize)
	diskFree := uint64(disk.Bavail) * uint64(disk.Bsize)
	hostname, _ := os.Hostname()
	osName, _ := commandOutput("/usr/bin/sw_vers", "-productName")
	osVersion, _ := commandOutput("/usr/bin/sw_vers", "-productVersion")
	kernel, _ := sysctlValue("kern.osrelease")
	return SystemStats{
		CPUPercent: cpu, CPUCores: runtime.NumCPU(), MemoryTotal: totalMemory, MemoryFree: freeMemory, MemoryUsed: totalMemory - freeMemory,
		DiskTotal: diskTotal, DiskFree: diskFree, DiskUsed: diskTotal - diskFree, UptimeSeconds: darwinUptime(),
		Hostname: hostname, LocalIP: localIP(), OS: strings.TrimSpace(osName + " " + osVersion), Kernel: kernel, Architecture: runtime.GOARCH,
	}, nil
}
