// Package paths defines platform-appropriate runtime defaults. Every default can
// be overridden by the CLI, which keeps packaged and portable installs possible.
package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

func DefaultConfig() string {
	switch runtime.GOOS {
	case "linux":
		return "/etc/switchboard/config.yaml"
	case "darwin":
		return "/Library/Application Support/Switchboard/config.yaml"
	case "windows":
		if root := os.Getenv("ProgramData"); root != "" {
			return filepath.Join(root, "Switchboard", "config.yaml")
		}
	}
	if root, err := os.UserConfigDir(); err == nil {
		return filepath.Join(root, "switchboard", "config.yaml")
	}
	return "config.yaml"
}

func DefaultDockerSocket() string {
	if runtime.GOOS == "windows" {
		return `//./pipe/docker_engine`
	}
	return "/var/run/docker.sock"
}

func DefaultDiskPath() string {
	if runtime.GOOS == "windows" {
		if drive := os.Getenv("SystemDrive"); drive != "" {
			return drive + `\`
		}
		return `C:\`
	}
	return "/"
}
