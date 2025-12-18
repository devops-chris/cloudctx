package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
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
	pterm.Info.Println("No initialization needed for Azure!")
	fmt.Println()
	pterm.FgGray.Println("Azure uses the Azure CLI for authentication.")
	pterm.FgGray.Println("Just run 'cloudctx azure login' to authenticate.")
	fmt.Println()
	return nil
}

