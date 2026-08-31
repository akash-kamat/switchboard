package main

import (
	"bytes"
	"strings"
	"testing"
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
