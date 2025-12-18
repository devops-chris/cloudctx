package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/devops-chris/cloudctx/internal/aws"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var awsWhoamiJSON bool

var awsWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current AWS identity",
	Long: `Display the current AWS identity.

Shows the account, user, and ARN of the currently authenticated AWS identity.
This is equivalent to 'aws sts get-caller-identity'.

Examples:
  cloudctx aws whoami
  cloudctx aws whoami --json`,
	RunE: runAWSWhoami,
}

func init() {
	awsCmd.AddCommand(awsWhoamiCmd)
	awsWhoamiCmd.Flags().BoolVar(&awsWhoamiJSON, "json", false, "output as JSON")
}

func runAWSWhoami(cmd *cobra.Command, args []string) error {
	p := aws.NewProvider(cfg.AWS.SSOStartURL, cfg.AWS.SSORegion, cfg.AWS.DefaultRegion)

	identity, err := p.WhoAmI()
	if err != nil {
		pterm.Error.Println("Failed to get identity")
		pterm.FgGray.Println("Are you logged in? Try 'cloudctx aws login'")
		return err
	}

	if awsWhoamiJSON {
		data, err := json.MarshalIndent(identity, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Get current profile
	currentProfile := os.Getenv("AWS_PROFILE")

	fmt.Println()
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println("AWS Identity")

	tableData := pterm.TableData{
		{"Property", "Value"},
	}

	if currentProfile != "" {
		tableData = append(tableData, []string{"Profile", pterm.FgCyan.Sprint(currentProfile)})
	}
	tableData = append(tableData, []string{"Account", identity.AccountID})
	tableData = append(tableData, []string{"User ID", identity.UserID})
	tableData = append(tableData, []string{"ARN", identity.ARN})
	tableData = append(tableData, []string{"Region", identity.Region})

	_ = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
	fmt.Println()

	return nil
}

