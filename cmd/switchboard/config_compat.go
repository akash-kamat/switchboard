package main

import "github.com/akash-kamat/switchboard/internal/config"

// These aliases keep the application behavior stable while package boundaries
// are introduced incrementally.
type Config = config.Config
type DashboardConfig = config.DashboardConfig
type Service = config.Service

const currentConfigVersion = config.CurrentVersion

func defaultConfig() Config                              { return config.Default() }
func loadConfig(path string) (Config, error)             { return config.Load(path) }
func parseConfig(data []byte) (Config, error)            { return config.Parse(data) }
func validateConfig(cfg *Config) error                   { return config.Validate(cfg) }
func marshalConfig(cfg Config) ([]byte, error)           { return config.Marshal(cfg) }
func saveConfig(path string, cfg Config) ([]byte, error) { return config.Save(path, cfg) }
