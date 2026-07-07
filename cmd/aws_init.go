package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/devops-chris/clihq/ui"
	"github.com/devops-chris/cloudctx/internal/config"
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
	fmt.Println(ui.SectionHeader("AWS SSO Configuration"))
	fmt.Println()

	ssoURL := cfg.AWS.SSOStartURL
	ssoRegion := "us-east-1"
	defaultRegion := "us-east-1"

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("SSO Start URL").
				Description("Example: https://your-org.awsapps.com/start").
				Value(&ssoURL),
			huh.NewInput().
				Title("SSO Region").
				Value(&ssoRegion),
			huh.NewInput().
				Title("Default region for profiles").
				Value(&defaultRegion),
		),
	).WithTheme(ui.Theme()).Run()
	if err != nil {
		return err
	}

	if ssoURL == "" {
		fmt.Println(ui.Error("SSO Start URL is required"))
		return nil
	}

	configDir := config.ConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

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
	fmt.Println(ui.Successf("Configuration saved to %s", configPath))
	fmt.Println()
	fmt.Println(ui.Info("Next steps:"))
	fmt.Println(ui.Cyan("  1. cloudctx aws login    # Authenticate with SSO"))
	fmt.Println(ui.Cyan("  2. cloudctx aws sync     # Fetch available profiles"))
	fmt.Println(ui.Cyan("  3. cloudctx aws          # Select a profile"))
	fmt.Println(ui.Subtle("To add another AWS org later: cloudctx aws org add"))
	fmt.Println()

	return nil
}
