package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold the workspace directory structure",
	Long:  `Creates the workspace with /logic, /ui, schema.json, config.yaml, .env, and .gitignore.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("⚡ Initializing os-agent workspace...")
		return scaffoldWorkspace(workspacePath)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func scaffoldWorkspace(root string) error {
	// Directories to create.
	dirs := []string{
		filepath.Join(root, "logic"),
		filepath.Join(root, "ui"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
		fmt.Printf("  📁 %s\n", d)
	}

	// Files to create (path → content).
	files := map[string]string{
		filepath.Join(root, "schema.json"): schemaJSON,
		filepath.Join(root, "config.yaml"): configYAML,
		filepath.Join(root, ".env"):        dotEnv,
		filepath.Join(root, ".env.example"): dotEnvExample,
		filepath.Join(root, ".gitignore"):  gitignoreContent,
	}
	for path, content := range files {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ⏭  %s (already exists, skipping)\n", path)
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
		fmt.Printf("  📄 %s\n", path)
	}

	fmt.Println()
	fmt.Println("✅ Workspace ready!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %s/.env with your OutSystems credentials\n", root)
	fmt.Printf("  2. Configure extension mappings in %s/config.yaml\n", root)
	fmt.Printf("  3. Run: os-agent watch -w %s\n", root)
	return nil
}

// --- Template content for scaffolded files ---

const schemaJSON = `{
  "name": "MyOutSystemsExtension",
  "description": "Extension schema managed by os-agent",
  "version": "1.0.0",
  "entities": [],
  "structures": [],
  "actions": []
}
`

const configYAML = `# os-agent configuration
# Maps local C# files to OutSystems Extension IDs.

# Authentication mode for the REST API
# Use "basic" for Personal Environments (username:password in .env)
# Use "bearer" for Enterprise Environments (Service Account token in .env)
auth_mode: "basic"

# The REST endpoint path to sync schema.json to
sync_endpoint: "/rest/v1/extensions/sync"

# The command to run when a .cs file changes.
# This is passed to os/exec, so use the full path or ensure it's on your PATH.
# Examples:
#   build_command: "dotnet build"
#   build_command: "msbuild /p:Configuration=Release"
#   build_command: "IntegrationStudio.exe /publish"
build_command: "dotnet build"

# Extension mappings: link local .cs files to OutSystems extension IDs.
extensions:
  # - file: "logic/MyExtension.cs"
  #   extension_id: "ext-abc-123"
  # - file: "logic/AnotherExt.cs"
  #   extension_id: "ext-def-456"
`

const dotEnv = `# OutSystems credentials — DO NOT COMMIT THIS FILE.
# For Personal Environments (basic auth): OS_AUTH_TOKEN=your_username:your_password
# For Enterprise Environments (bearer auth): OS_AUTH_TOKEN=your_service_account_token
OS_AUTH_TOKEN=your_username:your_password
OS_URL=https://your-environment.outsystemscloud.com
`

const dotEnvExample = `# OutSystems credentials (copy to .env and fill in values)
# For Personal Environments (basic auth): OS_AUTH_TOKEN=your_username:your_password
# For Enterprise Environments (bearer auth): OS_AUTH_TOKEN=your_service_account_token
OS_AUTH_TOKEN=
OS_URL=
`

const gitignoreContent = `# Secrets
.env
`
