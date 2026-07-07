package cmd

import (
	"fmt"

	"github.com/devops-chris/clihq/ui"
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
		fmt.Println(ui.Error("No AWS organization configured"))
		fmt.Println(ui.Subtle("Run 'cloudctx aws init' or add aws.organizations in config"))
		return nil
	}

	p, ok := getAWSProviderForOrg(orgKey)
	if !ok {
		fmt.Println(ui.Errorf("Unknown organization %q", orgKey))
		fmt.Println(ui.Subtle("Check aws.organizations in your config"))
		return nil
	}

	fmt.Println(ui.Infof("Opening browser for AWS SSO login (%s)...", orgKey))
	fmt.Println(ui.Subtle("Complete the authentication in your browser"))
	fmt.Println()

	if err := p.Login(); err != nil {
		fmt.Println(ui.Error("Login failed"))
		return err
	}

	fmt.Println(ui.Successf("Successfully logged in to AWS SSO (%s)", orgKey))
	fmt.Println(ui.Subtle("Run 'cloudctx aws sync' to update your profiles"))

	return nil
}
