package main

import "github.com/akash-kamat/switchboard/internal/platform/host"

func newSystemdBackend() serviceBackend            { return host.NewNativeBackend() }
func newSystemMetrics(path string) systemCollector { return host.NewSystemCollector(path) }
