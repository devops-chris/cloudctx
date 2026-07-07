package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/devops-chris/clihq/ui"
	"github.com/devops-chris/cloudctx/internal/aws"
	"github.com/devops-chris/cloudctx/internal/config"
	"github.com/spf13/cobra"
)

var awsOrgCmd = &cobra.Command{
	Use:   "org",
	Short: "Manage AWS organizations",
	Long: `Manage AWS organizations (SSO portals).

With multiple AWS accounts (e.g. work and personal), add each SSO portal as an org.
Then use --org with login and sync: cloudctx aws login --org work, cloudctx aws sync --org all.

Commands:
  add [name]     Add a new organization (prompts for SSO URL and regions)
  rename OLD NEW Rename an org (e.g. default -> work) for consistency
  remove NAME    Remove an org and its profiles from config and ~/.aws/config
  clean-credentials  Remove stale credential sections (fixes ghost [manual] entries after rename)`,
}

var awsOrgAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add an AWS organization",
	Long: `Add a new AWS organization (SSO portal) to your config.

Prompts for the org name (if not given), SSO start URL, and regions.
If this is your first org, it is set as the default. Otherwise you can set
default_organization in the config file to choose which org is used when
you omit --org.

Examples:
  cloudctx aws org add              # Prompts for name and URL
  cloudctx aws org add personal      # Add org named "personal", then prompt for URL`,
	RunE: runAWSOrgAdd,
}

var awsOrgRenameCmd = &cobra.Command{
	Use:   "rename [old] [new]",
	Short: "Rename an AWS organization",
	Long: `Rename an organization for consistency.

If you used 'aws init' (legacy single org), that org is named "default" internally.
To give it a real name: cloudctx aws org rename work  (one argument renames the only org)
Or explicitly: cloudctx aws org rename default work

Updates your config and ~/.aws/config (profile names, org labels, SSO session).
You may need to run 'cloudctx aws login --org <new>' again after renaming.

Examples:
  cloudctx aws org rename work              # Rename your only org (e.g. default) to "work"
  cloudctx aws org rename default work     # Same, explicit
  cloudctx aws org rename devopschris personal`,
	RunE: runAWSOrgRename,
}

var awsOrgRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an AWS organization",
	Long: `Remove an organization and all its profiles from your config and ~/.aws/config.

The org's SSO session and all cloudctx-managed profiles for that org are deleted.
If your current profile was from this org, it is cleared.

Examples:
  cloudctx aws org remove personal
  cloudctx aws org remove default   # Removes the legacy/single org (use with care)`,
	RunE: runAWSOrgRemove,
}

func init() {
	awsCmd.AddCommand(awsOrgCmd)
	awsOrgCmd.AddCommand(awsOrgAddCmd)
	awsOrgCmd.AddCommand(awsOrgRenameCmd)
	awsOrgCmd.AddCommand(awsOrgRemoveCmd)
	awsOrgCmd.AddCommand(awsOrgCleanCredentialsCmd)
}

var awsOrgCleanCredentialsCmd = &cobra.Command{
	Use:   "clean-credentials",
	Short: "Remove stale credential sections",
	Long: `Remove from ~/.aws/credentials any section that looks like a cloudctx profile (org/account:role)
but no longer exists in config. Use this once to fix ghost [manual] entries if you had them before
rename/remove were updated to clean credentials automatically. You usually don't need it for new renames.`,
	RunE: runAWSOrgCleanCredentials,
}

