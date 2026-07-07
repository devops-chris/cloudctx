package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/devops-chris/clihq/ui"
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
	p := getAWSListProvider()

	identity, err := p.WhoAmI()
	if err != nil {
		fmt.Println(ui.Error("Failed to get identity"))
		fmt.Println(ui.Subtle("Are you logged in? Try 'cloudctx aws login'"))
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

	currentProfile := os.Getenv("AWS_PROFILE")

	fmt.Println()
	fmt.Println(ui.Banner("cloudctx", "cloud context switcher"))
	fmt.Println()
	fmt.Println(ui.SectionHeader("AWS Identity"))
	fmt.Println()

	headers := []string{"Property", "Value"}
	var rows [][]string
	if currentProfile != "" {
		rows = append(rows, []string{"Profile", ui.Cyan(currentProfile)})
	}
	rows = append(rows,
		[]string{"Account", identity.AccountID},
		[]string{"User ID", identity.UserID},
		[]string{"ARN", identity.ARN},
		[]string{"Region", identity.Region},
	)

	fmt.Println(ui.Table(headers, rows))
	return nil
}
