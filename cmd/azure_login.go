package cmd

import (
	"fmt"

	"github.com/devops-chris/cloudctx/internal/azure"
	"github.com/devops-chris/cloudctx/internal/ui"
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
	fmt.Println(ui.Info("Opening browser for Azure login..."))
	fmt.Println()

	if err := p.Login(); err != nil {
		fmt.Println(ui.Errorf("Login failed: %v", err))
		return err
	}

	fmt.Println()
	fmt.Println(ui.Success("Successfully logged in to Azure"))
	fmt.Println(ui.Subtle("Run 'cloudctx azure' to select a subscription"))

	return nil
}