func runAWSOrgAdd(cmd *cobra.Command, args []string) error {
	orgName := ""
	if len(args) >= 1 {
		orgName = strings.TrimSpace(strings.ToLower(args[0]))
	}

	if orgName == "" {
		fmt.Println()
		var entered string
		err := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Org name").
				Description("Short name for this organization (e.g. work, personal)").
				Value(&entered),
		)).WithTheme(ui.Theme()).Run()
		if err != nil {
			return err
		}
		orgName = strings.TrimSpace(strings.ToLower(entered))
		if orgName == "" {
			fmt.Println(ui.Error("Org name is required"))
			return nil
		}
	}

	if strings.Contains(orgName, "/") || strings.Contains(orgName, " ") {
		fmt.Println(ui.Error("Org name cannot contain '/' or spaces"))
		return nil
	}

	orgs := cfg.AWSOrgs()
	if orgs != nil {
		if _, exists := orgs[orgName]; exists {
			fmt.Println(ui.Errorf("Organization %q already exists", orgName))
			return nil
		}
	}

	ssoURL := ""
	ssoRegion := "us-east-1"
	defaultRegion := "us-east-1"

	fmt.Println()
	err := huh.NewForm(huh.NewGroup(
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
	)).WithTheme(ui.Theme()).Run()
	if err != nil {
		return err
	}

	ssoURL = strings.TrimSpace(ssoURL)
	if ssoURL == "" {
		fmt.Println(ui.Error("SSO Start URL is required"))
		return nil
	}
	if strings.TrimSpace(ssoRegion) == "" {
		ssoRegion = "us-east-1"
	}
	if strings.TrimSpace(defaultRegion) == "" {
		defaultRegion = "us-east-1"
	}

	if cfg.AWS.Organizations == nil {
		cfg.AWS.Organizations = make(map[string]config.AWSOrgConfig)
		if cfg.AWS.SSOStartURL != "" {
			cfg.AWS.Organizations["default"] = config.AWSOrgConfig{
				SSOStartURL:   cfg.AWS.SSOStartURL,
				SSORegion:     cfg.AWS.SSORegion,
				DefaultRegion: cfg.AWS.DefaultRegion,
			}
		}
	}
	cfg.AWS.Organizations[orgName] = config.AWSOrgConfig{
		SSOStartURL:   ssoURL,
		SSORegion:     ssoRegion,
		DefaultRegion: defaultRegion,
	}
	if cfg.AWS.DefaultOrganization == "" {
		cfg.AWS.DefaultOrganization = orgName
	}

	configPath := cfgFile
	if configPath == "" {
		configPath = config.ConfigPath()
	}
	if err := config.WriteConfig(configPath, cfg); err != nil {
		fmt.Println(ui.Errorf("Failed to write config: %v", err))
		return err
	}

	fmt.Println()
	fmt.Println(ui.Successf("Added organization %q to %s", orgName, configPath))
	fmt.Println()
	fmt.Println(ui.Info("Next steps:"))
	fmt.Println(ui.Cyan("  cloudctx aws login --org " + orgName))
	fmt.Println(ui.Cyan("  cloudctx aws sync --org " + orgName))
	fmt.Println(ui.Subtle("Or sync all orgs: cloudctx aws sync --org all"))
	fmt.Println()
	return nil
}

func runAWSOrgRename(cmd *cobra.Command, args []string) error {
	oldName := ""
	newName := ""
	orgs := cfg.AWSOrgs()
	if len(orgs) == 0 {
		fmt.Println(ui.Error("No organizations configured"))
		fmt.Println(ui.Subtle("Add one with: cloudctx org add [name]"))
		return nil
	}

	if len(args) == 0 {
		if len(orgs) == 1 {
			for k := range orgs {
				oldName = k
				break
			}
			fmt.Println()
			err := huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("New org name").
					Description(fmt.Sprintf("Renaming %q — enter the new name (e.g. work, personal)", oldName)).
					Value(&newName),
			)).WithTheme(ui.Theme()).Run()
			if err != nil {
				return err
			}
			newName = strings.TrimSpace(strings.ToLower(newName))
			if newName == "" {
				fmt.Println(ui.Error("Org name is required"))
				return nil
			}
		} else {
			names := make([]string, 0, len(orgs))
			for k := range orgs {
				names = append(names, k)
			}
			sort.Strings(names)
			fmt.Println(ui.Error("You have multiple orgs; specify old and new name:"))
			fmt.Println(ui.Subtlef("  Your orgs: %s", strings.Join(names, ", ")))
			fmt.Println(ui.Subtle("  cloudctx org rename <old-name> <new-name>"))
			fmt.Println(ui.Subtlef("Example: cloudctx org rename %s work", names[0]))
			return nil
		}
	} else if len(args) == 1 {
		newName = strings.TrimSpace(strings.ToLower(args[0]))
		if newName == "" {
			fmt.Println(ui.Error("Usage: cloudctx org rename <new-name>   (when you have one org)"))
			fmt.Println(ui.Subtle("   or: cloudctx org rename <old-name> <new-name>"))
			return nil
		}
		if len(orgs) > 1 {
			names := make([]string, 0, len(orgs))
			for k := range orgs {
				names = append(names, k)
			}
			sort.Strings(names)
			fmt.Println(ui.Error("You have multiple orgs; specify both old and new name:"))
			fmt.Println(ui.Subtlef("  Your orgs: %s", strings.Join(names, ", ")))
			fmt.Println(ui.Subtle("  cloudctx org rename <old-name> " + newName))
			return nil
		}
		for k := range orgs {
			oldName = k
			break
		}
	} else if len(args) >= 2 {
		oldName = strings.TrimSpace(strings.ToLower(args[0]))
		newName = strings.TrimSpace(strings.ToLower(args[1]))
	}

	if oldName == "" || newName == "" {
		fmt.Println(ui.Error("Usage: cloudctx org rename <new-name>   (single org)"))
		fmt.Println(ui.Subtle("   or: cloudctx org rename <old-name> <new-name>"))
		fmt.Println(ui.Subtle("Example: cloudctx org rename work   or   cloudctx org rename default work"))
		return nil
	}
	if strings.Contains(newName, "/") || strings.Contains(newName, " ") {
		fmt.Println(ui.Error("New org name cannot contain '/' or spaces"))
		return nil
	}
	if _, ok := orgs[oldName]; !ok {
		fmt.Println(ui.Errorf("Organization %q not found", oldName))
		fmt.Println(ui.Subtle("Use 'cloudctx list' or 'cloudctx aws -l' to see orgs, or check your config"))
		return nil
	}
	if _, exists := orgs[newName]; exists && oldName != newName {
		fmt.Println(ui.Errorf("Organization %q already exists", newName))
		return nil
	}

	if oldName == "default" && cfg.AWS.Organizations == nil && cfg.AWS.SSOStartURL != "" {
		cfg.AWS.Organizations = map[string]config.AWSOrgConfig{
			"default": {
				SSOStartURL:   cfg.AWS.SSOStartURL,
				SSORegion:     cfg.AWS.SSORegion,
				DefaultRegion: cfg.AWS.DefaultRegion,
			},
		}
		if cfg.AWS.DefaultOrganization == "" {
			cfg.AWS.DefaultOrganization = "default"
		}
	}

	cfg.AWS.Organizations[newName] = cfg.AWS.Organizations[oldName]
	delete(cfg.AWS.Organizations, oldName)
	if cfg.AWS.DefaultOrganization == oldName {
		cfg.AWS.DefaultOrganization = newName
	}

	configPath := cfgFile
	if configPath == "" {
		configPath = config.ConfigPath()
	}
	if err := config.WriteConfig(configPath, cfg); err != nil {
		fmt.Println(ui.Errorf("Failed to write config: %v", err))
		return err
	}

	home, _ := os.UserHomeDir()
	awsConfigPath := filepath.Join(home, ".aws", "config")
	stateDir := config.ConfigDir()
	if err := aws.RenameOrg(awsConfigPath, stateDir, oldName, newName); err != nil {
		fmt.Println(ui.Errorf("Failed to update AWS config: %v", err))
		return err
	}

	fmt.Println(ui.Successf("Renamed organization %q to %q", oldName, newName))
	fmt.Println(ui.Subtle("You may need to run: cloudctx login --org " + newName))
	fmt.Println()
	return nil
}

