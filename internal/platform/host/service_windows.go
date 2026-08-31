//go:build windows

package host

import (
	"fmt"
	"time"

	"github.com/akash-kamat/switchboard/internal/config"
	"github.com/akash-kamat/switchboard/internal/platform"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type windowsServiceBackend struct{}

func NewNativeBackend() platform.ServiceBackend { return windowsServiceBackend{} }

func openWindowsService(name string) (*mgr.Mgr, *mgr.Service, error) {
	manager, err := mgr.Connect()
	if err != nil {
		return nil, nil, fmt.Errorf("connect to Windows Service Manager: %w", err)
	}
	service, err := manager.OpenService(name)
	if err != nil {
		manager.Disconnect()
		return nil, nil, fmt.Errorf("open Windows service %q: %w", name, err)
	}
	return manager, service, nil
}

func windowsState(state svc.State) string {
	switch state {
	case svc.Running:
		return "active"
	case svc.Stopped:
		return "inactive"
	case svc.StartPending, svc.ContinuePending:
		return "activating"
	case svc.StopPending, svc.PausePending:
		return "deactivating"
	case svc.Paused:
		return "paused"
	default:
		return "unknown"
	}
}

func (windowsServiceBackend) State(item config.Service) (platform.ServiceState, error) {
	state := platform.ServiceState{Service: item, Status: "unknown"}
	manager, service, err := openWindowsService(item.Unit)
	if err != nil {
		return state, err
	}
	defer manager.Disconnect()
	defer service.Close()
	status, err := service.Query()
	if err != nil {
		return state, fmt.Errorf("query Windows service %q: %w", item.Unit, err)
	}
	settings, configErr := service.Config()
	if configErr != nil {
		return state, fmt.Errorf("read Windows service %q config: %w", item.Unit, configErr)
	}
	state.Status = windowsState(status.State)
	state.Autostart = settings.StartType == windows.SERVICE_AUTO_START
	return state, nil
}

func waitWindowsState(service *mgr.Service, wanted svc.State) error {
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		status, err := service.Query()
		if err != nil {
			return err
		}
		if status.State == wanted {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for service state %d", wanted)
}

func (windowsServiceBackend) Action(item config.Service, action string) error {
	manager, service, err := openWindowsService(item.Unit)
	if err != nil {
		return err
	}
	defer manager.Disconnect()
	defer service.Close()
	switch action {
	case "start":
		return service.Start()
	case "stop":
		if _, err := service.Control(svc.Stop); err != nil {
			return err
		}
		return waitWindowsState(service, svc.Stopped)
	case "restart":
		if _, err := service.Control(svc.Stop); err != nil {
			return err
		}
		if err := waitWindowsState(service, svc.Stopped); err != nil {
			return err
		}
		return service.Start()
	default:
		return fmt.Errorf("unsupported Windows service action %q", action)
	}
}

func (windowsServiceBackend) SetAutostart(item config.Service, enabled bool) error {
	manager, service, err := openWindowsService(item.Unit)
	if err != nil {
		return err
	}
	defer manager.Disconnect()
	defer service.Close()
	settings, err := service.Config()
	if err != nil {
		return err
	}
	settings.StartType = windows.SERVICE_DEMAND_START
	if enabled {
		settings.StartType = windows.SERVICE_AUTO_START
	}
	return service.UpdateConfig(settings)
}
