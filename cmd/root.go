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
	Use:   "cloudctx [profile]",
	Short: "Switch between cloud contexts easily",
	Long: `cloudctx - A unified CLI for switching between cloud contexts.

Supports AWS profiles, Azure subscriptions, and more.

Quick Start:
  cloudctx                  # Interactive profile picker (uses default cloud)
  cloudctx <profile>        # Set specific profile
  cloudctx -c               # Show current profile
  cloudctx -l               # List all profiles
  
  cloudctx aws login        # AWS SSO login
  cloudctx aws sync         # Sync profiles from SSO
  cloudctx aws whoami       # Show current AWS identity

Configuration:
  cloudctx uses ~/.config/cloudctx/config.yaml for settings.
  
Environment Variables:
  CLOUDCTX_DEFAULT_CLOUD        Default cloud provider (aws, azure)
  CLOUDCTX_AWS_SSO_START_URL    AWS SSO portal URL
  CLOUDCTX_AWS_SSO_REGION       AWS SSO region`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRoot,
}

var (
	rootShowCurrent bool
	rootShowList    bool
)

func runRoot(cmd *cobra.Command, args []string) error {
	// Use default cloud (aws for now)
	if cfg.DefaultCloud != "aws" && cfg.DefaultCloud != "" {
		return fmt.Errorf("unsupported cloud: %s", cfg.DefaultCloud)
	}

	// Delegate to AWS command
	_ = awsCmd.Flags().Set("current", fmt.Sprintf("%v", rootShowCurrent))
	_ = awsCmd.Flags().Set("list", fmt.Sprintf("%v", rootShowList))
	return awsCmd.RunE(cmd, args)
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
	rootCmd.Flags().BoolVarP(&rootShowCurrent, "current", "c", false, "show current profile")
	rootCmd.Flags().BoolVarP(&rootShowList, "list", "l", false, "list all profiles")

	// Add shortcuts for common commands when using default cloud
	rootCmd.AddCommand(createShortcut("init", "Initialize cloud configuration", awsInitCmd))
	rootCmd.AddCommand(createShortcut("login", "Login to cloud provider", awsLoginCmd))
	rootCmd.AddCommand(createShortcut("sync", "Sync profiles from cloud", awsSyncCmd))
	rootCmd.AddCommand(createShortcut("whoami", "Show current identity", awsWhoamiCmd))
}

// createShortcut creates a root-level shortcut to a cloud-specific command
func createShortcut(name, short string, target *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short + " (uses default cloud)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return target.RunE(cmd, args)
		},
	}
}

func initConfig() {
	cfg = config.Load(cfgFile)
}

