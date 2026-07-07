package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/devops-chris/clihq/ui"
	"github.com/devops-chris/cloudctx/internal/azure"
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
		fmt.Println(ui.Error("Failed to get identity"))
		fmt.Println(ui.Subtle("Run 'cloudctx azure login' to authenticate"))
		return err
	}

	if azureWhoamiJSON {
		data, _ := json.MarshalIndent(identity, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Println()
	fmt.Println(ui.SectionHeader("Azure Identity"))
	fmt.Println()

	rows := [][]string{
		{"Subscription", identity.AccountName},
		{"Subscription ID", identity.AccountID},
		{"User", identity.UserID},
	}
	fmt.Println(ui.Table([]string{"Field", "Value"}, rows))
	fmt.Println()

	return nil
}
