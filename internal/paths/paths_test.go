package paths

import (
	"path/filepath"
	"testing"
)

func TestRuntimeDefaultsAreUsable(t *testing.T) {
	if DefaultConfig() == "" || DefaultDockerSocket() == "" || DefaultDiskPath() == "" {
		t.Fatal("runtime defaults must not be empty")
	}
	if !filepath.IsAbs(DefaultConfig()) {
		t.Fatalf("default config path must be absolute: %q", DefaultConfig())
	}
}
