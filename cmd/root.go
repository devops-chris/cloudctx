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

Omitting the provider runs the command against default_cloud (set in config; default: aws).
So 'ctx login' = 'ctx aws login' when default is AWS; same for list, sync, etc.

AWS:
  ctx aws                   Interactive profile picker
  ctx aws <profile>         Switch to profile
  ctx aws list   (or -l)    List profiles
  ctx aws current (or -c)   Show current profile
  ctx aws init              Configure SSO
  ctx aws login             SSO login
  ctx aws sync              Sync profiles from SSO
  ctx aws whoami            Show identity
  ctx aws org add|rename|...  Manage AWS organizations (SSO portals)

Azure:
  ctx azure                 Interactive subscription picker
  ctx azure <subscription>  Switch to subscription
  ctx azure list   (or -l)  List subscriptions
  ctx azure current (or -c) Show current subscription
  ctx azure login           Azure login (opens browser)
  ctx azure whoami          Show identity

Shortcuts (same commands, no provider — use default_cloud):
  ctx                       Picker
  ctx <name>                Switch to profile/subscription
  ctx list       (or -l)    List
  ctx current    (or -c)    Show current
  ctx login                 Login (org of current profile if AWS and no --org)
  ctx init                  Init
  ctx sync                  Sync
  ctx whoami                Whoami
  ctx version    (or -v)    Version
  ctx org add|rename|...    AWS orgs only (shortcut when default is AWS; else use ctx aws org ...)
  ctx doctor                 Check config and AWS setup

Flags: -l/-c/-v = list/current/version. --org = AWS org for login/sync/list. Config: ~/.config/cloudctx/config.yaml`,
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

	// --org is used by AWS (login/sync/list); on root so "cloudctx sync --org all" works
	rootCmd.PersistentFlags().StringVar(&awsOrg, "org", "", "AWS org for login/sync/list (e.g. work, or all for sync)")

	// Add shortcuts for common commands (routed based on default cloud)
	rootCmd.AddCommand(createLoginShortcut())
	rootCmd.AddCommand(createWhoamiShortcut())
	rootCmd.AddCommand(createListShortcut())
	rootCmd.AddCommand(createCurrentShortcut())
	rootCmd.AddCommand(createInitShortcut())
	rootCmd.AddCommand(createSyncShortcut())
	rootCmd.AddCommand(createOrgShortcut())
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

// createInitShortcut creates init shortcut that routes to default cloud
func createInitShortcut() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize default cloud configuration (uses default cloud)",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cfg.DefaultCloud {
			case "azure", "az":
				return azureInitCmd.RunE(cmd, args)
			default:
				return awsInitCmd.RunE(cmd, args)
			}
		},
	}
}

// createSyncShortcut creates sync shortcut that routes to default cloud
func createSyncShortcut() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync default cloud profiles/subscriptions (uses default cloud)",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cfg.DefaultCloud {
			case "azure", "az":
				return azureSyncCmd.RunE(cmd, args)
			default:
				return awsSyncCmd.RunE(cmd, args)
			}
		},
	}
}

// createOrgShortcut creates root-level "org" (AWS only). When default is AWS, ctx org ... works; else use ctx aws org ...
func createOrgShortcut() *cobra.Command {
	orgCmd := &cobra.Command{
		Use:   "org",
		Short: "Manage AWS organizations (shortcut when default is AWS)",
		Long:  "AWS only. When default_cloud is aws, use ctx org add|rename|remove|clean-credentials. When default is Azure, use ctx aws org ...",
	}
	orgCmd.AddCommand(createOrgSubShortcut("add", awsOrgAddCmd))
	orgCmd.AddCommand(createOrgSubShortcut("rename", awsOrgRenameCmd))
	orgCmd.AddCommand(createOrgSubShortcut("remove", awsOrgRemoveCmd))
	orgCmd.AddCommand(createOrgSubShortcut("clean-credentials", awsOrgCleanCredentialsCmd))
	return orgCmd
}

func createOrgSubShortcut(use string, target *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   target.Use,
		Short: target.Short,
		Long:  target.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.DefaultCloud == "azure" || cfg.DefaultCloud == "az" {
				return fmt.Errorf("org is for AWS only (default cloud is %q). Use: cloudctx aws org %s ...", cfg.DefaultCloud, use)
			}
			return target.RunE(cmd, args)
		},
	}
}

func initConfig() {
	cfg = config.Load(cfgFile)
}

