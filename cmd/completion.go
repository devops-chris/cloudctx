package cmd

import (
	"github.com/devops-chris/cloudctx/internal/azure"
	"github.com/spf13/cobra"
)

func init() {
	// AWS profile names: cloudctx aws <TAB>
	awsCmd.ValidArgsFunction = completeAWSProfiles

	// Azure subscription names: cloudctx azure <TAB>
	azureCmd.ValidArgsFunction = completeAzureSubscriptions

	// Org sub-commands
	awsOrgRemoveCmd.ValidArgsFunction = completeOrgNames
	awsOrgRenameCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first argument (the old name); second is a free-form new name
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return completeOrgNames(cmd, args, toComplete)
	}
}

func completeAWSProfiles(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	p := getAWSListProvider()
	contexts, err := p.ListContexts()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, 0, len(contexts))
	for _, ctx := range contexts {
		names = append(names, ctx.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func completeAzureSubscriptions(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	p := azure.NewProvider(cfg.Azure.DefaultLocation)
	contexts, err := p.ListContexts()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, 0, len(contexts))
	for _, ctx := range contexts {
		names = append(names, ctx.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func completeOrgNames(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	orgs := cfg.AWSOrgs()
	names := make([]string, 0, len(orgs)+1)
	names = append(names, "all")
	for k := range orgs {
		names = append(names, k)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
