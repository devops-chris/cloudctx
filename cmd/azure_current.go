package cmd

import (
	"github.com/spf13/cobra"
)

var azureCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current Azure subscription",
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = azureCmd.Flags().Set("current", "true")
		return azureCmd.RunE(cmd, args)
	},
}

func init() {
	azureCmd.AddCommand(azureCurrentCmd)
}

