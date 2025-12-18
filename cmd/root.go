package cmd

import (
	"fmt"
	"os"

	"github.com/devops-chris/cloudctx/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg       *config.Config
	cfgFile   string
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// SetVersion sets the version info from build flags
func SetVersion(v, c, d string) {
	version = v
	commit = c
	buildDate = d
}

var rootCmd = &cobra.Command{
	Use:   "cloudctx",
	Short: "Switch between cloud contexts easily",
	Long: `cloudctx - A unified CLI for switching between cloud contexts.

Supports AWS profiles, Azure subscriptions, and more.

Quick Start:
  cloudctx aws              # Interactive AWS profile picker
  cloudctx aws login        # AWS SSO login
  cloudctx aws sync         # Sync profiles from SSO
  cloudctx aws whoami       # Show current AWS identity
  cloudctx aws list         # List all AWS profiles
  cloudctx aws <profile>    # Set specific AWS profile

Configuration:
  cloudctx uses ~/.config/cloudctx/config.yaml for settings.
  
Environment Variables:
  CLOUDCTX_DEFAULT_CLOUD    Default cloud provider (aws, azure)
  CLOUDCTX_AWS_SSO_START_URL    AWS SSO portal URL
  CLOUDCTX_AWS_SSO_REGION       AWS SSO region`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: show help or run default cloud
		_ = cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/cloudctx/config.yaml)")
}

func initConfig() {
	cfg = config.Load(cfgFile)
}

