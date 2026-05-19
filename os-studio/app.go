package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/parth/os-agent/internal/config"
	"github.com/parth/os-agent/internal/sync"
)

// App struct
type App struct {
	ctx context.Context
}

// GeneratePayload represents the data sent from the UI
type GeneratePayload struct {
	Mermaid string `json:"mermaid"`
	JSON    string `json:"json"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GenerateApp parses the input and scaffolds the application
func (a *App) GenerateApp(payload string) (string, error) {
	var data GeneratePayload
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return "", fmt.Errorf("failed to parse input payload: %w", err)
	}

	if data.JSON == "" && data.Mermaid != "" {
		return "", fmt.Errorf("mermaid to JSON parsing is not yet supported. Please drop a schema.json file")
	}

	if data.JSON == "" {
		return "", fmt.Errorf("no schema.json provided")
	}

	// We are running inside os-studio, so the workspace is at ../my-os-workspace
	workspaceRoot, err := filepath.Abs("../my-os-workspace")
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	schemaPath := filepath.Join(workspaceRoot, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(data.JSON), 0644); err != nil {
		return "", fmt.Errorf("failed to write schema.json: %w", err)
	}

	// Scaffold C# Extension File
	if err := ScaffoldCSharp(data.JSON, workspaceRoot); err != nil {
		return "", fmt.Errorf("failed to scaffold C# logic: %w", err)
	}

	// Load .env explicitly
	envPath := filepath.Join(workspaceRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("Warning: Could not load %s, continuing with system env vars\n", envPath)
	}

	// Load configuration
	configPath := filepath.Join(workspaceRoot, "config.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Trigger the MetadataSyncer
	syncer := sync.NewMetadataSyncer(cfg)
	if err := syncer.Sync(schemaPath); err != nil {
		return "", fmt.Errorf("outSystems sync failed: %w", err)
	}

	return fmt.Sprintf("Application successfully synced to %s%s", cfg.BaseURL, cfg.SyncEndpoint), nil
}
