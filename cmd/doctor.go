package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/devops-chris/cloudctx/internal/config"
	"github.com/devops-chris/cloudctx/internal/ui"
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
	fmt.Println(ui.Banner())

	// ── Config ───────────────────────────────────────────────────────────────

	fmt.Print(ui.CheckSection("Config"))

	configPath := cfgFile
	if configPath == "" {
		configPath = config.ConfigPath()
	}
	configExists := false
	if st, err := os.Stat(configPath); err == nil && st.Mode().IsRegular() {
		configExists = true
	}

	if configExists {
		fmt.Println(ui.CheckPass("Config file found: " + configPath))
	} else {
		fmt.Println(ui.CheckFail("Config file not found: "+configPath, "Run 'cloudctx aws init' to create it"))
	}

	if cfg.DefaultCloud != "" {
		fmt.Println(ui.CheckPass("default_cloud: " + cfg.DefaultCloud))
	} else {
		fmt.Println(ui.CheckWarn("default_cloud not set", "Add default_cloud: aws (or azure) to your config"))
	}

	if cfg.DefaultCloud == "azure" || cfg.DefaultCloud == "az" {
		fmt.Println()
		fmt.Println(ui.CheckWarn("Default cloud is Azure; skipping AWS checks", "Use 'cloudctx aws ...' for AWS-specific commands"))
		fmt.Println()
		return nil
	}

	// ── AWS ──────────────────────────────────────────────────────────────────

	fmt.Print(ui.CheckSection("AWS"))

	if awsCLIInPath() {
		fmt.Println(ui.CheckPass("AWS CLI found in PATH"))
	} else {
		fmt.Println(ui.CheckFail("AWS CLI not found in PATH", "Install with: brew install awscli"))
	}

	orgs := cfg.AWSOrgs()
	defaultOrg := cfg.AWSDefaultOrg()
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
		fmt.Println(ui.CheckPass(fmt.Sprintf("%d org(s) configured: %s", len(orgs), strings.Join(names, ", "))))
		fmt.Println(ui.CheckPass("Default org: " + def))
	} else {
		fmt.Println(ui.CheckFail("No AWS orgs configured", "Run 'cloudctx aws init' or 'cloudctx aws org add'"))
	}

	// ── Current context ───────────────────────────────────────────────────────

	fmt.Print(ui.CheckSection("Current context"))

	p := getAWSListProvider()
	contexts, err := p.ListContexts()
	if err != nil {
		fmt.Println(ui.CheckWarn(fmt.Sprintf("Could not list profiles: %v", err), ""))
		fmt.Println()
		return nil
	}
	fmt.Println(ui.CheckPass(fmt.Sprintf("%d profile(s) in ~/.aws/config", len(contexts))))

	current, _ := p.CurrentContext()
	if current != nil {
		fmt.Println(ui.CheckPass("Current profile: " + current.Name))
		if current.Org != "" {
			fmt.Println(ui.CheckPass("Current org: " + current.Org))
		}
	} else {
		fmt.Println(ui.CheckWarn("No profile set", "Run 'cloudctx aws <profile>' to select one"))
	}

	// ── Credentials / SSO ────────────────────────────────────────────────────

	fmt.Print(ui.CheckSection("Credentials / SSO"))

	_, err = p.WhoAmI()
	if err != nil {
		fmt.Println(ui.CheckFail(
			"Cannot get AWS identity — credentials or SSO token invalid/missing",
			"Fix: cloudctx login   (or cloudctx login --org <org>)\nError: "+err.Error(),
		))
	} else {
		fmt.Println(ui.CheckPass("AWS identity verified — you can call AWS APIs"))
	}

	fmt.Println()
	return nil
}

func awsCLIInPath() bool {
	_, err := exec.LookPath("aws")
	return err == nil
}
