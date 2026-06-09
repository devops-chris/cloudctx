package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// shellInitCmd prints a shell function that keeps AWS_PROFILE in sync with the
// profile cloudctx switches to. A subprocess can't change its parent shell's
// environment, so prompts like Starship (which read AWS_PROFILE) never see a
// plain `ctx aws <profile>` switch. This wrapper closes that gap: after the real
// binary runs, it exports AWS_PROFILE from the state file cloudctx already writes
// on every switch (~/.config/cloudctx/aws_current).
//
// Usage: add to your shell rc file:
//
//	eval "$(ctx shell-init zsh)"   # or bash
var shellInitCmd = &cobra.Command{
	Use:   "shell-init [zsh|bash]",
	Short: "Print shell integration so AWS_PROFILE follows the active profile (for Starship etc.)",
	Long: `Print a shell function that keeps the AWS_PROFILE environment variable in sync
with the profile cloudctx switches to.

cloudctx switches profiles by rewriting the [default] section of ~/.aws/config, so
it works without any shell setup. But prompt tools like Starship read the AWS_PROFILE
environment variable to show the active profile, and a CLI can't set an env var in the
shell that launched it. This integration installs a small shell function that does it
for you.

Add this to your ~/.zshrc (or ~/.bashrc):

  eval "$(ctx shell-init zsh)"     # zsh
  eval "$(ctx shell-init bash)"    # bash

Then open a new shell. After that, 'ctx aws <profile>' (or the interactive picker)
updates AWS_PROFILE automatically and Starship reflects it immediately.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"zsh", "bash"},
	// No cloud-provider work here; just emit shell code, so don't load config.
	RunE: runShellInit,
}

func init() {
	rootCmd.AddCommand(shellInitCmd)
}

func runShellInit(cmd *cobra.Command, args []string) error {
	shell := ""
	if len(args) == 1 {
		shell = args[0]
	} else {
		shell = filepath.Base(os.Getenv("SHELL"))
	}

	switch shell {
	case "zsh", "bash":
		// supported
	case "":
		return fmt.Errorf("could not detect shell; specify one: ctx shell-init zsh|bash")
	default:
		return fmt.Errorf("unsupported shell %q (supported: zsh, bash)", shell)
	}

	// Name the function after however the user invoked us (ctx or cloudctx),
	// so `command <name>` re-dispatches to the real binary.
	invocation := filepath.Base(os.Args[0])
	if invocation == "" || invocation == "." {
		invocation = "cloudctx"
	}

	// State file written by (*aws.Provider).SetContext on every switch.
	home, _ := os.UserHomeDir()
	stateFile := filepath.Join(home, ".config", "cloudctx", "aws_current")

	fmt.Printf(shellInitTemplate, invocation, invocation, stateFile)
	return nil
}

// shellInitTemplate is valid in both zsh and bash. Placeholders, in order:
// function name, binary name, state file path.
const shellInitTemplate = `# cloudctx shell integration. Add to your rc file:
#   eval "$(%[1]s shell-init)"
export CLOUDCTX_SHELL_INTEGRATION=1
%[1]s() {
  command %[2]s "$@"
  local _cctx_ec=$?
  local _cctx_state="%[3]s"
  if [ -r "$_cctx_state" ]; then
    local _cctx_profile
    _cctx_profile="$(cat "$_cctx_state" 2>/dev/null)"
    if [ -n "$_cctx_profile" ]; then
      export AWS_PROFILE="$_cctx_profile"
    fi
  fi
  return $_cctx_ec
}
`
