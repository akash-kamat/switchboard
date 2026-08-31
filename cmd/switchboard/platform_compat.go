package main

import (
	"github.com/akash-kamat/switchboard/internal/docker"
	"github.com/akash-kamat/switchboard/internal/platform"
)

type ServiceState = platform.ServiceState
type SystemStats = platform.SystemStats
type serviceBackend = platform.ServiceBackend
type systemCollector = platform.SystemCollector

func newDockerBackend(socket string) serviceBackend { return docker.New(socket) }
