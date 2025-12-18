package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
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
	pterm.Info.Println("Sync is not needed for Azure!")
	fmt.Println()
	pterm.FgGray.Println("Unlike AWS, Azure subscriptions are fetched live from Azure CLI.")
	pterm.FgGray.Println("Just run 'cloudctx azure' to see and switch subscriptions.")
	fmt.Println()
	return nil
}

