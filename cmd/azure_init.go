package cmd

import (
	"fmt"

	"github.com/devops-chris/clihq/ui"
	"github.com/spf13/cobra"
)

var azureInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Not needed for Azure",
	Long:  `Azure uses Azure CLI for authentication - no additional setup required.`,
	RunE:  runAzureInit,
}

func init() {
	azureCmd.AddCommand(azureInitCmd)
}

func runAzureInit(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(ui.Info("No initialization needed for Azure!"))
	fmt.Println()
	fmt.Println(ui.Subtle("Azure uses the Azure CLI for authentication."))
	fmt.Println(ui.Subtle("Just run 'cloudctx azure login' to authenticate."))
	fmt.Println()
	return nil
}
