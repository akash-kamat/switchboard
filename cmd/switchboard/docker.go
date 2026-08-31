package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type dockerBackend struct {
	client *http.Client
}

func newDockerBackend(socket string) *dockerBackend {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 3 * time.Second}).DialContext(ctx, "unix", socket)
		},
		DisableCompression: true,
	}
	return &dockerBackend{client: &http.Client{Transport: transport, Timeout: 8 * time.Second}}
}

func (d *dockerBackend) request(method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, "http://docker/v1.41"+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("docker socket: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(limited, &e)
		if e.Message == "" {
			e.Message = strings.TrimSpace(string(limited))
		}
		return fmt.Errorf("docker API returned %s: %s", resp.Status, e.Message)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode docker response: %w", err)
		}
	}
	return nil
}

func dockerContainerPath(name, suffix string) string {
	return "/containers/" + url.PathEscape(name) + suffix
}

func (d *dockerBackend) State(s Service) (ServiceState, error) {
	state := ServiceState{Service: s, Status: "unknown"}
	var inspect struct {
		State struct {
			Status string `json:"Status"`
		} `json:"State"`
		HostConfig struct {
			RestartPolicy struct {
				Name string `json:"Name"`
			} `json:"RestartPolicy"`
		} `json:"HostConfig"`
	}
	if err := d.request(http.MethodGet, dockerContainerPath(s.Container, "/json"), nil, &inspect); err != nil {
		return state, err
	}
	state.Status = inspect.State.Status
	policy := inspect.HostConfig.RestartPolicy.Name
	state.Autostart = policy == "always" || policy == "unless-stopped" || policy == "on-failure"
	if state.Status != "running" {
		return state, nil
	}

	var stats struct {
		CPUStats struct {
			CPUUsage struct {
				Total  uint64   `json:"total_usage"`
				PerCPU []uint64 `json:"percpu_usage"`
			} `json:"cpu_usage"`
			System uint64 `json:"system_cpu_usage"`
			Online uint64 `json:"online_cpus"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				Total uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			System uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Stats struct {
				Cache        uint64 `json:"cache"`
				InactiveFile uint64 `json:"inactive_file"`
			} `json:"stats"`
		} `json:"memory_stats"`
	}
	if err := d.request(http.MethodGet, dockerContainerPath(s.Container, "/stats?stream=false"), nil, &stats); err != nil {
		return state, err
	}
	var cpuDelta, systemDelta uint64
	if stats.CPUStats.CPUUsage.Total >= stats.PreCPUStats.CPUUsage.Total {
		cpuDelta = stats.CPUStats.CPUUsage.Total - stats.PreCPUStats.CPUUsage.Total
	}
	if stats.CPUStats.System >= stats.PreCPUStats.System {
		systemDelta = stats.CPUStats.System - stats.PreCPUStats.System
	}
	cores := stats.CPUStats.Online
	if cores == 0 {
		cores = uint64(len(stats.CPUStats.CPUUsage.PerCPU))
	}
	if systemDelta > 0 {
		state.CPU = float64(cpuDelta) / float64(systemDelta) * float64(cores) * 100
	}
	cache := stats.MemoryStats.Stats.InactiveFile
	if cache == 0 {
		cache = stats.MemoryStats.Stats.Cache
	}
	if stats.MemoryStats.Usage >= cache {
		state.Memory = stats.MemoryStats.Usage - cache
	}
	return state, nil
}

func (d *dockerBackend) Action(s Service, action string) error {
	suffix := "/" + action
	if action == "stop" || action == "restart" {
		suffix += "?t=10"
	}
	return d.request(http.MethodPost, dockerContainerPath(s.Container, suffix), nil, nil)
}

func (d *dockerBackend) SetAutostart(s Service, enabled bool) error {
	policy := "no"
	if enabled {
		policy = "always"
	}
	return d.request(http.MethodPost, dockerContainerPath(s.Container, "/update"), map[string]any{
		"RestartPolicy": map[string]any{"Name": policy, "MaximumRetryCount": 0},
	}, nil)
}
