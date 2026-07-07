package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/devops-chris/cloudctx/internal/aws"
	"github.com/devops-chris/cloudctx/internal/provider"
	"github.com/devops-chris/cloudctx/internal/ui"
	"github.com/spf13/cobra"
)

// orgPalette assigns distinct colors to orgs (deterministic by sorted name).
var orgPalette = []lipgloss.Style{
	lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0097A7", Dark: "#4DD0E1"}),
	lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#7B1FA2", Dark: "#CE93D8"}),
	lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1565C0", Dark: "#90CAF9"}),
	lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00838F", Dark: "#80DEEA"}),
	lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#6A1B9A", Dark: "#E040FB"}),
	lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1976D2", Dark: "#64B5F6"}),
}

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
		return ui.Subtle("-")
	}
	idx, ok := colorIndex[org]
	if !ok {
		return ui.Subtle(org)
	}
	return orgPalette[idx].Render(org)
}

func sprintOrgTag(org string, colorIndex map[string]int) string {
	if org == "" {
		return ui.Subtle("[]")
	}
	idx, ok := colorIndex[org]
	if !ok {
		return ui.Subtle("[" + org + "]")
	}
	return orgPalette[idx].Render("[" + org + "]")
}

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
	awsOrg         string
)

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().BoolVarP(&awsShowCurrent, "current", "c", false, "show current profile")
	awsCmd.Flags().BoolVarP(&awsShowList, "list", "l", false, "list all profiles")
	awsCmd.Flags().BoolVar(&awsSSOOnly, "sso", false, "show only SSO-synced profiles")
	awsCmd.Flags().BoolVar(&awsManualOnly, "manual", false, "show only manually created profiles")
}

func runAWS(cmd *cobra.Command, args []string) error {
	p := getAWSListProvider()

	if awsShowCurrent {
		return showCurrentAWS(p)
	}
	if awsShowList {
		return listAWS(p)
	}
	if len(args) == 1 {
		return setAWS(p, args[0])
	}
	return interactiveAWS(p)
}

func showCurrentAWS(p *aws.Provider) error {
	current, err := p.CurrentContext()
	if err != nil {
		return err
	}
	if current == nil {
		fmt.Println(ui.Warning("No AWS profile set"))
		fmt.Println(ui.Subtle("Set one with: cloudctx aws <profile>"))
		return nil
	}
	fmt.Println(current.Name)
	return nil
}

func filterContexts(contexts []provider.Context) []provider.Context {
	if !awsSSOOnly && !awsManualOnly {
		return contexts
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
		fmt.Println(ui.Warning("No AWS profiles found"))
		if awsOrg != "" && awsOrg != "all" {
			fmt.Println(ui.Subtle("Try without --org to see all orgs, or use --org all"))
		} else if awsSSOOnly {
			fmt.Println(ui.Subtle("Run 'cloudctx aws sync' to fetch profiles from SSO"))
		} else if awsManualOnly {
			fmt.Println(ui.Subtle("No manually created profiles found"))
		} else {
			fmt.Println(ui.Subtle("Run 'cloudctx aws sync' to fetch profiles from SSO"))
		}
		return nil
	}

	fmt.Println()
	fmt.Println(ui.Banner())
	fmt.Println()
	headerTitle := "AWS Profiles"
	if awsOrg != "" && awsOrg != "all" {
		headerTitle = "AWS Profiles (org: " + awsOrg + ")"
	}
	fmt.Println(ui.SectionHeader(headerTitle))
	fmt.Println()

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

	const maxProfileLen = 40
	var rows [][]string
	for _, ctx := range contexts {
		marker := " "
		name := awsProfileDisplayName(ctx.Name, ctx.Org)
		if len(name) > maxProfileLen {
			name = name[:maxProfileLen-3] + "..."
		}
		if ctx.Active {
			marker = "*"
			name = ui.Green(name)
		}
		source := ui.Yellow("manual")
		if ctx.Managed {
			source = ui.Cyan("sso")
		}
		row := []string{marker, name, ctx.AccountID, ctx.Role, ctx.Region, source}
		if showOrg {
			row = []string{marker, name, sprintOrg(ctx.Org, orgColors), ctx.AccountID, ctx.Role, ctx.Region, source}
		}
		rows = append(rows, row)
	}

	fmt.Println(ui.Table(headers, rows))

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

	var matches []provider.Context
	for _, ctx := range contexts {
		if ctx.Name == name {
			matches = []provider.Context{ctx}
			break
		}
		if strings.Contains(ctx.Name, name) {
			matches = append(matches, ctx)
		}
	}

	if len(matches) == 0 {
		fmt.Println(ui.Errorf("No profile matching '%s'", name))
		return nil
	}
	if len(matches) > 1 {
		return pickFromMatches(p, matches)
	}
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
		fmt.Println(ui.Warning("No AWS profiles found"))
		if awsOrg != "" && awsOrg != "all" {
			fmt.Println(ui.Subtle("Try without --org to see all orgs"))
		} else if awsSSOOnly {
			fmt.Println(ui.Subtle("Run 'cloudctx aws sync' to fetch profiles from SSO"))
		} else if awsManualOnly {
			fmt.Println(ui.Subtle("No manually created profiles found"))
		} else {
			fmt.Println(ui.Subtle("Run 'cloudctx aws sync' to fetch profiles from SSO"))
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

	items := make([]pickItem, len(contexts))
	for i, ctx := range contexts {
		display := awsProfileDisplayName(ctx.Name, ctx.Org)
		sourceWord := "manual"
		sourceStyled := ui.Yellow("[manual]")
		if ctx.Managed {
			sourceWord = "sso"
			sourceStyled = ui.Green("[sso]")
		}
		marker := "  "
		if ctx.Name == currentName {
			marker = ui.Green("* ")
		}
		if showOrg && ctx.Org != "" {
			orgTag := sprintOrgTag(ctx.Org, orgColors)
			items[i].display = fmt.Sprintf("%s%s %s  %s", marker, orgTag, display, sourceStyled)
		} else {
			items[i].display = fmt.Sprintf("%s%s  %s", marker, display, sourceStyled)
		}
		items[i].search = strings.Join([]string{display, ctx.Org, ctx.AccountID, ctx.Role, sourceWord}, " ")
		items[i].value = ctx.Name
	}

	fmt.Println()
	fmt.Println(ui.Infof("Found %d profiles", len(contexts)))
	fmt.Println(ui.Subtle("Type to filter (any order) • ↑/↓ to move • Enter to select • Esc to cancel"))
	fmt.Println()

	profileName, ok := runPicker(items, 20)
	if !ok {
		return nil
	}
	return selectProfile(p, profileName)
}

func selectProfile(p *aws.Provider, name string) error {
	if err := p.SetContext(name); err != nil {
		fmt.Println(ui.Errorf("Failed to set profile: %v", err))
		return err
	}

	fmt.Println()
	fmt.Println(ui.Successf("Switched to %s", ui.Highlight(name)))

	if os.Getenv("CLOUDCTX_SHELL_INTEGRATION") == "" {
		if envProfile := os.Getenv("AWS_PROFILE"); envProfile != "" && envProfile != name {
			fmt.Println()
			fmt.Println(ui.Warningf("Note: AWS_PROFILE=%s is set and will override this", envProfile))
			fmt.Println(ui.Subtle("Run: unset AWS_PROFILE"))
			fmt.Println(ui.Subtle(`Or for automatic syncing: eval "$(ctx shell-init)" in your shell rc`))
		}
	}

	return nil
}
