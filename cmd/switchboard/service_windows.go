//go:build windows

package main

import (
	"context"
	"fmt"
	"io"

	"golang.org/x/sys/windows/svc"
)

type windowsServiceHandler struct{ options serveOptions }

func (handler windowsServiceHandler) Execute(_ []string, requests <-chan svc.ChangeRequest, statuses chan<- svc.Status) (bool, uint32) {
	const accepted = svc.AcceptStop | svc.AcceptShutdown
	statuses <- svc.Status{State: svc.StartPending}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- serve(ctx, handler.options.configPath, handler.options.dockerSocket, handler.options.diskPath)
	}()
	statuses <- svc.Status{State: svc.Running, Accepts: accepted}
	for {
		select {
		case err := <-done:
			if err != nil {
				return true, 1
			}
			return false, 0
		case request := <-requests:
			switch request.Cmd {
			case svc.Interrogate:
				statuses <- request.CurrentStatus
			case svc.Stop, svc.Shutdown:
				statuses <- svc.Status{State: svc.StopPending}
				cancel()
				if err := <-done; err != nil {
					return true, 1
				}
				return false, 0
			}
		}
	}
}

func runServiceCommand(args []string, stderr io.Writer) int {
	options, exitCode := parseServeOptions(args, stderr)
	if exitCode != 0 {
		return exitCode
	}
	if err := svc.Run("Switchboard", windowsServiceHandler{options: options}); err != nil {
		fmt.Fprintln(stderr, "Windows service:", err)
		return 1
	}
	return 0
}
