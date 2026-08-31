//go:build windows

package docker

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Microsoft/go-winio"
)

func newDockerBackend(pipe string) *dockerBackend {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return winio.DialPipeContext(ctx, pipe)
		},
		DisableCompression: true,
	}
	return &dockerBackend{client: &http.Client{Transport: transport, Timeout: 15 * time.Second}}
}
