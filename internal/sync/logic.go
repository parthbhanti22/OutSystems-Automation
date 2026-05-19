package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/parth/os-agent/internal/config"
)

// LogicSyncer handles triggering builds/publishes when .cs files change.
type LogicSyncer struct {
	cfg *config.Config
}

// NewLogicSyncer creates a syncer that invokes the configured build command.
func NewLogicSyncer(cfg *config.Config) *LogicSyncer {
	return &LogicSyncer{cfg: cfg}
}

// Sync executes the build command for a changed .cs file.
// It pipes stdout/stderr directly to the terminal for real-time debugging.
func (l *LogicSyncer) Sync(ctx context.Context, filePath string, workspaceRoot string) error {
	relPath, err := filepath.Rel(workspaceRoot, filePath)
	if err != nil {
		relPath = filePath
	}

	// Look up extension mapping for enhanced logging.
	ext := l.cfg.LookupExtension(relPath)
	if ext != nil {
		fmt.Printf("  🔧 Building extension %s (file: %s)\n", ext.ExtensionID, relPath)
	} else {
		fmt.Printf("  🔧 Building for changed file: %s (no extension mapping)\n", relPath)
	}

	// Split the build command into program + args.
	parts := strings.Fields(l.cfg.BuildCommand)
	if len(parts) == 0 {
		return fmt.Errorf("build_command is empty in config.yaml")
	}

	//nolint:gosec // Build command is intentionally user-configured.
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = workspaceRoot

	// Pipe stdout/stderr directly to the terminal — zero-copy, no buffering.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set the changed file as an environment variable for the build script.
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("OS_CHANGED_FILE=%s", relPath),
	)
	if ext != nil {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("OS_EXTENSION_ID=%s", ext.ExtensionID),
		)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build command failed (exit: %v): %w", cmd.ProcessState.ExitCode(), err)
	}

	fmt.Printf("  ✅ Build succeeded for %s\n", relPath)
	return nil
}
