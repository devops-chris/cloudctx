package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/devops-chris/clihq/ui"
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
		fmt.Println(ui.Error("No AWS organization configured"))
		fmt.Println()
		fmt.Println(ui.Info("Configure it in ~/.config/cloudctx/config.yaml:"))
		fmt.Println()
		fmt.Println(ui.Cyan("  aws:"))
		fmt.Println(ui.Cyan("    sso_start_url: https://your-org.awsapps.com/start"))
		fmt.Println(ui.Cyan("    sso_region: us-east-1"))
		fmt.Println()
		fmt.Println(ui.Subtle("Or add aws.organizations for multi-org."))
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
			names := make([]string, 0, len(orgs))
			for k := range orgs {
				names = append(names, k)
			}
			sort.Strings(names)
			fmt.Println(ui.Errorf("Unknown organization %q", awsOrg))
			fmt.Println(ui.Subtlef("Known orgs: %s (check aws.organizations in config)", strings.Join(names, ", ")))
			return nil
		}
		toSync = []string{awsOrg}
	} else {
		def := cfg.AWSDefaultOrg()
		if def == "" {
			fmt.Println(ui.Error("No default organization; use --org <name> or --org all"))
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
		var syncErr error
		_ = spinner.New().
			Title("Syncing " + orgKey + " from AWS SSO...").
			Action(func() { syncErr = p.Sync() }).
			Run()
		if syncErr != nil {
			fmt.Println(ui.Error("Sync failed for " + orgKey))
			fmt.Println(ui.Subtle("Try running 'cloudctx aws login --org " + orgKey + "' first"))
			failed = append(failed, orgKey)
			continue
		}
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
		fmt.Println(ui.Successf("Synced %d profiles for %s", count, orgKey))
	}

	fmt.Println()
	if len(failed) > 0 {
		fmt.Println(ui.Warningf("Failed for %d org(s): %s", len(failed), strings.Join(failed, ", ")))
		fmt.Println(ui.Subtle("Run 'cloudctx aws login --org <name>' for each, then sync again."))
		return fmt.Errorf("sync failed for: %s", strings.Join(failed, ", "))
	}
	fmt.Println(ui.Subtle("Run 'cloudctx aws' to select a profile"))
	return nil
}
