package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

type fakeBackend struct {
	mu      sync.Mutex
	actions []string
	auto    []bool
	err     error
}

func (f *fakeBackend) State(s Service) (ServiceState, error) {
	return ServiceState{Service: s, Status: "running", Autostart: true, CPU: 2.5, Memory: 1024}, f.err
}
func (f *fakeBackend) Action(_ Service, action string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.actions = append(f.actions, action)
	return f.err
}
func (f *fakeBackend) SetAutostart(_ Service, enabled bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.auto = append(f.auto, enabled)
	return f.err
}

type fakeSystem struct{}

func (fakeSystem) Stats() (SystemStats, error) { return SystemStats{CPUPercent: 10}, nil }

func testApp() (*app, *fakeBackend) {
	b := &fakeBackend{}
	cfg := defaultConfig()
	cfg.Services = []Service{{Name: "My Service", Type: "docker", Container: "demo", Group: "Test"}}
	return newApp(cfg, b, b, fakeSystem{}), b
}

func TestListServices(t *testing.T) {
	a, _ := testApp()
	r := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var states []ServiceState
	if err := json.NewDecoder(w.Body).Decode(&states); err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 || states[0].Name != "My Service" || states[0].Status != "running" {
		t.Fatalf("unexpected response: %#v", states)
	}
	if got := w.Header().Get("Content-Security-Policy"); got == "" {
		t.Error("missing CSP header")
	}
}

func TestActionsOnlyTargetConfiguredServices(t *testing.T) {
	a, backend := testApp()
	for _, tc := range []struct {
		path string
		want int
	}{
		{"/api/services/My%20Service/restart", http.StatusOK},
		{"/api/services/Unknown/restart", http.StatusNotFound},
		{"/api/services/My%20Service/delete", http.StatusNotFound},
	} {
		w := httptest.NewRecorder()
		a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodPost, tc.path, nil))
		if w.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.path, w.Code, tc.want)
		}
	}
	if len(backend.actions) != 1 || backend.actions[0] != "restart" {
		t.Fatalf("actions = %#v", backend.actions)
	}
}

func TestAutostartRequiresExplicitBoolean(t *testing.T) {
	a, backend := testApp()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/services/My%20Service/autostart", strings.NewReader(`{"enabled":false}`))
	a.routes().ServeHTTP(w, r)
	if w.Code != http.StatusOK || len(backend.auto) != 1 || backend.auto[0] {
		t.Fatalf("status=%d auto=%v", w.Code, backend.auto)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/api/services/My%20Service/autostart", strings.NewReader(`{}`))
	a.routes().ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestBackendErrorIsReported(t *testing.T) {
	a, backend := testApp()
	backend.err = errors.New("unavailable")
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/services/My%20Service/start", nil))
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestCrossOriginMutationIsRejected(t *testing.T) {
	a, backend := testApp()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://switchboard.local/api/services/My%20Service/start", nil)
	r.Header.Set("Origin", "https://attacker.example")
	a.routes().ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}
	if len(backend.actions) != 0 {
		t.Fatalf("unexpected actions: %v", backend.actions)
	}
}

func TestEmbeddedDashboardIsServed(t *testing.T) {
	a, _ := testApp()
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Switchboard") {
		t.Fatalf("status=%d body=%q", w.Code, w.Body.String())
	}
}

func TestEmbeddedImagesAreServed(t *testing.T) {
	a, _ := testApp()
	for _, path := range []string{"/favicon.png", "/github.png"} {
		w := httptest.NewRecorder()
		a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/png" {
			t.Fatalf("%s: status=%d content-type=%q", path, w.Code, w.Header().Get("Content-Type"))
		}
	}
}

func TestEmbeddedIconSpriteIsServed(t *testing.T) {
	a, _ := testApp()
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/icons.svg", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Header().Get("Content-Type"), "image/svg+xml") || !strings.Contains(w.Body.String(), `id="jellyfin"`) {
		t.Fatalf("status=%d content-type=%q", w.Code, w.Header().Get("Content-Type"))
	}
}

func TestConfigValidationDoesNotWrite(t *testing.T) {
	a, _ := testApp()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/config/validate", strings.NewReader(`{"yaml":"dashboard:\n  refresh_seconds: 1\nservices: []\n"}`))
	a.routes().ServeHTTP(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestSystemAPIContract(t *testing.T) {
	a, _ := testApp()
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/system", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"cpuPercent", "cpuCores", "memoryUsedBytes", "memoryFreeBytes", "memoryTotalBytes", "swapUsedBytes", "swapFreeBytes", "swapTotalBytes", "diskUsedBytes", "diskFreeBytes", "diskTotalBytes", "temperatureCelsius", "loadOne", "uptimeSeconds", "hostname", "localIp", "os", "kernel", "architecture"} {
		if _, ok := body[field]; !ok {
			t.Errorf("response is missing %q", field)
		}
	}
}

func TestConfigAPIContract(t *testing.T) {
	a, _ := testApp()
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/config", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body struct {
		Config Config `json:"config"`
		YAML   string `json:"yaml"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Config.Version != 1 || !strings.Contains(body.YAML, "version: 1") {
		t.Fatalf("unexpected config response: %#v", body)
	}
}

func TestAPIRoutesRejectWrongMethods(t *testing.T) {
	a, _ := testApp()
	for _, path := range []string{"/api/system", "/api/services", "/api/config"} {
		w := httptest.NewRecorder()
		a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodDelete, path, nil))
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("DELETE %s: status = %d, want %d", path, w.Code, http.StatusMethodNotAllowed)
		}
	}
}

func TestConfigUpdatePersistsAndReloadsApp(t *testing.T) {
	path := writeTestConfig(t, "services: []\n")
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	b := &fakeBackend{}
	a := newApp(cfg, b, b, fakeSystem{}, path)
	body := `{"yaml":"dashboard:\n  refresh_seconds: 15\nservices:\n  - name: Added\n    type: docker\n    container: added\n"}`
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	a.mu.RLock()
	_, exists := a.byName["Added"]
	a.mu.RUnlock()
	if !exists {
		t.Fatal("saved service was not loaded into the running app")
	}
	loaded, err := loadConfig(path)
	if err != nil || loaded.Dashboard.RefreshSeconds != 15 {
		t.Fatalf("loaded config = %#v, err = %v", loaded, err)
	}
}

func TestInvalidConfigUpdateLeavesFileUntouched(t *testing.T) {
	path := writeTestConfig(t, "services: []\n")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	b := &fakeBackend{}
	a := newApp(cfg, b, b, fakeSystem{}, path)
	w := httptest.NewRecorder()
	body := `{"yaml":"services:\n  - name: Broken\n    type: docker\n"}`
	a.routes().ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(body)))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("invalid update changed the config file")
	}
}
