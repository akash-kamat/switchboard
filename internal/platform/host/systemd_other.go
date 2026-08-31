//go:build !linux

package host

import (
	"fmt"
	"runtime"

	"github.com/akash-kamat/switchboard/internal/config"
	"github.com/akash-kamat/switchboard/internal/platform"
)

type unsupportedNativeBackend struct{}

// NewNativeBackend reports native service control as unavailable until the
// Windows Service and launchd adapters are implemented.
func NewNativeBackend() platform.ServiceBackend { return unsupportedNativeBackend{} }

func (unsupportedNativeBackend) State(service config.Service) (platform.ServiceState, error) {
	return platform.ServiceState{Service: service, Status: "unsupported"}, unsupportedNativeError()
}

func (unsupportedNativeBackend) Action(config.Service, string) error { return unsupportedNativeError() }

func (unsupportedNativeBackend) SetAutostart(config.Service, bool) error {
	return unsupportedNativeError()
}

func unsupportedNativeError() error {
	return fmt.Errorf("native service control is not yet supported on %s", runtime.GOOS)
}
