package cmd

import (
	"fmt"

	"github.com/devops-chris/cloudctx/internal/azure"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var azureLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Azure",
	Long:  `Authenticate with Azure using az login (opens browser).`,
	RunE:  runAzureLogin,
}

func init() {
	azureCmd.AddCommand(azureLoginCmd)
}

func runAzureLogin(cmd *cobra.Command, args []string) error {
	p := azure.NewProvider(cfg.Azure.DefaultLocation)

	fmt.Println()
	pterm.Info.Println("Opening browser for Azure login...")
	fmt.Println()

	if err := p.Login(); err != nil {
		pterm.Error.Printf("Login failed: %v\n", err)
		return err
	}

	fmt.Println()
	pterm.Success.Println("Successfully logged in to Azure")
	pterm.FgGray.Println("Run 'cloudctx azure' to select a subscription")

	return nil
}

