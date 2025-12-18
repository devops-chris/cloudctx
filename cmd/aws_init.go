package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devops-chris/cloudctx/internal/config"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var awsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AWS SSO configuration",
	Long: `Set up cloudctx for AWS SSO.

This interactive command will configure your AWS SSO settings:
- SSO Start URL (your organization's AWS SSO portal)
- SSO Region
- Default AWS region for profiles

Examples:
  cloudctx aws init`,
	RunE: runAWSInit,
}

func init() {
	awsCmd.AddCommand(awsInitCmd)
}

func runAWSInit(cmd *cobra.Command, args []string) error {
	fmt.Println()
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println("AWS SSO Configuration")
	fmt.Println()

	// SSO Start URL
	pterm.Info.Println("Enter your AWS SSO portal URL")
	pterm.FgGray.Println("Example: https://your-org.awsapps.com/start")
	fmt.Println()

	ssoURL, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultValue(cfg.AWS.SSOStartURL).
		Show("SSO Start URL")

	if ssoURL == "" {
		pterm.Error.Println("SSO Start URL is required")
		return nil
	}

	// SSO Region
	fmt.Println()
	pterm.Info.Println("Enter your AWS SSO region")
	fmt.Println()

	ssoRegion, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("us-east-1").
		Show("SSO Region")

	// Default Region
	fmt.Println()
	pterm.Info.Println("Enter default AWS region for profiles")
	fmt.Println()

	defaultRegion, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("us-east-1").
		Show("Default Region")

	// Create config directory
	configDir := config.ConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := fmt.Sprintf(`# cloudctx configuration

# Default cloud provider
default_cloud: aws

# AWS settings
aws:
  sso_start_url: %s
  sso_region: %s
  default_region: %s

# Azure settings (future)
# azure:
#   tenant_id: ""
#   subscription_id: ""
`, ssoURL, ssoRegion, defaultRegion)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Println()
	pterm.Success.Printf("Configuration saved to %s\n", configPath)
	fmt.Println()
	pterm.Info.Println("Next steps:")
	pterm.FgCyan.Println("  1. cloudctx aws login    # Authenticate with SSO")
	pterm.FgCyan.Println("  2. cloudctx aws sync     # Fetch available profiles")
	pterm.FgCyan.Println("  3. cloudctx aws          # Select a profile")
	fmt.Println()

	return nil
}

