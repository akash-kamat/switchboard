//go:build darwin

package host

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/akash-kamat/switchboard/internal/config"
	"github.com/akash-kamat/switchboard/internal/platform"
)

type launchdBackend struct{}

func NewNativeBackend() platform.ServiceBackend { return launchdBackend{} }

func launchctl(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "/bin/launchctl", args...).CombinedOutput()
	if ctx.Err() != nil {
		return "", fmt.Errorf("launchctl timed out")
	}
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("launchctl %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func launchTarget(label string) string { return "system/" + label }

func (launchdBackend) State(item config.Service) (platform.ServiceState, error) {
	state := platform.ServiceState{Service: item, Status: "inactive"}
	out, err := launchctl("print", launchTarget(item.Unit))
	if err != nil {
		return state, err
	}
	state.Autostart = true
	for _, line := range strings.Split(out, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), " = ")
		if ok && key == "state" {
			if value == "running" {
				state.Status = "active"
			} else {
				state.Status = value
			}
		}
	}
	return state, nil
}

func (launchdBackend) Action(item config.Service, action string) error {
	target := launchTarget(item.Unit)
	switch action {
	case "start":
		_, err := launchctl("kickstart", target)
		return err
	case "stop":
		_, err := launchctl("kill", "SIGTERM", target)
		return err
	case "restart":
		_, err := launchctl("kickstart", "-k", target)
		return err
	default:
		return fmt.Errorf("unsupported launchd action %q", action)
	}
}

func (launchdBackend) SetAutostart(item config.Service, enabled bool) error {
	action := "disable"
	if enabled {
		action = "enable"
	}
	_, err := launchctl(action, launchTarget(item.Unit))
	return err
}
