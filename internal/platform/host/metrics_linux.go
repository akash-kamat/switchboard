//go:build linux

package host

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/akash-kamat/switchboard/internal/platform"
)

type SystemStats = platform.SystemStats

type systemMetrics struct {
	diskPath string
	cpuMu    sync.Mutex
	cpuTotal uint64
	cpuIdle  uint64
}

func newSystemMetrics(path string) *systemMetrics {
	total, idle, _ := readCPUStat()
	return &systemMetrics{diskPath: path, cpuTotal: total, cpuIdle: idle}
}

// NewSystemCollector returns the native Linux metrics collector.
func NewSystemCollector(path string) platform.SystemCollector { return newSystemMetrics(path) }

func readCPUStat() (total uint64, idle uint64, err error) {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(strings.SplitN(string(b), "\n", 2)[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, fmt.Errorf("invalid /proc/stat")
	}
	// Linux reports guest and guest_nice after steal, but those times are
	// already included in user and nice. Stop at steal to avoid counting them
	// twice.
	for i, field := range fields[1:] {
		if i > 7 {
			break
		}
		value, parseErr := strconv.ParseUint(field, 10, 64)
		if parseErr != nil {
			return 0, 0, parseErr
		}
		total += value
		if i == 3 || i == 4 {
			idle += value
		}
	}
	return total, idle, nil
}

func (m *systemMetrics) cpuPercent() (float64, error) {
	total, idle, err := readCPUStat()
	if err != nil {
		return 0, err
	}

	m.cpuMu.Lock()
	previousTotal, previousIdle := m.cpuTotal, m.cpuIdle
	m.cpuTotal, m.cpuIdle = total, idle
	m.cpuMu.Unlock()

	if previousTotal == 0 || total <= previousTotal || idle < previousIdle {
		return 0, nil
	}
	totalDelta := total - previousTotal
	idleDelta := idle - previousIdle
	if idleDelta > totalDelta {
		return 0, nil
	}
	return float64(totalDelta-idleDelta) / float64(totalDelta) * 100, nil
}

func meminfo() (map[string]uint64, error) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	values := make(map[string]uint64)
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			value, _ := strconv.ParseUint(fields[1], 10, 64)
			values[strings.TrimSuffix(fields[0], ":")] = value * 1024
		}
	}
	if values["MemTotal"] == 0 {
		return nil, fmt.Errorf("MemTotal not found")
	}
	return values, nil
}

func readFirstFloat(path string) float64 {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	return v
}

func cpuTemperature() float64 {
	paths, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
	for _, path := range paths {
		if value := readFirstFloat(path); value != 0 {
			if value > 1000 {
				value /= 1000
			}
			return value
		}
	}
	return 0
}

func osName() string {
	b, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Linux"
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
		}
	}
	return "Linux"
}

func localIP() string {
	// A UDP dial does not send traffic, but asks Linux which local address it
	// would use for the default route. This avoids accidentally showing a
	// Docker bridge address before eth0/wlan0.
	connection, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 80})
	if err == nil {
		defer connection.Close()
		if address, ok := connection.LocalAddr().(*net.UDPAddr); ok && usableIPv4(address.IP) {
			return address.IP.String()
		}
	}

	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 || virtualInterface(iface.Name) {
			continue
		}
		addresses, _ := iface.Addrs()
		for _, address := range addresses {
			ip, _, parseErr := net.ParseCIDR(address.String())
			if parseErr == nil && usableIPv4(ip) {
				return ip.String()
			}
		}
	}
	return "—"
}

func usableIPv4(ip net.IP) bool {
	return ip.To4() != nil && !ip.IsLoopback() && !ip.IsUnspecified() && !ip.IsLinkLocalUnicast()
}

func virtualInterface(name string) bool {
	name = strings.ToLower(name)
	for _, prefix := range []string{"docker", "br-", "veth", "virbr", "lo"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func readTrimmed(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return "—"
	}
	return strings.TrimSpace(string(b))
}

func (m *systemMetrics) Stats() (SystemStats, error) {
	cpuPercent, err := m.cpuPercent()
	if err != nil {
		return SystemStats{}, err
	}
	memory, err := meminfo()
	if err != nil {
		return SystemStats{}, err
	}
	var disk syscall.Statfs_t
	if err := syscall.Statfs(m.diskPath, &disk); err != nil {
		return SystemStats{}, err
	}
	diskTotal := disk.Blocks * uint64(disk.Bsize)
	diskFree := disk.Bavail * uint64(disk.Bsize)
	uptimeFields := strings.Fields(readTrimmed("/proc/uptime"))
	uptime := 0.0
	if len(uptimeFields) > 0 {
		uptime, _ = strconv.ParseFloat(uptimeFields[0], 64)
	}
	loadFields := strings.Fields(readTrimmed("/proc/loadavg"))
	loadOne := 0.0
	if len(loadFields) > 0 {
		loadOne, _ = strconv.ParseFloat(loadFields[0], 64)
	}
	hostname, _ := os.Hostname()
	stats := SystemStats{
		CPUPercent:  cpuPercent,
		CPUCores:    runtime.NumCPU(),
		MemoryTotal: memory["MemTotal"], MemoryFree: memory["MemAvailable"], MemoryUsed: memory["MemTotal"] - memory["MemAvailable"],
		SwapTotal: memory["SwapTotal"], SwapFree: memory["SwapFree"], SwapUsed: memory["SwapTotal"] - memory["SwapFree"],
		DiskTotal: diskTotal, DiskFree: diskFree, DiskUsed: diskTotal - diskFree,
		Temperature: cpuTemperature(), LoadOne: loadOne, UptimeSeconds: uint64(uptime),
		Hostname: hostname, LocalIP: localIP(), OS: osName(), Kernel: readTrimmed("/proc/sys/kernel/osrelease"), Architecture: runtime.GOARCH,
	}
	return stats, nil
}
