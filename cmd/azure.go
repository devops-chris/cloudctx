package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/devops-chris/cloudctx/internal/azure"
	"github.com/devops-chris/cloudctx/internal/provider"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var azureCmd = &cobra.Command{
	Use:   "azure [subscription]",
	Short: "Manage Azure subscriptions",
	Long: `Manage Azure subscriptions and authentication.

Without arguments, opens an interactive subscription picker.
With a subscription name, sets that subscription directly.

Examples:
  cloudctx azure                    # Interactive picker
  cloudctx azure my-subscription    # Set specific subscription
  cloudctx azure -c                 # Show current subscription
  cloudctx azure -l                 # List all subscriptions`,
	Aliases: []string{"az"},
	Args:    cobra.MaximumNArgs(1),
	RunE:    runAzure,
}

var (
	azureShowCurrent bool
	azureShowList    bool
)

func init() {
	rootCmd.AddCommand(azureCmd)

	azureCmd.Flags().BoolVarP(&azureShowCurrent, "current", "c", false, "show current subscription")
	azureCmd.Flags().BoolVarP(&azureShowList, "list", "l", false, "list all subscriptions")
}

func runAzure(cmd *cobra.Command, args []string) error {
	p := azure.NewProvider(cfg.Azure.DefaultLocation)

	// Show current subscription
	if azureShowCurrent {
		return showCurrentAzure(p)
	}

	// List all subscriptions
	if azureShowList {
		return listAzure(p)
	}

	// Set specific subscription
	if len(args) == 1 {
		return setAzure(p, args[0])
	}

	// Interactive picker
	return interactiveAzure(p)
}

func showCurrentAzure(p *azure.Provider) error {
	current, err := p.CurrentContext()
	if err != nil {
		return err
	}

	if current == nil {
		pterm.Warning.Println("No Azure subscription set")
		pterm.FgGray.Println("Run 'cloudctx azure login' to authenticate")
		return nil
	}

	fmt.Println(current.Name)
	return nil
}

func listAzure(p *azure.Provider) error {
	contexts, err := p.ListContexts()
	if err != nil {
		pterm.Error.Println("Failed to list subscriptions")
		pterm.FgGray.Println("Run 'cloudctx azure login' to authenticate")
		return err
	}

	if len(contexts) == 0 {
		pterm.Warning.Println("No Azure subscriptions found")
		pterm.FgGray.Println("Run 'cloudctx azure login' to authenticate")
		return nil
	}

	fmt.Println()
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println("Azure Subscriptions")

	tableData := pterm.TableData{
		{"", "Subscription", "Subscription ID"},
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
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	fmt.Printf("\nTotal: %d subscription(s)\n\n", len(contexts))

	return nil
}

func setAzure(p *azure.Provider, name string) error {
	contexts, err := p.ListContexts()
	if err != nil {
		return err
	}

	// Find matching subscriptions
	var matches []provider.Context
	for _, ctx := range contexts {
		if ctx.Name == name || ctx.AccountID == name {
			// Exact match
			matches = []provider.Context{ctx}
			break
		}
		if strings.Contains(strings.ToLower(ctx.Name), strings.ToLower(name)) {
			matches = append(matches, ctx)
		}
	}

	if len(matches) == 0 {
		pterm.Error.Printf("No subscription matching '%s'\n", name)
		return nil
	}

	if len(matches) > 1 {
		// Multiple matches - show picker
		return pickAzureFromMatches(p, matches)
	}

	// Single match - set it
	return selectAzureSubscription(p, matches[0].Name)
}

func interactiveAzure(p *azure.Provider) error {
	contexts, err := p.ListContexts()
	if err != nil {
		pterm.Error.Println("Failed to list subscriptions")
		pterm.FgGray.Println("Run 'cloudctx azure login' to authenticate")
		return err
	}

	if len(contexts) == 0 {
		pterm.Warning.Println("No Azure subscriptions found")
		pterm.FgGray.Println("Run 'cloudctx azure login' to authenticate")
		return nil
	}

	return pickAzureFromMatches(p, contexts)
}

func pickAzureFromMatches(p *azure.Provider, contexts []provider.Context) error {
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
	pterm.Info.Printf("Found %d subscriptions\n", len(contexts))
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

	// Extract subscription name (remove marker)
	subName := strings.TrimPrefix(selected, "* ")
	subName = strings.TrimPrefix(subName, "  ")

	return selectAzureSubscription(p, subName)
}

func selectAzureSubscription(p *azure.Provider, name string) error {
	if err := p.SetContext(name); err != nil {
		pterm.Error.Printf("Failed to set subscription: %v\n", err)
		return err
	}

	fmt.Println()
	pterm.Success.Printf("Switched to %s\n", pterm.FgCyan.Sprint(name))

	// Check if AZURE_SUBSCRIPTION_ID is set
	if envSub := os.Getenv("AZURE_SUBSCRIPTION_ID"); envSub != "" {
		fmt.Println()
		pterm.Warning.Printf("Note: AZURE_SUBSCRIPTION_ID is set and may override this\n")
		pterm.FgGray.Println("Run: unset AZURE_SUBSCRIPTION_ID")
	}

	return nil
}

