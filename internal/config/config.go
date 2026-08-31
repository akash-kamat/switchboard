package config

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

var allowedOverviewMetrics = map[string]bool{
	"cpu": true, "memory": true, "storage": true, "temperature": true, "load": true, "swap": true,
}

var allowedSystemDetails = map[string]bool{
	"hostname": true, "local_ip": true, "os": true, "uptime": true, "kernel": true, "architecture": true,
}

const currentConfigVersion = 1

// CurrentVersion is the newest configuration schema understood by this build.
const CurrentVersion = currentConfigVersion

type Config struct {
	Version   int             `yaml:"version" json:"version"`
	Listen    string          `yaml:"listen" json:"listen"`
	Dashboard DashboardConfig `yaml:"dashboard,omitempty" json:"dashboard"`
	Services  []Service       `yaml:"services" json:"services"`
}

type DashboardConfig struct {
	RefreshSeconds int      `yaml:"refresh_seconds,omitempty" json:"refreshSeconds"`
	Theme          string   `yaml:"theme,omitempty" json:"theme"`
	Background     string   `yaml:"background,omitempty" json:"background"`
	Overview       []string `yaml:"overview,omitempty" json:"overview"`
	SystemDetails  []string `yaml:"system_details,omitempty" json:"systemDetails"`
}

type Service struct {
	Name            string `yaml:"name" json:"name"`
	Icon            string `yaml:"icon,omitempty" json:"icon,omitempty"`
	Type            string `yaml:"type" json:"type"`
	Unit            string `yaml:"unit,omitempty" json:"unit,omitempty"`
	Container       string `yaml:"container,omitempty" json:"container,omitempty"`
	Href            string `yaml:"href,omitempty" json:"href"`
	Description     string `yaml:"description,omitempty" json:"description"`
	Group           string `yaml:"group,omitempty" json:"group"`
	ConfigAutostart *bool  `yaml:"autostart,omitempty" json:"autostart,omitempty"`
}

func defaultConfig() Config {
	return Config{
		Version: 1,
		Listen:  ":8080",
		Dashboard: DashboardConfig{
			RefreshSeconds: 30,
			Theme:          "light",
			Background:     "#4ec669",
			Overview:       []string{"cpu", "memory", "storage"},
			SystemDetails:  []string{"hostname", "local_ip", "os", "uptime", "kernel"},
		},
		Services: []Service{},
	}
}

func loadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	return parseConfig(b)
}

func parseConfig(b []byte) (Config, error) {
	cfg := defaultConfig()
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := validateConfig(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Version != currentConfigVersion {
		return fmt.Errorf("config version must be %d (got %d)", currentConfigVersion, cfg.Version)
	}
	cfg.Listen = strings.TrimSpace(cfg.Listen)
	if cfg.Listen == "" {
		cfg.Listen = ":8080"
	}
	if cfg.Dashboard.RefreshSeconds == 0 {
		cfg.Dashboard.RefreshSeconds = 30
	}
	if cfg.Dashboard.RefreshSeconds < 5 || cfg.Dashboard.RefreshSeconds > 3600 {
		return fmt.Errorf("dashboard.refresh_seconds must be between 5 and 3600")
	}
	cfg.Dashboard.Theme = strings.ToLower(strings.TrimSpace(cfg.Dashboard.Theme))
	if cfg.Dashboard.Theme == "" {
		cfg.Dashboard.Theme = "light"
	}
	if cfg.Dashboard.Theme != "light" && cfg.Dashboard.Theme != "dark" && cfg.Dashboard.Theme != "system" {
		return fmt.Errorf("dashboard.theme must be light, dark, or system")
	}
	if cfg.Dashboard.Background == "" {
		cfg.Dashboard.Background = "#4ec669"
	}
	if !hexColor.MatchString(cfg.Dashboard.Background) {
		return fmt.Errorf("dashboard.background must be a six-digit hex color such as #4ec669")
	}
	if err := validateChoices("dashboard.overview", cfg.Dashboard.Overview, allowedOverviewMetrics); err != nil {
		return err
	}
	if err := validateChoices("dashboard.system_details", cfg.Dashboard.SystemDetails, allowedSystemDetails); err != nil {
		return err
	}

	seen := make(map[string]bool)
	for i := range cfg.Services {
		s := &cfg.Services[i]
		s.Name = strings.TrimSpace(s.Name)
		s.Icon = strings.TrimSpace(s.Icon)
		s.Type = strings.ToLower(strings.TrimSpace(s.Type))
		s.Unit = strings.TrimSpace(s.Unit)
		s.Container = strings.TrimSpace(s.Container)
		s.Group = strings.TrimSpace(s.Group)
		s.Href = strings.TrimSpace(s.Href)
		if s.Group == "" {
			s.Group = "Other"
		}
		if s.Name == "" {
			return fmt.Errorf("services[%d]: name is required", i)
		}
		if seen[s.Name] {
			return fmt.Errorf("services[%d]: duplicate name %q", i, s.Name)
		}
		seen[s.Name] = true
		if s.Href != "" {
			u, err := url.Parse(s.Href)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
				return fmt.Errorf("service %q: href must be an http or https URL", s.Name)
			}
		}
		if s.Icon != "" && s.Icon != "auto" {
			if u, err := url.Parse(s.Icon); err == nil && u.IsAbs() {
				if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
					return fmt.Errorf("service %q: icon URL must use http or https", s.Name)
				}
			} else if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{1,64}$`, s.Icon); !matched {
				return fmt.Errorf("service %q: icon must be auto, a simple icon name, or an http/https URL", s.Name)
			}
		}
		switch s.Type {
		case "docker":
			if s.Container == "" {
				return fmt.Errorf("service %q: container is required", s.Name)
			}
		case "systemd":
			if s.Unit == "" {
				return fmt.Errorf("service %q: unit is required", s.Name)
			}
		default:
			return fmt.Errorf("service %q: type must be docker or systemd", s.Name)
		}
	}
	return nil
}

func validateChoices(field string, choices []string, allowed map[string]bool) error {
	seen := make(map[string]bool)
	for i, choice := range choices {
		choice = strings.ToLower(strings.TrimSpace(choice))
		choices[i] = choice
		if !allowed[choice] {
			return fmt.Errorf("%s contains unsupported value %q", field, choice)
		}
		if seen[choice] {
			return fmt.Errorf("%s contains duplicate value %q", field, choice)
		}
		seen[choice] = true
	}
	return nil
}

func marshalConfig(cfg Config) ([]byte, error) {
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode config: %w", err)
	}
	return b, nil
}

func saveConfig(path string, cfg Config) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("configuration is read-only in this run")
	}
	b, err := marshalConfig(cfg)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".switchboard-config-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary config: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0644); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("set config permissions: %w", err)
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("write config: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("sync config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("close config: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return nil, fmt.Errorf("replace config: %w", err)
	}
	return b, nil
}

// Default returns a configuration populated with safe application defaults.
func Default() Config { return defaultConfig() }

// Load reads, parses, defaults, and validates a configuration file.
func Load(path string) (Config, error) { return loadConfig(path) }

// Parse parses, defaults, and validates YAML configuration data.
func Parse(data []byte) (Config, error) { return parseConfig(data) }

// Validate normalizes and validates a configuration value in place.
func Validate(cfg *Config) error { return validateConfig(cfg) }

// Marshal validates and returns normalized YAML configuration data.
func Marshal(cfg Config) ([]byte, error) { return marshalConfig(cfg) }

// Save validates and atomically replaces the configuration file.
func Save(path string, cfg Config) ([]byte, error) { return saveConfig(path, cfg) }
