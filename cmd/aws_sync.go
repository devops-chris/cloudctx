package cmd

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var awsSyncCmd = &cobra.Command{
	Use:   "sync [--org NAME|all]",
	Short: "Sync AWS profiles from SSO",
	Long: `Synchronize AWS profiles from your SSO portal.

This command fetches all accounts and roles you have access to via SSO
and creates/updates AWS CLI profiles in ~/.aws/config.

Requires a valid SSO session - run 'cloudctx aws login' (and optionally --org) first.

With multiple organizations:
  cloudctx aws sync             # Sync default org only
  cloudctx aws sync --org work   # Sync the "work" org only
  cloudctx aws sync --org all    # Sync every configured org

Examples:
  cloudctx aws sync
  cloudctx aws sync --org all`,
	RunE: runAWSSync,
}

func init() {
	awsCmd.AddCommand(awsSyncCmd)
}

func runAWSSync(cmd *cobra.Command, args []string) error {
	orgs := cfg.AWSOrgs()
	if len(orgs) == 0 {
		pterm.Error.Println("No AWS organization configured")
		fmt.Println()
		pterm.Info.Println("Configure it in ~/.config/cloudctx/config.yaml:")
		fmt.Println()
		pterm.FgCyan.Println("  aws:")
		pterm.FgCyan.Println("    sso_start_url: https://your-org.awsapps.com/start")
		pterm.FgCyan.Println("    sso_region: us-east-1")
		fmt.Println()
		pterm.Info.Println("Or add aws.organizations for multi-org.")
		return nil
	}

	// Resolve which org(s) to sync
	var toSync []string
	if awsOrg == "all" {
		for k := range orgs {
			toSync = append(toSync, k)
		}
	} else if awsOrg != "" {
		if _, ok := orgs[awsOrg]; !ok {
			pterm.Error.Printf("Unknown organization %q\n", awsOrg)
			return nil
		}
		toSync = []string{awsOrg}
	} else {
		def := cfg.AWSDefaultOrg()
		if def == "" {
			pterm.Error.Println("No default organization; use --org <name> or --org all")
			return nil
		}
		toSync = []string{def}
	}

	var failed []string
	for _, orgKey := range toSync {
		p, ok := getAWSProviderForOrg(orgKey)
		if !ok {
			continue
		}
		spinner, _ := pterm.DefaultSpinner.Start("Syncing " + orgKey + " from AWS SSO...")
		err := p.Sync()
		if err != nil {
			spinner.Fail("Sync failed for " + orgKey)
			pterm.FgGray.Println("Try running 'cloudctx aws login --org " + orgKey + "' first")
			failed = append(failed, orgKey)
			continue
		}
		_ = spinner.Stop()
		contexts, err := p.ListContexts()
		if err != nil {
			failed = append(failed, orgKey)
			continue
		}
		count := 0
		for _, c := range contexts {
			if c.Org == orgKey || (orgKey == "default" && c.Org == "") {
				count++
			}
		}
		pterm.Success.Printf("Synced %d profiles for %s\n", count, orgKey)
	}

	fmt.Println()
	if len(failed) > 0 {
		pterm.Warning.Printf("Failed for %d org(s): %s\n", len(failed), strings.Join(failed, ", "))
		pterm.FgGray.Println("Run 'cloudctx aws login --org <name>' for each, then sync again.")
		return fmt.Errorf("sync failed for: %s", strings.Join(failed, ", "))
	}
	pterm.FgGray.Println("Run 'cloudctx aws' to select a profile")
	return nil
}

