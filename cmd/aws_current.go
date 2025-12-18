package cmd

import (
	"github.com/spf13/cobra"
)

var awsCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current AWS profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = awsCmd.Flags().Set("current", "true")
		return awsCmd.RunE(cmd, args)
	},
}

func init() {
	awsCmd.AddCommand(awsCurrentCmd)
}

