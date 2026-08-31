package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfigDefaultsAndValidates(t *testing.T) {
	path := writeTestConfig(t, "services:\n  - name: Demo\n    type: docker\n    container: demo\n")
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Listen != ":8080" || cfg.Services[0].Group != "Other" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	if cfg.Dashboard.RefreshSeconds != 30 || len(cfg.Dashboard.Overview) != 3 {
		t.Fatalf("unexpected dashboard defaults: %#v", cfg.Dashboard)
	}
}

func TestLoadConfigRejectsInvalidServices(t *testing.T) {
	cases := []string{
		"services:\n  - name: Demo\n    type: docker\n",
		"services:\n  - name: Demo\n    type: other\n",
		"services:\n  - name: Demo\n    type: systemd\n    unit: a\n  - name: Demo\n    type: systemd\n    unit: b\n",
		"unknown: true\n",
		"dashboard:\n  refresh_seconds: 2\nservices: []\n",
		"dashboard:\n  background: green\nservices: []\n",
		"dashboard:\n  overview: [cpu, weather]\nservices: []\n",
		"services:\n  - name: Demo\n    icon: javascript:alert(1)\n    type: docker\n    container: demo\n",
	}
	for _, contents := range cases {
		if _, err := loadConfig(writeTestConfig(t, contents)); err == nil {
			t.Errorf("expected error for %q", strings.ReplaceAll(contents, "\n", " "))
		}
	}
}

func TestSaveConfigIsAtomicAndReloadable(t *testing.T) {
	path := writeTestConfig(t, "services: []\n")
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Dashboard.Background = "#123456"
	cfg.Services = append(cfg.Services, Service{Name: "Demo", Icon: "jellyfin", Type: "docker", Container: "demo"})
	if _, err := saveConfig(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Dashboard.Background != "#123456" || len(loaded.Services) != 1 || loaded.Services[0].Icon != "jellyfin" {
		t.Fatalf("saved config = %#v", loaded)
	}
}
