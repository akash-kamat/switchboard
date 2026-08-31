//go:build windows

package host

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"unsafe"

	"github.com/akash-kamat/switchboard/internal/platform"
	"golang.org/x/sys/windows"
)

type SystemStats = platform.SystemStats

var (
	kernel32           = windows.NewLazySystemDLL("kernel32.dll")
	procGetSystemTimes = kernel32.NewProc("GetSystemTimes")
	procGlobalMemory   = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetTickCount64 = kernel32.NewProc("GetTickCount64")
)

type memoryStatusEx struct {
	Length, MemoryLoad                                 uint32
	TotalPhys, AvailPhys, TotalPageFile, AvailPageFile uint64
	TotalVirtual, AvailVirtual, AvailExtendedVirtual   uint64
}

type systemMetrics struct {
	diskPath          string
	cpuMu             sync.Mutex
	cpuTotal, cpuIdle uint64
}

func NewSystemCollector(path string) platform.SystemCollector {
	total, idle, _ := windowsCPUTimes()
	return &systemMetrics{diskPath: path, cpuTotal: total, cpuIdle: idle}
}

func windowsCPUTimes() (uint64, uint64, error) {
	var idle, kernel, user windows.Filetime
	ok, _, callErr := procGetSystemTimes.Call(uintptr(unsafe.Pointer(&idle)), uintptr(unsafe.Pointer(&kernel)), uintptr(unsafe.Pointer(&user)))
	if ok == 0 {
		return 0, 0, callErr
	}
	idleValue := uint64(idle.Nanoseconds())
	return uint64(kernel.Nanoseconds()) + uint64(user.Nanoseconds()), idleValue, nil
}

func (m *systemMetrics) cpuPercent() (float64, error) {
	total, idle, err := windowsCPUTimes()
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

func windowsMemory() (memoryStatusEx, error) {
	status := memoryStatusEx{Length: uint32(unsafe.Sizeof(memoryStatusEx{}))}
	ok, _, callErr := procGlobalMemory.Call(uintptr(unsafe.Pointer(&status)))
	if ok == 0 {
		return status, callErr
	}
	return status, nil
}

func (m *systemMetrics) Stats() (SystemStats, error) {
	cpu, err := m.cpuPercent()
	if err != nil {
		return SystemStats{}, fmt.Errorf("CPU metrics: %w", err)
	}
	memory, err := windowsMemory()
	if err != nil {
		return SystemStats{}, fmt.Errorf("memory metrics: %w", err)
	}
	diskPath, err := windows.UTF16PtrFromString(m.diskPath)
	if err != nil {
		return SystemStats{}, err
	}
	var available, total, free uint64
	if err := windows.GetDiskFreeSpaceEx(diskPath, &available, &total, &free); err != nil {
		return SystemStats{}, fmt.Errorf("storage metrics: %w", err)
	}
	uptimeMS, _, _ := procGetTickCount64.Call()
	hostname, _ := os.Hostname()
	version := windows.RtlGetVersion()
	return SystemStats{
		CPUPercent: cpu, CPUCores: runtime.NumCPU(),
		MemoryTotal: memory.TotalPhys, MemoryFree: memory.AvailPhys, MemoryUsed: memory.TotalPhys - memory.AvailPhys,
		SwapTotal: memory.TotalPageFile, SwapFree: memory.AvailPageFile, SwapUsed: memory.TotalPageFile - memory.AvailPageFile,
		DiskTotal: total, DiskFree: available, DiskUsed: total - available,
		UptimeSeconds: uint64(uptimeMS) / 1000, Hostname: hostname, LocalIP: localIP(),
		OS:     fmt.Sprintf("Windows %d.%d", version.MajorVersion, version.MinorVersion),
		Kernel: fmt.Sprintf("%d.%d.%d", version.MajorVersion, version.MinorVersion, version.BuildNumber), Architecture: runtime.GOARCH,
	}, nil
}
