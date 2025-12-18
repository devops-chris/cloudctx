package cmd

import (
	"github.com/spf13/cobra"
)

var awsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all AWS profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = awsCmd.Flags().Set("list", "true")
		return awsCmd.RunE(cmd, args)
	},
}

func init() {
	awsCmd.AddCommand(awsListCmd)
}

