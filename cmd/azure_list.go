package cmd

import (
	"github.com/spf13/cobra"
)

var azureListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all Azure subscriptions",
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = azureCmd.Flags().Set("list", "true")
		return azureCmd.RunE(cmd, args)
	},
}

func init() {
	azureCmd.AddCommand(azureListCmd)
}

