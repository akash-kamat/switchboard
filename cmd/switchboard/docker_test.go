package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func dockerResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestDockerState(t *testing.T) {
	d := &dockerBackend{client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/json"):
			return dockerResponse(`{"State":{"Status":"running"},"HostConfig":{"RestartPolicy":{"Name":"always"}}}`), nil
		case strings.HasSuffix(r.URL.Path, "/stats") && r.URL.Query().Get("stream") == "false":
			return dockerResponse(`{"cpu_stats":{"cpu_usage":{"total_usage":300,"percpu_usage":[1,1,1,1]},"system_cpu_usage":2000,"online_cpus":4},"precpu_stats":{"cpu_usage":{"total_usage":100},"system_cpu_usage":1000},"memory_stats":{"usage":1000,"stats":{"inactive_file":100}}}`), nil
		default:
			t.Fatalf("unexpected Docker request: %s", r.URL.String())
			return nil, nil
		}
	})}}
	state, err := d.State(Service{Name: "Demo", Type: "docker", Container: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != "running" || !state.Autostart || state.CPU != 80 || state.Memory != 900 {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestDockerActionAndAutostartRequests(t *testing.T) {
	var requests []*http.Request
	d := &dockerBackend{client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests = append(requests, r)
		return dockerResponse(`{}`), nil
	})}}
	s := Service{Container: "demo"}
	if err := d.Action(s, "restart"); err != nil {
		t.Fatal(err)
	}
	if err := d.SetAutostart(s, false); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 2 || requests[0].URL.Path != "/v1.41/containers/demo/restart" || requests[0].URL.Query().Get("t") != "10" {
		t.Fatalf("unexpected action request: %#v", requests)
	}
	if requests[1].URL.Path != "/v1.41/containers/demo/update" {
		t.Fatalf("unexpected update path: %s", requests[1].URL.Path)
	}
}
