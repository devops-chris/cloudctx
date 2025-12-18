package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/devops-chris/cloudctx/internal/aws"
	"github.com/devops-chris/cloudctx/internal/provider"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var awsCmd = &cobra.Command{
	Use:   "aws [profile]",
	Short: "Manage AWS profiles",
	Long: `Manage AWS profiles and SSO authentication.

Without arguments, opens an interactive profile picker.
With a profile name, sets that profile directly.

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
)

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().BoolVarP(&awsShowCurrent, "current", "c", false, "show current profile")
	awsCmd.Flags().BoolVarP(&awsShowList, "list", "l", false, "list all profiles")
}

func runAWS(cmd *cobra.Command, args []string) error {
	p := aws.NewProvider(cfg.AWS.SSOStartURL, cfg.AWS.SSORegion, cfg.AWS.DefaultRegion)

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

func listAWS(p *aws.Provider) error {
	contexts, err := p.ListContexts()
	if err != nil {
		return err
	}

	if len(contexts) == 0 {
		pterm.Warning.Println("No AWS profiles found")
		pterm.FgGray.Println("Run 'cloudctx aws sync' to fetch profiles from SSO")
		return nil
	}

	fmt.Println()
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println("AWS Profiles")

	tableData := pterm.TableData{
		{"", "Profile", "Account ID", "Role", "Region"},
	}

	for _, ctx := range contexts {
		marker := " "
		name := ctx.Name
		if ctx.Active {
			marker = "*"
			name = pterm.FgGreen.Sprint(ctx.Name)
		}
		tableData = append(tableData, []string{
			marker,
			name,
			ctx.AccountID,
			ctx.Role,
			ctx.Region,
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	fmt.Printf("\nTotal: %d profile(s)\n\n", len(contexts))

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

	if len(contexts) == 0 {
		pterm.Warning.Println("No AWS profiles found")
		pterm.FgGray.Println("Run 'cloudctx aws sync' to fetch profiles from SSO")
		return nil
	}

	return pickFromMatches(p, contexts)
}

func pickFromMatches(p *aws.Provider, contexts []provider.Context) error {
	// Get current to mark it
	current, _ := p.CurrentContext()
	currentName := ""
	if current != nil {
		currentName = current.Name
	}

	// Build options
	options := make([]string, len(contexts))
	for i, ctx := range contexts {
		if ctx.Name == currentName {
			options[i] = fmt.Sprintf("* %s", ctx.Name)
		} else {
			options[i] = fmt.Sprintf("  %s", ctx.Name)
		}
	}

	fmt.Println()
	pterm.Info.Printf("Found %d profiles\n", len(contexts))
	pterm.FgGray.Println("Type to filter • Enter to select • Ctrl+C to cancel")
	fmt.Println()

	selected, err := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithFilter(true).
		WithMaxHeight(20).
		Show()

	if err != nil {
		return nil // User cancelled
	}

	// Extract profile name (remove marker)
	profileName := strings.TrimPrefix(selected, "* ")
	profileName = strings.TrimPrefix(profileName, "  ")

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

	// Check if AWS_PROFILE is set (which would override our default)
	if envProfile := os.Getenv("AWS_PROFILE"); envProfile != "" && envProfile != name {
		fmt.Println()
		pterm.Warning.Printf("Note: AWS_PROFILE=%s is set and will override this\n", envProfile)
		pterm.FgGray.Println("Run: unset AWS_PROFILE")
	}

	return nil
}

