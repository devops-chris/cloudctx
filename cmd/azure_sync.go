package cmd

import (
	"fmt"

	"github.com/devops-chris/cloudctx/internal/ui"
	"github.com/spf13/cobra"
)

var azureSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Not needed for Azure",
	Long:  `Azure subscriptions are fetched live - no sync required.`,
	RunE:  runAzureSync,
}

func init() {
	azureCmd.AddCommand(azureSyncCmd)
}

func runAzureSync(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(ui.Info("Sync is not needed for Azure!"))
	fmt.Println()
	fmt.Println(ui.Subtle("Unlike AWS, Azure subscriptions are fetched live from Azure CLI."))
	fmt.Println(ui.Subtle("Just run 'cloudctx azure' to see and switch subscriptions."))
	fmt.Println()
	return nil
}
