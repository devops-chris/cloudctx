package cmd

import (
	"fmt"

	"github.com/devops-chris/cloudctx/internal/aws"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var awsLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to AWS SSO",
	Long: `Authenticate with AWS SSO.

Opens your browser to complete the SSO authentication flow.
After login, your SSO credentials will be cached for subsequent commands.

Examples:
  cloudctx aws login`,
	RunE: runAWSLogin,
}

func init() {
	awsCmd.AddCommand(awsLoginCmd)
}

func runAWSLogin(cmd *cobra.Command, args []string) error {
	p := aws.NewProvider(cfg.AWS.SSOStartURL, cfg.AWS.SSORegion, cfg.AWS.DefaultRegion)

	pterm.Info.Println("Opening browser for AWS SSO login...")
	pterm.FgGray.Println("Complete the authentication in your browser")
	fmt.Println()

	err := p.Login()
	if err != nil {
		pterm.Error.Println("Login failed")
		return err
	}

	pterm.Success.Println("Successfully logged in to AWS SSO")
	pterm.FgGray.Println("Run 'cloudctx aws sync' to update your profiles")

	return nil
}

