package server

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/akash-kamat/switchboard/internal/config"
	"github.com/akash-kamat/switchboard/internal/platform"
)

type Config = config.Config
type Service = config.Service
type ServiceState = platform.ServiceState
type SystemStats = platform.SystemStats
type serviceBackend = platform.ServiceBackend
type systemCollector = platform.SystemCollector

var marshalConfig = config.Marshal
var parseConfig = config.Parse
var validateConfig = config.Validate
var saveConfig = config.Save

//go:embed web/*
var webFiles embed.FS

type app struct {
	mu         sync.RWMutex
	config     Config
	configPath string
	docker     serviceBackend
	systemd    serviceBackend
	system     systemCollector
	byName     map[string]Service
}

func newApp(cfg Config, docker, systemd serviceBackend, system systemCollector, configPath ...string) *app {
	path := ""
	if len(configPath) > 0 {
		path = configPath[0]
	}
	a := &app{config: cfg, configPath: path, docker: docker, systemd: systemd, system: system}
	a.reindex()
	return a
}

// New constructs the complete HTTP handler, including the embedded dashboard.
func New(cfg config.Config, dockerBackend, nativeBackend platform.ServiceBackend, system platform.SystemCollector, configPath string) http.Handler {
	return newApp(cfg, dockerBackend, nativeBackend, system, configPath).routes()
}

func (a *app) reindex() {
	a.byName = make(map[string]Service, len(a.config.Services))
	for _, service := range a.config.Services {
		a.byName[service.Name] = service
	}
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/services", a.listServices)
	mux.HandleFunc("POST /api/services/", a.serviceAction)
	mux.HandleFunc("GET /api/system", a.systemStats)
	mux.HandleFunc("GET /api/config", a.getConfig)
	mux.HandleFunc("PUT /api/config", a.putConfig)
	mux.HandleFunc("POST /api/config/validate", a.validateConfigYAML)
	assets, _ := fs.Sub(webFiles, "web")
	mux.Handle("GET /", http.FileServer(http.FS(assets)))
	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; img-src 'self' data: http: https:")
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions && !sameOrigin(r) {
			writeError(w, http.StatusForbidden, "cross-origin requests are not allowed")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	return err == nil && u.Host == r.Host && (u.Scheme == "http" || u.Scheme == "https")
}

func (a *app) backend(s Service) serviceBackend {
	if s.Type == "docker" {
		return a.docker
	}
	return a.systemd
}

func (a *app) listServices(w http.ResponseWriter, _ *http.Request) {
	a.mu.RLock()
	services := append([]Service(nil), a.config.Services...)
	a.mu.RUnlock()
	states := make([]ServiceState, len(services))
	var wg sync.WaitGroup
	for i, service := range services {
		i, service := i, service
		wg.Add(1)
		go func() {
			defer wg.Done()
			state, err := a.backend(service).State(service)
			if err != nil {
				if state.Name == "" {
					state.Service = service
				}
				if state.Status == "" {
					state.Status = "unknown"
				}
				state.Error = err.Error()
			}
			states[i] = state
		}()
	}
	wg.Wait()
	writeJSON(w, http.StatusOK, states)
}

func (a *app) serviceAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.EscapedPath(), "/api/services/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	name, err := url.PathUnescape(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service name")
		return
	}
	a.mu.RLock()
	service, ok := a.byName[name]
	a.mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	action := parts[1]
	backend := a.backend(service)
	if action == "autostart" {
		var req struct {
			Enabled *bool `json:"enabled"`
		}
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil || req.Enabled == nil {
			writeError(w, http.StatusBadRequest, "body must be {\"enabled\": true|false}")
			return
		}
		err = backend.SetAutostart(service, *req.Enabled)
	} else {
		if action != "start" && action != "stop" && action != "restart" {
			writeError(w, http.StatusNotFound, "unknown action")
			return
		}
		err = backend.Action(service, action)
	}
	if err != nil {
		log.Printf("%s %q: %v", action, name, err)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *app) systemStats(w http.ResponseWriter, _ *http.Request) {
	stats, err := a.system.Stats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (a *app) getConfig(w http.ResponseWriter, _ *http.Request) {
	a.mu.RLock()
	cfg := a.config
	a.mu.RUnlock()
	b, err := marshalConfig(cfg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"config": cfg, "yaml": string(b)})
}

func (a *app) putConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Config *Config `json:"config"`
		YAML   *string `json:"yaml"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024*1024))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if (req.Config == nil) == (req.YAML == nil) {
		writeError(w, http.StatusBadRequest, "provide exactly one of config or yaml")
		return
	}
	var cfg Config
	var err error
	if req.YAML != nil {
		cfg, err = parseConfig([]byte(*req.YAML))
	} else {
		cfg = *req.Config
		err = validateConfig(&cfg)
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	b, err := saveConfig(a.configPath, cfg)
	if err != nil {
		status := http.StatusInternalServerError
		if a.configPath == "" {
			status = http.StatusConflict
		}
		writeError(w, status, err.Error())
		return
	}
	a.mu.Lock()
	a.config = cfg
	a.reindex()
	a.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"config": cfg, "yaml": string(b)})
}

func (a *app) validateConfigYAML(w http.ResponseWriter, r *http.Request) {
	var req struct {
		YAML string `json:"yaml"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024*1024))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	cfg, err := parseConfig([]byte(req.YAML))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	b, err := marshalConfig(cfg)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"config": cfg, "yaml": string(b)})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil && !errors.Is(err, http.ErrHandlerTimeout) {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
