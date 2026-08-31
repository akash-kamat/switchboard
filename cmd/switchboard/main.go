package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/akash-kamat/switchboard/internal/server"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return runServeCommand(args, stderr)
	}

	switch args[0] {
	case "serve":
		return runServeCommand(args[1:], stderr)
	case "version":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "usage: switchboard version")
			return 2
		}
		fmt.Fprintf(stdout, "switchboard %s (commit %s, built %s, %s/%s, %s)\n", version, commit, buildDate, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return 0
	case "validate-config":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: switchboard validate-config <path>")
			return 2
		}
		if _, err := loadConfig(args[1]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "%s is valid\n", args[1])
		return 0
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  switchboard serve [-config path]
  switchboard version
  switchboard validate-config <path>

For compatibility, "switchboard -config path" is the same as "switchboard serve -config path".`)
}

func runServeCommand(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "config.yaml", "path to YAML configuration")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "serve does not accept positional arguments")
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := serve(ctx, *configPath); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func serve(ctx context.Context, configPath string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           server.New(cfg, newDockerBackend("/var/run/docker.sock"), newSystemdBackend(), newSystemMetrics("/"), configPath),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Switchboard listening on %s", cfg.Listen)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("server: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	}
}
