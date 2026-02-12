package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/devops-chris/cloudctx/internal/config"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check that cloudctx and AWS are configured correctly",
	Long: `Verify your cloudctx config, default provider, AWS orgs, current profile,
and that you can call AWS (e.g. credentials/SSO token are valid).

Use this to confirm things are set up right before running other commands.`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println()
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println("cloudctx doctor")

	configPath := cfgFile
	if configPath == "" {
		configPath = config.ConfigPath()
	}
	configExists := false
	if st, err := os.Stat(configPath); err == nil && st.Mode().IsRegular() {
		configExists = true
	}

	pterm.Info.Println("Config")
	table := pterm.TableData{
		{"Check", "Result"},
		{"Config file", configPath},
		{"Config exists", yesNo(configExists)},
		{"default_cloud", cfg.DefaultCloud},
	}
	if !configExists {
		table = append(table, []string{"", pterm.FgYellow.Sprint("Run 'cloudctx aws init' or create config")})
	}
	_ = pterm.DefaultTable.WithHasHeader().WithData(table).Render()
	fmt.Println()

	// AWS-specific checks when default is AWS or when we have AWS orgs
	orgs := cfg.AWSOrgs()
	defaultOrg := cfg.AWSDefaultOrg()
	if cfg.DefaultCloud == "azure" || cfg.DefaultCloud == "az" {
		pterm.Info.Println("Default cloud is Azure; skipping AWS checks (use 'cloudctx aws ...' for AWS).")
		fmt.Println()
		return nil
	}

	pterm.Info.Println("AWS")
	awsTable := pterm.TableData{
		{"Check", "Result"},
		{"AWS CLI in PATH", yesNo(awsCLIInPath())},
		{"AWS orgs configured", yesNo(len(orgs) > 0)},
	}
	if len(orgs) > 0 {
		names := make([]string, 0, len(orgs))
		for k := range orgs {
			names = append(names, k)
		}
		sort.Strings(names)
		def := defaultOrg
		if def == "" {
			def = "(none set)"
		}
		awsTable = append(awsTable, []string{"Org names", fmt.Sprintf("%v", names)})
		awsTable = append(awsTable, []string{"Default org", def})
	}
	_ = pterm.DefaultTable.WithHasHeader().WithData(awsTable).Render()
	fmt.Println()

	p := getAWSListProvider()
	contexts, err := p.ListContexts()
	if err != nil {
		pterm.Warning.Printf("Could not list profiles: %v\n", err)
		fmt.Println()
		return nil
	}
	current, _ := p.CurrentContext()
	currentName := ""
	currentOrg := ""
	if current != nil {
		currentName = current.Name
		currentOrg = current.Org
	}

	pterm.Info.Println("Current context")
	ctxTable := pterm.TableData{
		{"Check", "Result"},
		{"Profiles in ~/.aws", fmt.Sprintf("%d", len(contexts))},
		{"Current profile", currentName},
	}
	if currentOrg != "" {
		ctxTable = append(ctxTable, []string{"Current org", currentOrg})
	}
	if currentName == "" {
		ctxTable = append(ctxTable, []string{"", pterm.FgYellow.Sprint("No profile set — run 'cloudctx' or 'cloudctx aws <profile>'")})
	}
	_ = pterm.DefaultTable.WithHasHeader().WithData(ctxTable).Render()
	fmt.Println()

	pterm.Info.Println("Credentials / SSO")
	_, err = p.WhoAmI()
	if err != nil {
		pterm.Error.Println("Cannot get AWS identity (credentials or SSO token invalid/missing)")
		pterm.FgGray.Println("Fix: cloudctx login   (or cloudctx login --org <org>)")
		pterm.FgGray.Printf("Error: %v\n", err)
		fmt.Println()
		return nil
	}
	pterm.Success.Println("AWS identity OK — you can call AWS APIs.")
	fmt.Println()

	return nil
}

func yesNo(b bool) string {
	if b {
		return pterm.FgGreen.Sprint("yes")
	}
	return pterm.FgRed.Sprint("no")
}

func awsCLIInPath() bool {
	_, err := exec.LookPath("aws")
	return err == nil
}