func runAWSOrgRemove(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Println(ui.Error("Usage: cloudctx aws org remove <name>"))
		fmt.Println(ui.Subtle("Example: cloudctx aws org remove personal"))
		return nil
	}
	orgName := strings.TrimSpace(strings.ToLower(args[0]))
	if orgName == "" {
		fmt.Println(ui.Error("Org name is required"))
		return nil
	}

	orgs := cfg.AWSOrgs()
	if len(orgs) == 0 {
		fmt.Println(ui.Error("No organizations configured"))
		return nil
	}
	if _, ok := orgs[orgName]; !ok {
		fmt.Println(ui.Errorf("Organization %q not found", orgName))
		fmt.Println(ui.Subtle("Use 'cloudctx aws org add' to add orgs, or check your config"))
		return nil
	}

	if cfg.AWS.Organizations != nil {
		delete(cfg.AWS.Organizations, orgName)
		if cfg.AWS.DefaultOrganization == orgName {
			cfg.AWS.DefaultOrganization = ""
			for k := range cfg.AWS.Organizations {
				cfg.AWS.DefaultOrganization = k
				break
			}
		}
	}
	if orgName == "default" && cfg.AWS.Organizations == nil {
		cfg.AWS.SSOStartURL = ""
		cfg.AWS.SSORegion = ""
		cfg.AWS.DefaultRegion = ""
	}

	configPath := cfgFile
	if configPath == "" {
		configPath = config.ConfigPath()
	}
	if err := config.WriteConfig(configPath, cfg); err != nil {
		fmt.Println(ui.Errorf("Failed to write config: %v", err))
		return err
	}

	home, _ := os.UserHomeDir()
	awsConfigPath := filepath.Join(home, ".aws", "config")
	stateDir := config.ConfigDir()
	if err := aws.RemoveOrg(awsConfigPath, stateDir, orgName); err != nil {
		fmt.Println(ui.Errorf("Failed to update AWS config: %v", err))
		return err
	}

	fmt.Println(ui.Successf("Removed organization %q and its profiles", orgName))
	fmt.Println()
	return nil
}

func runAWSOrgCleanCredentials(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	awsConfigPath := filepath.Join(home, ".aws", "config")
	awsCredsPath := filepath.Join(home, ".aws", "credentials")
	removed, err := aws.CleanStaleCredentials(awsConfigPath, awsCredsPath)
	if err != nil {
		fmt.Println(ui.Errorf("Failed: %v", err))
		return err
	}
	if removed == 0 {
		fmt.Println(ui.Info("No stale credential sections found"))
		return nil
	}
	fmt.Println(ui.Successf("Removed %d stale section(s) from ~/.aws/credentials", removed))
	fmt.Println(ui.Subtle("Re-run your list or picker to confirm the ghost entries are gone."))
	fmt.Println()
	return nil
}
