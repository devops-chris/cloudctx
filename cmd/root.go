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

Supports AWS profiles and Azure subscriptions.

AWS Commands:
  cloudctx aws              # Interactive AWS profile picker
  cloudctx aws -l           # List AWS profiles
  cloudctx aws init         # Configure AWS SSO
  cloudctx aws login        # AWS SSO login
  cloudctx aws sync         # Sync profiles from SSO
  cloudctx aws whoami       # Show AWS identity

Azure Commands:
  cloudctx azure            # Interactive Azure subscription picker
  cloudctx azure -l         # List Azure subscriptions  
  cloudctx azure login      # Azure login (opens browser)
  cloudctx azure whoami     # Show Azure identity

Shortcuts (uses default_cloud from config, default: aws):
  cloudctx                  # Interactive picker
  cloudctx <name>           # Switch to profile/subscription
  cloudctx -l               # List all
  cloudctx -c               # Show current

Configuration:
  Config file: ~/.config/cloudctx/config.yaml
  Set 'default_cloud: azure' to use Azure as default.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRoot,
}

var (
	rootShowCurrent bool
	rootShowList    bool
)

func runRoot(cmd *cobra.Command, args []string) error {
	// Route to appropriate cloud provider
	switch cfg.DefaultCloud {
	case "azure", "az":
		_ = azureCmd.Flags().Set("current", fmt.Sprintf("%v", rootShowCurrent))
		_ = azureCmd.Flags().Set("list", fmt.Sprintf("%v", rootShowList))
		return azureCmd.RunE(cmd, args)
	case "aws", "":
		// AWS is default
		_ = awsCmd.Flags().Set("current", fmt.Sprintf("%v", rootShowCurrent))
		_ = awsCmd.Flags().Set("list", fmt.Sprintf("%v", rootShowList))
		return awsCmd.RunE(cmd, args)
	default:
		return fmt.Errorf("unsupported cloud: %s (supported: aws, azure)", cfg.DefaultCloud)
	}
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

	// Add shortcuts for common commands (routed based on default cloud)
	rootCmd.AddCommand(createLoginShortcut())
	rootCmd.AddCommand(createWhoamiShortcut())
	// Note: init and sync are AWS-specific for now
	rootCmd.AddCommand(createShortcut("init", "Initialize AWS SSO configuration", awsInitCmd))
	rootCmd.AddCommand(createShortcut("sync", "Sync AWS profiles from SSO", awsSyncCmd))
}

// createShortcut creates a root-level shortcut to a specific command
func createShortcut(name, short string, target *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return target.RunE(cmd, args)
		},
	}
}

// createLoginShortcut creates login shortcut that routes to default cloud
func createLoginShortcut() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login to cloud provider (uses default cloud)",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cfg.DefaultCloud {
			case "azure", "az":
				return azureLoginCmd.RunE(cmd, args)
			default:
				return awsLoginCmd.RunE(cmd, args)
			}
		},
	}
}

// createWhoamiShortcut creates whoami shortcut that routes to default cloud
func createWhoamiShortcut() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current identity (uses default cloud)",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cfg.DefaultCloud {
			case "azure", "az":
				return azureWhoamiCmd.RunE(cmd, args)
			default:
				return awsWhoamiCmd.RunE(cmd, args)
			}
		},
	}
}

func initConfig() {
	cfg = config.Load(cfgFile)
}

