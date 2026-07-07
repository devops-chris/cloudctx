package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/devops-chris/cloudctx/internal/azure"
	"github.com/devops-chris/cloudctx/internal/provider"
	"github.com/devops-chris/cloudctx/internal/ui"
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

	if azureShowCurrent {
		return showCurrentAzure(p)
	}
	if azureShowList {
		return listAzure(p)
	}
	if len(args) == 1 {
		return setAzure(p, args[0])
	}
	return interactiveAzure(p)
}

func showCurrentAzure(p *azure.Provider) error {
	current, err := p.CurrentContext()
	if err != nil {
		return err
	}
	if current == nil {
		fmt.Println(ui.Warning("No Azure subscription set"))
		fmt.Println(ui.Subtle("Run 'cloudctx azure login' to authenticate"))
		return nil
	}
	fmt.Println(current.Name)
	return nil
}

func listAzure(p *azure.Provider) error {
	contexts, err := p.ListContexts()
	if err != nil {
		fmt.Println(ui.Error("Failed to list subscriptions"))
		fmt.Println(ui.Subtle("Run 'cloudctx azure login' to authenticate"))
		return err
	}
	if len(contexts) == 0 {
		fmt.Println(ui.Warning("No Azure subscriptions found"))
		fmt.Println(ui.Subtle("Run 'cloudctx azure login' to authenticate"))
		return nil
	}

	fmt.Println()
	fmt.Println(ui.SectionHeader("Azure Subscriptions"))
	fmt.Println()

	var rows [][]string
	for _, ctx := range contexts {
		marker := " "
		name := ctx.Name
		if ctx.Active {
			marker = "*"
			name = ui.Green(ctx.Name)
		}
		rows = append(rows, []string{marker, name, ctx.AccountID})
	}
	fmt.Println(ui.Table([]string{"", "Subscription", "Subscription ID"}, rows))
	fmt.Printf("\nTotal: %d subscription(s)\n\n", len(contexts))

	return nil
}

func setAzure(p *azure.Provider, name string) error {
	contexts, err := p.ListContexts()
	if err != nil {
		return err
	}

	var matches []provider.Context
	for _, ctx := range contexts {
		if ctx.Name == name || ctx.AccountID == name {
			matches = []provider.Context{ctx}
			break
		}
		if strings.Contains(strings.ToLower(ctx.Name), strings.ToLower(name)) {
			matches = append(matches, ctx)
		}
	}

	if len(matches) == 0 {
		fmt.Println(ui.Errorf("No subscription matching '%s'", name))
		return nil
	}
	if len(matches) > 1 {
		return pickAzureFromMatches(p, matches)
	}
	return selectAzureSubscription(p, matches[0].Name)
}

func interactiveAzure(p *azure.Provider) error {
	contexts, err := p.ListContexts()
	if err != nil {
		fmt.Println(ui.Error("Failed to list subscriptions"))
		fmt.Println(ui.Subtle("Run 'cloudctx azure login' to authenticate"))
		return err
	}
	if len(contexts) == 0 {
		fmt.Println(ui.Warning("No Azure subscriptions found"))
		fmt.Println(ui.Subtle("Run 'cloudctx azure login' to authenticate"))
		return nil
	}
	return pickAzureFromMatches(p, contexts)
}

func pickAzureFromMatches(p *azure.Provider, contexts []provider.Context) error {
	current, _ := p.CurrentContext()
	currentName := ""
	if current != nil {
		currentName = current.Name
	}

	items := make([]pickItem, len(contexts))
	for i, ctx := range contexts {
		if ctx.Name == currentName {
			items[i].display = ui.Green("* ") + ctx.Name
		} else {
			items[i].display = "  " + ctx.Name
		}
		items[i].search = ctx.Name
		items[i].value = ctx.Name
	}

	fmt.Println()
	fmt.Println(ui.Infof("Found %d subscriptions", len(contexts)))
	fmt.Println(ui.Subtle("Type to filter (any order) • ↑/↓ to move • Enter to select • Esc to cancel"))
	fmt.Println()

	subName, ok := runPicker(items, 20)
	if !ok {
		return nil
	}
	return selectAzureSubscription(p, subName)
}

func selectAzureSubscription(p *azure.Provider, name string) error {
	if err := p.SetContext(name); err != nil {
		fmt.Println(ui.Errorf("Failed to set subscription: %v", err))
		return err
	}

	fmt.Println()
	fmt.Println(ui.Successf("Switched to %s", ui.Highlight(name)))

	if envSub := os.Getenv("AZURE_SUBSCRIPTION_ID"); envSub != "" {
		fmt.Println()
		fmt.Println(ui.Warning("Note: AZURE_SUBSCRIPTION_ID is set and may override this"))
		fmt.Println(ui.Subtle("Run: unset AZURE_SUBSCRIPTION_ID"))
	}

	return nil
}
