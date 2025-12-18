package cmd

import (
	"fmt"

	"github.com/devops-chris/cloudctx/internal/aws"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var awsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync AWS profiles from SSO",
	Long: `Synchronize AWS profiles from your SSO portal.

This command fetches all accounts and roles you have access to via SSO
and creates/updates AWS CLI profiles in ~/.aws/config.

Requires a valid SSO session - run 'cloudctx aws login' first if needed.

Examples:
  cloudctx aws sync`,
	RunE: runAWSSync,
}

func init() {
	awsCmd.AddCommand(awsSyncCmd)
}

func runAWSSync(cmd *cobra.Command, args []string) error {
	if cfg.AWS.SSOStartURL == "" {
		pterm.Error.Println("SSO Start URL not configured")
		fmt.Println()
		pterm.Info.Println("Configure it in ~/.config/cloudctx/config.yaml:")
		fmt.Println()
		pterm.FgCyan.Println("  aws:")
		pterm.FgCyan.Println("    sso_start_url: https://your-org.awsapps.com/start")
		pterm.FgCyan.Println("    sso_region: us-east-1")
		fmt.Println()
		pterm.Info.Println("Or set environment variable:")
		pterm.FgCyan.Println("  export CLOUDCTX_AWS_SSO_START_URL=https://your-org.awsapps.com/start")
		return nil
	}

	p := aws.NewProvider(cfg.AWS.SSOStartURL, cfg.AWS.SSORegion, cfg.AWS.DefaultRegion)

	spinner, _ := pterm.DefaultSpinner.Start("Syncing profiles from AWS SSO...")

	err := p.Sync()
	if err != nil {
		spinner.Fail("Sync failed")
		pterm.FgGray.Println("Try running 'cloudctx aws login' first")
		return err
	}

	_ = spinner.Stop()

	// Show results
	contexts, err := p.ListContexts()
	if err != nil {
		return err
	}

	pterm.Success.Printf("Synced %d profiles from AWS SSO\n", len(contexts))
	fmt.Println()
	pterm.FgGray.Println("Run 'cloudctx aws' to select a profile")

	return nil
}

