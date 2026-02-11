package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/devops-chris/cloudctx/internal/aws"
	"github.com/devops-chris/cloudctx/internal/config"
	"github.com/pterm/pterm"
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
		pterm.Info.Println("Enter a short name for this organization (e.g. work, personal)")
		pterm.FgGray.Println("Used with: cloudctx aws login --org " + pterm.Cyan("<name>"))
		fmt.Println()
		entered, _ := pterm.DefaultInteractiveTextInput.Show("Org name")
		orgName = strings.TrimSpace(strings.ToLower(entered))
		if orgName == "" {
			pterm.Error.Println("Org name is required")
			return nil
		}
	}

	// Disallow slash so profile names stay clean
	if strings.Contains(orgName, "/") || strings.Contains(orgName, " ") {
		pterm.Error.Println("Org name cannot contain '/' or spaces")
		return nil
	}

	orgs := cfg.AWSOrgs()
	if orgs != nil {
		if _, exists := orgs[orgName]; exists {
			pterm.Error.Printf("Organization %q already exists\n", orgName)
			return nil
		}
	}

	fmt.Println()
	pterm.Info.Printf("Adding organization %q\n", orgName)
	pterm.FgGray.Println("Example SSO URL: https://your-org.awsapps.com/start")
	fmt.Println()

	ssoURL, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("").
		Show("SSO Start URL")
	ssoURL = strings.TrimSpace(ssoURL)
	if ssoURL == "" {
		pterm.Error.Println("SSO Start URL is required")
		return nil
	}

	ssoRegion, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("us-east-1").
		Show("SSO Region")
	ssoRegion = strings.TrimSpace(ssoRegion)
	if ssoRegion == "" {
		ssoRegion = "us-east-1"
	}

	defaultRegion, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("us-east-1").
		Show("Default region for profiles")
	defaultRegion = strings.TrimSpace(defaultRegion)
	if defaultRegion == "" {
		defaultRegion = "us-east-1"
	}

	// Ensure we have an organizations map (convert from legacy if needed)
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
		pterm.Error.Printf("Failed to write config: %v\n", err)
		return err
	}

	pterm.Success.Printf("Added organization %q to %s\n", orgName, configPath)
	fmt.Println()
	pterm.Info.Println("Next steps:")
	pterm.FgCyan.Printf("  cloudctx aws login --org %s\n", orgName)
	pterm.FgCyan.Printf("  cloudctx aws sync --org %s\n", orgName)
	pterm.FgGray.Println("Or sync all orgs: cloudctx aws sync --org all")
	fmt.Println()
	return nil
}

