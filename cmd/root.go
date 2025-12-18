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

AWS:
  ctx aws                   Interactive profile picker
  ctx aws <profile>         Switch to profile
  ctx aws list   (or -l)    List profiles
  ctx aws current (or -c)   Show current profile
  ctx aws init              Configure SSO
  ctx aws login             SSO login
  ctx aws sync              Sync profiles from SSO
  ctx aws whoami            Show identity

Azure:
  ctx azure                 Interactive subscription picker
  ctx azure <subscription>  Switch to subscription
  ctx azure list   (or -l)  List subscriptions
  ctx azure current (or -c) Show current subscription
  ctx azure login           Azure login (opens browser)
  ctx azure whoami          Show identity

Shortcuts (routes to default_cloud, default: aws):
  ctx                       Interactive picker
  ctx <name>                Switch to profile/subscription
  ctx list       (or -l)    List all
  ctx current    (or -c)    Show current
  ctx login                 Login
  ctx whoami                Show identity
  ctx version    (or -v)    Show version

Note: -l/-c/-v are shortcuts for list/current/version commands.
      Use ONE or the OTHER, not both together.

Config: ~/.config/cloudctx/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRoot,
}

var (
	rootShowCurrent bool
	rootShowList    bool
)

func runRoot(cmd *cobra.Command, args []string) error {
	// Handle version flag
	if showVersion {
		fmt.Printf("cloudctx %s\n", version)
		return nil
	}

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

var showVersion bool

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/cloudctx/config.yaml)")
	rootCmd.Flags().BoolVarP(&rootShowCurrent, "current", "c", false, "show current profile")
	rootCmd.Flags().BoolVarP(&rootShowList, "list", "l", false, "list all profiles")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")

	// Add shortcuts for common commands (routed based on default cloud)
	rootCmd.AddCommand(createLoginShortcut())
	rootCmd.AddCommand(createWhoamiShortcut())
	rootCmd.AddCommand(createListShortcut())
	rootCmd.AddCommand(createCurrentShortcut())
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

// createListShortcut creates list shortcut that routes to default cloud
func createListShortcut() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all profiles/subscriptions (uses default cloud)",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cfg.DefaultCloud {
			case "azure", "az":
				_ = azureCmd.Flags().Set("list", "true")
				return azureCmd.RunE(cmd, args)
			default:
				_ = awsCmd.Flags().Set("list", "true")
				return awsCmd.RunE(cmd, args)
			}
		},
	}
}

// createCurrentShortcut creates current shortcut that routes to default cloud
func createCurrentShortcut() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current profile/subscription (uses default cloud)",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cfg.DefaultCloud {
			case "azure", "az":
				_ = azureCmd.Flags().Set("current", "true")
				return azureCmd.RunE(cmd, args)
			default:
				_ = awsCmd.Flags().Set("current", "true")
				return awsCmd.RunE(cmd, args)
			}
		},
	}
}

func initConfig() {
	cfg = config.Load(cfgFile)
}

