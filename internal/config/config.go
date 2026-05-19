package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all runtime configuration loaded from the environment and config.yaml.
type Config struct {
	AuthToken    string
	BaseURL      string
	AuthMode     string             `yaml:"auth_mode"`     // "bearer" or "basic"
	SyncEndpoint string             `yaml:"sync_endpoint"` // REST path appended to OS_URL
	BuildCommand string             `yaml:"build_command"`
	Extensions   []ExtensionMapping `yaml:"extensions"`
}

// ExtensionMapping links a local C# file path to an OutSystems Extension ID.
type ExtensionMapping struct {
	File        string `yaml:"file"`
	ExtensionID string `yaml:"extension_id"`
}

// Load reads config.yaml from the given path and merges in environment variables.
// Environment variables (OS_AUTH_TOKEN, OS_URL) take precedence.
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Overlay environment variables.
	cfg.AuthToken = os.Getenv("OS_AUTH_TOKEN")
	cfg.BaseURL = os.Getenv("OS_URL")

	if cfg.BuildCommand == "" {
		cfg.BuildCommand = "dotnet build"
	}
	if cfg.AuthMode == "" {
		cfg.AuthMode = "basic"
	}
	if cfg.SyncEndpoint == "" {
		cfg.SyncEndpoint = "/rest/v1/extensions/sync"
	}

	return cfg, nil
}

// LookupExtension finds the extension mapping for a given file path.
// Returns nil if no mapping is found.
func (c *Config) LookupExtension(filePath string) *ExtensionMapping {
	for i := range c.Extensions {
		if c.Extensions[i].File == filePath {
			return &c.Extensions[i]
		}
	}
	return nil
}