func runAWSOrgRename(cmd *cobra.Command, args []string) error {
	oldName := ""
	newName := ""
	orgs := cfg.AWSOrgs()
	if orgs == nil || len(orgs) == 0 {
		pterm.Error.Println("No organizations configured")
		pterm.FgGray.Println("Add one with: cloudctx org add [name]")
		return nil
	}

	if len(args) == 0 {
		// No args: if single org, prompt for new name; otherwise show usage
		if len(orgs) == 1 {
			for k := range orgs {
				oldName = k
				break
			}
			pterm.Info.Printf("You have one org %q. Enter the new name (e.g. work, personal):\n", oldName)
			fmt.Println()
			entered, _ := pterm.DefaultInteractiveTextInput.Show("New org name")
			newName = strings.TrimSpace(strings.ToLower(entered))
			if newName == "" {
				pterm.Error.Println("Org name is required")
				return nil
			}
		} else {
			names := make([]string, 0, len(orgs))
			for k := range orgs {
				names = append(names, k)
			}
			sort.Strings(names)
			pterm.Error.Println("You have multiple orgs; specify old and new name:")
			pterm.FgGray.Printf("  Your orgs: %s\n", strings.Join(names, ", "))
			pterm.FgGray.Println("  cloudctx org rename <old-name> <new-name>")
			pterm.FgGray.Printf("Example: cloudctx org rename %s work\n", names[0])
			return nil
		}
	} else if len(args) == 1 {
		newName = strings.TrimSpace(strings.ToLower(args[0]))
		if newName == "" {
			pterm.Error.Println("Usage: cloudctx org rename <new-name>   (when you have one org)")
			pterm.FgGray.Println("   or: cloudctx org rename <old-name> <new-name>")
			return nil
		}
		if len(orgs) > 1 {
			names := make([]string, 0, len(orgs))
			for k := range orgs {
				names = append(names, k)
			}
			sort.Strings(names)
			pterm.Error.Println("You have multiple orgs; specify both old and new name:")
			pterm.FgGray.Printf("  Your orgs: %s\n", strings.Join(names, ", "))
			pterm.FgGray.Println("  cloudctx org rename <old-name> " + newName)
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
		pterm.Error.Println("Usage: cloudctx org rename <new-name>   (single org)")
		pterm.FgGray.Println("   or: cloudctx org rename <old-name> <new-name>")
		pterm.FgGray.Println("Example: cloudctx org rename work   or   cloudctx org rename default work")
		return nil
	}
	if strings.Contains(newName, "/") || strings.Contains(newName, " ") {
		pterm.Error.Println("New org name cannot contain '/' or spaces")
		return nil
	}

	if _, ok := orgs[oldName]; !ok {
		pterm.Error.Printf("Organization %q not found\n", oldName)
		pterm.FgGray.Println("Use 'cloudctx list' or 'cloudctx aws -l' to see orgs, or check your config")
		return nil
	}
	if _, exists := orgs[newName]; exists && oldName != newName {
		pterm.Error.Printf("Organization %q already exists\n", newName)
		return nil
	}

	// If renaming "default" and we only have legacy config, migrate to organizations first
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

	// Update cloudctx config
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
		pterm.Error.Printf("Failed to write config: %v\n", err)
		return err
	}

	// Update ~/.aws/config and state
	home, _ := os.UserHomeDir()
	awsConfigPath := filepath.Join(home, ".aws", "config")
	stateDir := config.ConfigDir()
	if err := aws.RenameOrg(awsConfigPath, stateDir, oldName, newName); err != nil {
		pterm.Error.Printf("Failed to update AWS config: %v\n", err)
		return err
	}

	pterm.Success.Printf("Renamed organization %q to %q\n", oldName, newName)
	pterm.FgGray.Println("You may need to run: cloudctx login --org " + newName)
	fmt.Println()
	return nil
}

func runAWSOrgRemove(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		pterm.Error.Println("Usage: cloudctx aws org remove <name>")
		pterm.FgGray.Println("Example: cloudctx aws org remove personal")
		return nil
	}
	orgName := strings.TrimSpace(strings.ToLower(args[0]))
	if orgName == "" {
		pterm.Error.Println("Org name is required")
		return nil
	}

	orgs := cfg.AWSOrgs()
	if orgs == nil || len(orgs) == 0 {
		pterm.Error.Println("No organizations configured")
		return nil
	}
	if _, ok := orgs[orgName]; !ok {
		pterm.Error.Printf("Organization %q not found\n", orgName)
		pterm.FgGray.Println("Use 'cloudctx aws org add' to add orgs, or check your config")
		return nil
	}

	// Remove from cloudctx config
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
		// Legacy config: clear the top-level AWS fields
		cfg.AWS.SSOStartURL = ""
		cfg.AWS.SSORegion = ""
		cfg.AWS.DefaultRegion = ""
	}

	configPath := cfgFile
	if configPath == "" {
		configPath = config.ConfigPath()
	}
	if err := config.WriteConfig(configPath, cfg); err != nil {
		pterm.Error.Printf("Failed to write config: %v\n", err)
		return err
	}

	home, _ := os.UserHomeDir()
	awsConfigPath := filepath.Join(home, ".aws", "config")
	stateDir := config.ConfigDir()
	if err := aws.RemoveOrg(awsConfigPath, stateDir, orgName); err != nil {
		pterm.Error.Printf("Failed to update AWS config: %v\n", err)
		return err
	}

	pterm.Success.Printf("Removed organization %q and its profiles\n", orgName)
	fmt.Println()
	return nil
}

func runAWSOrgCleanCredentials(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	awsConfigPath := filepath.Join(home, ".aws", "config")
	awsCredsPath := filepath.Join(home, ".aws", "credentials")
	removed, err := aws.CleanStaleCredentials(awsConfigPath, awsCredsPath)
	if err != nil {
		pterm.Error.Printf("Failed: %v\n", err)
		return err
	}
	if removed == 0 {
		pterm.Info.Println("No stale credential sections found")
		return nil
	}
	pterm.Success.Printf("Removed %d stale section(s) from ~/.aws/credentials\n", removed)
	pterm.FgGray.Println("Re-run your list or picker to confirm the ghost entries are gone.")
	fmt.Println()
	return nil
}
