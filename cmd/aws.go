package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/devops-chris/cloudctx/internal/aws"
	"github.com/devops-chris/cloudctx/internal/provider"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// orgPalette assigns distinct colors to orgs (deterministic by sorted name). Same org = same color every time.
var orgPalette = []*pterm.Style{
	pterm.NewStyle(pterm.FgCyan),
	pterm.NewStyle(pterm.FgMagenta),
	pterm.NewStyle(pterm.FgBlue),
	pterm.NewStyle(pterm.FgLightCyan),
	pterm.NewStyle(pterm.FgLightMagenta),
	pterm.NewStyle(pterm.FgLightBlue),
}

// orgColorMap returns a map of org name -> palette index for the given set of orgs (sorted for determinism).
func orgColorMap(orgsSeen map[string]bool) map[string]int {
	names := make([]string, 0, len(orgsSeen))
	for k := range orgsSeen {
		if k != "" {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	m := make(map[string]int, len(names))
	for i, n := range names {
		m[n] = i % len(orgPalette)
	}
	return m
}

func sprintOrg(org string, colorIndex map[string]int) string {
	if org == "" {
		return pterm.FgGray.Sprint("-")
	}
	idx, ok := colorIndex[org]
	if !ok {
		return pterm.FgGray.Sprint(org)
	}
	return orgPalette[idx].Sprint(org)
}

func sprintOrgTag(org string, colorIndex map[string]int) string {
	if org == "" {
		return pterm.FgGray.Sprint("[]")
	}
	idx, ok := colorIndex[org]
	if !ok {
		return pterm.FgGray.Sprint("[" + org + "]")
	}
	return orgPalette[idx].Sprint("[" + org + "]")
}

// getAWSListProvider returns a provider used for list/set/current/picker (reads all profiles from config).
func getAWSListProvider() *aws.Provider {
	orgs := cfg.AWSOrgs()
	defOrg := cfg.AWSDefaultOrg()
	if defOrg != "" && orgs != nil {
		if o, ok := orgs[defOrg]; ok {
			return aws.NewProvider("", o.SSOStartURL, o.SSORegion, o.DefaultRegion)
		}
	}
	return aws.NewProvider("", cfg.AWS.SSOStartURL, cfg.AWS.SSORegion, cfg.AWS.DefaultRegion)
}

// getAWSProviderForOrg returns a provider for the given org key (for login/sync). Returns nil, nil if org not found.
func getAWSProviderForOrg(orgKey string) (*aws.Provider, bool) {
	orgs := cfg.AWSOrgs()
	if orgs == nil {
		return nil, false
	}
	o, ok := orgs[orgKey]
	if !ok {
		return nil, false
	}
	return aws.NewProvider(orgKey, o.SSOStartURL, o.SSORegion, o.DefaultRegion), true
}

var awsCmd = &cobra.Command{
	Use:   "aws [profile]",
	Short: "Manage AWS profiles",
	Long: `Manage AWS profiles and SSO authentication.

Without arguments, opens an interactive profile picker.
With a profile name, sets that profile directly.

With multiple AWS organizations (SSO portals), use --org with login and sync:
  cloudctx aws login --org work      # Log in to the "work" org
  cloudctx aws sync --org work       # Sync profiles for that org
  cloudctx aws sync --org all        # Sync all configured orgs
Add orgs via: cloudctx aws org add

Examples:
  cloudctx aws                    # Interactive picker
  cloudctx aws my-account:admin   # Set specific profile
  cloudctx aws -c                 # Show current profile
  cloudctx aws -l                 # List all profiles`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAWS,
}

var (
	awsShowCurrent bool
	awsShowList    bool
	awsSSOOnly     bool
	awsManualOnly  bool
	awsOrg         string // --org: target org for login/sync (empty = default org)
)

func init() {
	rootCmd.AddCommand(awsCmd)

	// --org is on rootCmd so "cloudctx sync --org all" and "cloudctx aws sync --org all" both work
	awsCmd.Flags().BoolVarP(&awsShowCurrent, "current", "c", false, "show current profile")
	awsCmd.Flags().BoolVarP(&awsShowList, "list", "l", false, "list all profiles")
	awsCmd.Flags().BoolVar(&awsSSOOnly, "sso", false, "show only SSO-synced profiles")
	awsCmd.Flags().BoolVar(&awsManualOnly, "manual", false, "show only manually created profiles")
}

func runAWS(cmd *cobra.Command, args []string) error {
	p := getAWSListProvider()

	// Show current profile
	if awsShowCurrent {
		return showCurrentAWS(p)
	}

	// List all profiles
	if awsShowList {
		return listAWS(p)
	}

	// Set specific profile
	if len(args) == 1 {
		return setAWS(p, args[0])
	}

	// Interactive picker
	return interactiveAWS(p)
}

func showCurrentAWS(p *aws.Provider) error {
	current, err := p.CurrentContext()
	if err != nil {
		return err
	}

	if current == nil {
		pterm.Warning.Println("No AWS profile set")
		pterm.FgGray.Println("Set one with: cloudctx aws <profile>")
		return nil
	}

	fmt.Println(current.Name)
	return nil
}

// filterContexts applies the --sso and --manual flags
func filterContexts(contexts []provider.Context) []provider.Context {
	if !awsSSOOnly && !awsManualOnly {
		return contexts // No filter
	}

	var filtered []provider.Context
	for _, ctx := range contexts {
		if awsSSOOnly && ctx.Managed {
			filtered = append(filtered, ctx)
		} else if awsManualOnly && !ctx.Managed {
			filtered = append(filtered, ctx)
		}
	}
	return filtered
}

// filterByOrg filters to profiles from the given org. Empty or "all" = no filter.
func filterByOrg(contexts []provider.Context, org string) []provider.Context {
	if org == "" || org == "all" {
		return contexts
	}
	var out []provider.Context
	for _, ctx := range contexts {
		if ctx.Org == org {
			out = append(out, ctx)
		}
	}
	return out
}

// awsProfileDisplayName returns the profile name for display: account:role only (strip org/ prefix when present).
func awsProfileDisplayName(name, org string) string {
	if org != "" && strings.HasPrefix(name, org+"/") {
		return name[len(org)+1:]
	}
	return name
}

func listAWS(p *aws.Provider) error {
	contexts, err := p.ListContexts()
	if err != nil {
		return err
	}

	contexts = filterContexts(contexts)
	contexts = filterByOrg(contexts, awsOrg)

	if len(contexts) == 0 {
		pterm.Warning.Println("No AWS profiles found")
		if awsOrg != "" && awsOrg != "all" {
			pterm.FgGray.Printf("Try without --org to see all orgs, or use --org all\n")
		} else if awsSSOOnly {
			pterm.FgGray.Println("Run 'cloudctx aws sync' to fetch profiles from SSO")
		} else if awsManualOnly {
			pterm.FgGray.Println("No manually created profiles found")
		} else {
			pterm.FgGray.Println("Run 'cloudctx aws sync' to fetch profiles from SSO")
		}
		return nil
	}

	fmt.Println()
	headerTitle := "AWS Profiles"
	if awsOrg != "" && awsOrg != "all" {
		headerTitle = "AWS Profiles (org: " + awsOrg + ")"
	}
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println(headerTitle)

	// Show Org column when 2+ orgs, or when the single org has a real name (not "default") so renamed org shows
	orgsSeen := make(map[string]bool)
	for _, ctx := range contexts {
		if ctx.Org != "" {
			orgsSeen[ctx.Org] = true
		}
	}
	showOrg := len(orgsSeen) > 1
	if len(orgsSeen) == 1 {
		for k := range orgsSeen {
			if k != "default" {
				showOrg = true
			}
			break
		}
	}
	orgColors := orgColorMap(orgsSeen)

	headers := []string{"", "Profile", "Account ID", "Role", "Region", "Source"}
	if showOrg {
		headers = []string{"", "Profile", "Org", "Account ID", "Role", "Region", "Source"}
	}
	tableData := pterm.TableData{headers}

	const maxProfileLen = 40
	for _, ctx := range contexts {
		marker := " "
		name := awsProfileDisplayName(ctx.Name, ctx.Org)
		if len(name) > maxProfileLen {
			name = name[:maxProfileLen-3] + "..."
		}
		if ctx.Active {
			marker = "*"
			name = pterm.FgGreen.Sprint(name)
		}
		// [sso] = profile uses SSO (sso_session or sso_account_id in config); [manual] = key-based (see ListContexts in internal/aws/provider.go)
		source := pterm.FgYellow.Sprint("manual")
		if ctx.Managed {
			source = pterm.FgCyan.Sprint("sso")
		}
		row := []string{marker, name, ctx.AccountID, ctx.Role, ctx.Region, source}
		if showOrg {
			org := sprintOrg(ctx.Org, orgColors)
			row = []string{marker, name, org, ctx.AccountID, ctx.Role, ctx.Region, source}
		}
		tableData = append(tableData, row)
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	filterNote := ""
	if awsSSOOnly {
		filterNote = " (sso only)"
	} else if awsManualOnly {
		filterNote = " (manual only)"
	}
	fmt.Printf("\nTotal: %d profile(s)%s\n\n", len(contexts), filterNote)

	return nil
}

func setAWS(p *aws.Provider, name string) error {
	contexts, err := p.ListContexts()
	if err != nil {
		return err
	}

	// Find matching profiles
	var matches []provider.Context
	for _, ctx := range contexts {
		if ctx.Name == name {
			// Exact match
			matches = []provider.Context{ctx}
			break
		}
		if strings.Contains(ctx.Name, name) {
			matches = append(matches, ctx)
		}
	}

	if len(matches) == 0 {
		pterm.Error.Printf("No profile matching '%s'\n", name)
		return nil
	}

	if len(matches) > 1 {
		// Multiple matches - show picker
		return pickFromMatches(p, matches)
	}

	// Single match - set it
	return selectProfile(p, matches[0].Name)
}

func interactiveAWS(p *aws.Provider) error {
	contexts, err := p.ListContexts()
	if err != nil {
		return err
	}

	contexts = filterContexts(contexts)
	contexts = filterByOrg(contexts, awsOrg)

	if len(contexts) == 0 {
		pterm.Warning.Println("No AWS profiles found")
		if awsOrg != "" && awsOrg != "all" {
			pterm.FgGray.Printf("Try without --org to see all orgs\n")
		} else if awsSSOOnly {
			pterm.FgGray.Println("Run 'cloudctx aws sync' to fetch profiles from SSO")
		} else if awsManualOnly {
			pterm.FgGray.Println("No manually created profiles found")
		} else {
			pterm.FgGray.Println("Run 'cloudctx aws sync' to fetch profiles from SSO")
		}
		return nil
	}

	return pickFromMatches(p, contexts)
}

func pickFromMatches(p *aws.Provider, contexts []provider.Context) error {
	current, _ := p.CurrentContext()
	currentName := ""
	if current != nil {
		currentName = current.Name
	}

	orgsSeen := make(map[string]bool)
	for _, ctx := range contexts {
		if ctx.Org != "" {
			orgsSeen[ctx.Org] = true
		}
	}
	showOrg := len(orgsSeen) > 1
	if len(orgsSeen) == 1 {
		for k := range orgsSeen {
			if k != "default" {
				showOrg = true
			}
			break
		}
	}
	orgColors := orgColorMap(orgsSeen)

	// [sso]/[manual] from ctx.Managed (uses SSO = sso_session or sso_account_id; see ListContexts in internal/aws/provider.go)
	items := make([]pickItem, len(contexts))
	for i, ctx := range contexts {
		display := awsProfileDisplayName(ctx.Name, ctx.Org)
		sourceWord := "manual"
		sourceStyled := pterm.FgYellow.Sprint("[manual]")
		if ctx.Managed {
			sourceWord = "sso"
			sourceStyled = pterm.FgGreen.Sprint("[sso]")
		}
		marker := "  "
		if ctx.Name == currentName {
			marker = pterm.FgGreen.Sprint("* ") // current row
		}
		if showOrg && ctx.Org != "" {
			orgTag := sprintOrgTag(ctx.Org, orgColors)
			items[i].display = fmt.Sprintf("%s%s %s  %s", marker, orgTag, display, sourceStyled)
		} else {
			items[i].display = fmt.Sprintf("%s%s  %s", marker, display, sourceStyled)
		}
		// Filter against profile/org/account/role/source so any of those tokens match, in any order.
		items[i].search = strings.Join([]string{display, ctx.Org, ctx.AccountID, ctx.Role, sourceWord}, " ")
		items[i].value = ctx.Name
	}

	fmt.Println()
	pterm.Info.Printf("Found %d profiles\n", len(contexts))
	pterm.FgGray.Println("Type to filter (any order) • ↑/↓ to move • Enter to select • Esc to cancel")
	fmt.Println()

	profileName, ok := runPicker(items, 20)
	if !ok {
		return nil
	}

	return selectProfile(p, profileName)
}

func selectProfile(p *aws.Provider, name string) error {
	// Update ~/.aws/config [default] section
	if err := p.SetContext(name); err != nil {
		pterm.Error.Printf("Failed to set profile: %v\n", err)
		return err
	}

	fmt.Println()
	pterm.Success.Printf("Switched to %s\n", pterm.FgCyan.Sprint(name))

	// Check if AWS_PROFILE is set (which would override our default). When the
	// shell integration is active (see "ctx shell-init"), the wrapper resets
	// AWS_PROFILE to this profile right after we return, so they stay in sync and
	// the warning would be misleading.
	if os.Getenv("CLOUDCTX_SHELL_INTEGRATION") == "" {
		if envProfile := os.Getenv("AWS_PROFILE"); envProfile != "" && envProfile != name {
			fmt.Println()
			pterm.Warning.Printf("Note: AWS_PROFILE=%s is set and will override this\n", envProfile)
			pterm.FgGray.Println("Run: unset AWS_PROFILE")
			pterm.FgGray.Println("Or for automatic syncing: eval \"$(ctx shell-init)\" in your shell rc")
		}
	}

	return nil
}

