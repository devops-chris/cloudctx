package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/devops-chris/cloudctx/internal/azure"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var azureWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current Azure identity",
	Long:  `Display information about the current Azure subscription and user.`,
	RunE:  runAzureWhoami,
}

var azureWhoamiJSON bool

func init() {
	azureCmd.AddCommand(azureWhoamiCmd)
	azureWhoamiCmd.Flags().BoolVar(&azureWhoamiJSON, "json", false, "output as JSON")
}

func runAzureWhoami(cmd *cobra.Command, args []string) error {
	p := azure.NewProvider(cfg.Azure.DefaultLocation)

	identity, err := p.WhoAmI()
	if err != nil {
		pterm.Error.Println("Failed to get identity")
		pterm.FgGray.Println("Run 'cloudctx azure login' to authenticate")
		return err
	}

	if azureWhoamiJSON {
		data, _ := json.MarshalIndent(identity, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Println()
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println("Azure Identity")

	tableData := pterm.TableData{
		{"Field", "Value"},
		{"Subscription", identity.AccountName},
		{"Subscription ID", identity.AccountID},
		{"User", identity.UserID},
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	fmt.Println()

	return nil
}

