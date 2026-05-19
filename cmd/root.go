package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	workspacePath string
	version       = "dev"
)

// SetVersion allows main.go to inject the build-time version.
func SetVersion(v string) {
	version = v
}

var rootCmd = &cobra.Command{
	Use:   "os-agent",
	Short: "High-performance CLI orchestrator for OutSystems",
	Long: `
   ____  _____        _    ____  _____ _   _ _____
  / __ \/ ____|      / \  / ___|| ____| \ | |_   _|
 | |  | | (___      / _ \| |  _ |  _| |  \| | | |
 | |  | |\___ \    / ___ \ |_| || |___| |\  | | |
 |  __/|____) |  /_/   \_\____|_____|_| \_| |_|
  \___/ |_____/

  os-agent — Bridge local code to OutSystems.
  Watches your workspace and syncs schema + logic changes in real time.`,
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip env loading for the init command — the .env may not exist yet.
		if cmd.Name() == "init" {
			return nil
		}

		envPath := fmt.Sprintf("%s/.env", workspacePath)
		if err := godotenv.Load(envPath); err != nil {
			fmt.Fprintf(os.Stderr, "⚠  Could not load %s: %v\n", envPath, err)
			fmt.Fprintf(os.Stderr, "   Run 'os-agent init' first, then configure your .env file.\n")
			// Non-fatal: environment vars may be set externally.
		}

		// Validate required environment variables.
		if os.Getenv("OS_AUTH_TOKEN") == "" {
			fmt.Fprintln(os.Stderr, "⚠  OS_AUTH_TOKEN is not set. Metadata sync will fail.")
		}
		if os.Getenv("OS_URL") == "" {
			fmt.Fprintln(os.Stderr, "⚠  OS_URL is not set. Metadata sync will fail.")
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&workspacePath, "workspace", "w", "./workspace",
		"Path to the workspace directory",
	)
}

// Execute is called by main.go to start the CLI.
func Execute() {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
