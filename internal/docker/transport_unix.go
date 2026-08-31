//go:build !windows

package docker

import (
	"context"
	"net"
	"net/http"
	"time"
)

func newDockerBackend(socket string) *dockerBackend {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 3 * time.Second}).DialContext(ctx, "unix", socket)
		},
		DisableCompression: true,
	}
	return &dockerBackend{client: &http.Client{Transport: transport, Timeout: 15 * time.Second}}
}
