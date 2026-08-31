package main

import (
	"bytes"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/akash-kamat/switchboard/internal/config"
)

func TestVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	for _, want := range []string{"switchboard", version, commit, buildDate} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("version output %q does not contain %q", stdout.String(), want)
		}
	}
}

func TestDefaultPortFallsBackAndPersists(t *testing.T) {
	blocker, err := net.Listen("tcp", ":8080")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "in use") {
			// Something outside the test already provides the collision we need.
			blocker = nil
		} else {
			t.Fatal(err)
		}
	}
	if blocker != nil {
		defer blocker.Close()
	}

	path := writeTestConfig(t, "version: 1\nlisten: ':8080'\nservices: []\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := listenWithDefaultFallback(&cfg, path)
	if err != nil {
		t.Fatal(err)
	}
	listener.Close()
	if cfg.Listen == ":8080" {
		t.Fatal("expected a fallback port")
	}
	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(saved), "listen: :808") {
		t.Fatalf("fallback address was not saved: %s", saved)
	}
}

func TestValidateConfigCommand(t *testing.T) {
	path := writeTestConfig(t, "services: []\n")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"validate-config", path}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "is valid") {
		t.Fatalf("stdout = %q", stdout.String())
	}

	badPath := writeTestConfig(t, "unknown: true\n")
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"validate-config", badPath}, &stdout, &stderr); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestCLIUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"unknown"}, {"version", "extra"}, {"validate-config"}, {"serve", "extra"}} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Errorf("run(%q) exit code = %d, want 2", args, code)
		}
	}
}

func TestLoopbackListenDetection(t *testing.T) {
	for _, address := range []string{"127.0.0.1:8080", "[::1]:8080", "localhost:8080"} {
		if !isLoopbackListen(address) {
			t.Errorf("%q should be loopback", address)
		}
	}
	for _, address := range []string{":8080", "0.0.0.0:8080", "192.168.1.2:8080", "bad"} {
		if isLoopbackListen(address) {
			t.Errorf("%q should be exposed", address)
		}
	}
}
