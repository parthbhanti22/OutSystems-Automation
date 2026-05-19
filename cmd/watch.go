package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/parth/os-agent/internal/config"
	"github.com/parth/os-agent/internal/sync"
	"github.com/parth/os-agent/internal/watcher"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Start the agentic file watcher loop",
	Long: `Monitors the workspace for file changes and syncs them to OutSystems:
  • schema.json changes → HTTP POST to OutSystems REST endpoint
  • *.cs changes       → Triggers the configured build command`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWatch(workspacePath)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}

func runWatch(root string) error {
	// Resolve to absolute path for cleaner logs.
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	// Verify workspace exists.
	if _, err := os.Stat(absRoot); os.IsNotExist(err) {
		return fmt.Errorf("workspace not found at %s — run 'os-agent init' first", absRoot)
	}

	// Load config.
	configPath := filepath.Join(absRoot, "config.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Print startup summary.
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           os-agent • watch mode             │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Printf("  Workspace:     %s\n", absRoot)
	fmt.Printf("  Build command: %s\n", cfg.BuildCommand)
	fmt.Printf("  Extensions:    %d mapped\n", len(cfg.Extensions))
	if cfg.BaseURL != "" {
		fmt.Printf("  OS URL:        %s\n", cfg.BaseURL)
	} else {
		fmt.Println("  OS URL:        ⚠ not set")
	}
	fmt.Println()

	// Create sync engines.
	metaSyncer := sync.NewMetadataSyncer(cfg)
	logicSyncer := sync.NewLogicSyncer(cfg)

	// Set up context with signal handling for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		fmt.Printf("\n⚡ Received %s, shutting down...\n", sig)
		cancel()
	}()

	// Define event handlers.
	onSchema := func(path string) {
		if err := metaSyncer.Sync(path); err != nil {
			fmt.Fprintf(os.Stderr, "  ❌ Metadata sync error: %v\n", err)
		}
	}

	onLogic := func(path string) {
		if err := logicSyncer.Sync(ctx, path, absRoot); err != nil {
			fmt.Fprintf(os.Stderr, "  ❌ Logic sync error: %v\n", err)
		}
	}

	// Create and start the file watcher.
	w, err := watcher.New(ctx, absRoot, onSchema, onLogic)
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	return w.Start()
}
