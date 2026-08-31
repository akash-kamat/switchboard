//go:build linux

package host

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/akash-kamat/switchboard/internal/config"
	"github.com/akash-kamat/switchboard/internal/platform"
)

type Service = config.Service
type ServiceState = platform.ServiceState

type systemdBackend struct{}

func newSystemdBackend() *systemdBackend { return &systemdBackend{} }

// NewNativeBackend returns the systemd service adapter on Linux.
func NewNativeBackend() platform.ServiceBackend { return newSystemdBackend() }

func systemctl(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "systemctl", args...).CombinedOutput()
	if ctx.Err() != nil {
		return "", fmt.Errorf("systemctl timed out")
	}
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("systemctl %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (b *systemdBackend) State(s Service) (ServiceState, error) {
	state := ServiceState{Service: s, Status: "unknown"}
	out, err := systemctl("show", s.Unit, "--property=ActiveState", "--property=MainPID", "--no-pager")
	if err != nil {
		return state, err
	}
	var pid int
	for _, line := range strings.Split(out, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		switch key {
		case "ActiveState":
			state.Status = value
		case "MainPID":
			pid, _ = strconv.Atoi(value)
		}
	}
	enabled, enabledErr := systemctl("is-enabled", s.Unit)
	state.Autostart = enabledErr == nil && (enabled == "enabled" || enabled == "enabled-runtime" || enabled == "static")
	if pid > 0 && state.Status == "active" {
		state.Memory, _ = processMemory(pid)
		firstProcess, firstTotal, err1 := processCPU(pid)
		time.Sleep(120 * time.Millisecond)
		secondProcess, secondTotal, err2 := processCPU(pid)
		if err1 == nil && err2 == nil && secondTotal > firstTotal && secondProcess >= firstProcess {
			state.CPU = float64(secondProcess-firstProcess) / float64(secondTotal-firstTotal) * float64(runtime.NumCPU()) * 100
		}
	}
	return state, nil
}

func processCPU(pid int) (uint64, uint64, error) {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, 0, err
	}
	line := string(b)
	end := strings.LastIndex(line, ")")
	if end < 0 {
		return 0, 0, fmt.Errorf("invalid process stat")
	}
	fields := strings.Fields(line[end+1:])
	if len(fields) < 13 {
		return 0, 0, fmt.Errorf("short process stat")
	}
	user, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	system, err := strconv.ParseUint(fields[12], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	total, _, err := readCPUStat()
	return user + system, total, err
}

func processMemory(pid int) (uint64, error) {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "VmRSS:" {
			kb, err := strconv.ParseUint(fields[1], 10, 64)
			return kb * 1024, err
		}
	}
	return 0, fmt.Errorf("VmRSS not found")
}

func (b *systemdBackend) Action(s Service, action string) error {
	_, err := systemctl(action, s.Unit)
	return err
}

func (b *systemdBackend) SetAutostart(s Service, enabled bool) error {
	action := "disable"
	if enabled {
		action = "enable"
	}
	_, err := systemctl(action, s.Unit)
	return err
}
