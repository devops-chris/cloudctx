package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var awsLoginCmd = &cobra.Command{
	Use:   "login [--org NAME]",
	Short: "Login to AWS SSO",
	Long: `Authenticate with AWS SSO.

Opens your browser to complete the SSO authentication flow.
After login, your SSO credentials will be cached for subsequent commands.

Log in to the org of the profile in context. If you pass --org, that org is used instead.
  cloudctx aws login           # Org of current profile (or config default)
  cloudctx aws login --org work # Use this org

Examples:
  cloudctx aws login
  cloudctx aws login --org personal`,
	RunE: runAWSLogin,
}

func init() {
	awsCmd.AddCommand(awsLoginCmd)
}

func runAWSLogin(cmd *cobra.Command, args []string) error {
	// If --org passed, use it; else log in to the org of the profile in context
	orgKey := awsOrg
	if orgKey == "" {
		p := getAWSListProvider()
		if current, err := p.CurrentContext(); err == nil && current != nil && current.Org != "" {
			orgKey = current.Org
		}
	}
	if orgKey == "" {
		orgKey = cfg.AWSDefaultOrg()
	}
	if orgKey == "" {
		pterm.Error.Println("No AWS organization configured")
		pterm.FgGray.Println("Run 'cloudctx aws init' or add aws.organizations in config")
		return nil
	}

	p, ok := getAWSProviderForOrg(orgKey)
	if !ok {
		pterm.Error.Printf("Unknown organization %q\n", orgKey)
		pterm.FgGray.Println("Check aws.organizations in your config")
		return nil
	}

	pterm.Info.Printf("Opening browser for AWS SSO login (%s)...\n", orgKey)
	pterm.FgGray.Println("Complete the authentication in your browser")
	fmt.Println()

	err := p.Login()
	if err != nil {
		pterm.Error.Println("Login failed")
		return err
	}

	pterm.Success.Printf("Successfully logged in to AWS SSO (%s)\n", orgKey)
	pterm.FgGray.Println("Run 'cloudctx aws sync' to update your profiles")

	return nil
}

