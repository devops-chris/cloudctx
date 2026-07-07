# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2026-07-07

### Changed
- **UI reskin — pterm replaced with Charm/Lipgloss** — all output now uses the same purple palette, rounded table borders, and huh interactive forms as platformr. The `internal/ui` package provides a shared style layer (`Success`, `Error`, `Warning`, `Info`, `Subtle`, `Highlight`, `CheckPass`, `CheckFail`, `CheckWarn`) so lockr and future tools can stay visually consistent.
- **`doctor` now uses a checklist layout** — each check renders as `✓ / ✗ / ⚠` with indented hints instead of tables, matching platformr's doctor style.
- **`aws -l`, `aws whoami` show the cloudctx banner** — the rounded border header box appears at the top of output-heavy commands.
- **Interactive inputs use huh forms** — `aws init` and `org add` now use the themed huh form components instead of pterm text prompts.
- **`aws sync` uses a huh spinner** — the sync animation matches platformr's spinner style.

## [0.4.3] - 2026-06-10

### Fixed
- **`aws sync --org` now shows valid org names on error** — when you pass an unknown `--org` value, the error message lists the configured org names so you don't have to go dig through the config file.

## [0.4.2] - 2026-06-10

### Fixed
- **Shell integration now syncs with direnv** — `ctx shell-init` installs a `precmd`/`PROMPT_COMMAND` hook that writes `AWS_PROFILE` back to the cloudctx state file whenever an external tool (e.g. direnv) changes it. Previously, switching directories in a direnv project left prompt tools (Starship, Powerline) showing the stale cloudctx profile rather than the directory-scoped one. Re-run `eval "$(ctx shell-init)"` or open a new shell to pick up the new behavior.

## [0.4.1] - 2026-06-09

### Fixed
- **Shell integration now persists across sessions** — `ctx shell-init` seeds `AWS_PROFILE` when a new shell starts (from `~/.config/cloudctx/aws_current`), so your prompt shows the last-switched profile in every new terminal without having to run `ctx` again. Previously `AWS_PROFILE` was only set after the first `ctx` call in a session. Re-open your shell after upgrading to pick up the new behavior.

## [0.4.0] - 2026-06-09

### Added
- **Shell integration for prompts** — `ctx shell-init zsh|bash` prints a shell function that keeps the `AWS_PROFILE` environment variable in sync with the profile you switch to, so prompts like Starship and powerlevel10k display the active profile. It's optional: add `eval "$(ctx shell-init zsh)"` (or `bash`) to your shell rc file. All other functionality works without it, since cloudctx still switches by rewriting `[default]` in `~/.aws/config`. The Homebrew install now prints these setup instructions as caveats. See README "Shell integration".

### Changed
- **Interactive picker** — Replaced the built-in picker with a custom one:
  - Filtering now matches space-separated terms in **any order** (e.g. `admin prod` matches `prod-account:AdminRole`), and matches against org, account ID, role, and source as well as the profile name.
  - **Esc** (or **Ctrl+C**) cancels and leaves your current profile unchanged — previously cancelling could switch to the highlighted row.
  - Applies to both the AWS and Azure pickers.

## [0.3.1] - 2025-02-11

### Added
- **`cloudctx doctor`** — Check that cloudctx and AWS are configured correctly: config file and default_cloud, AWS CLI in PATH, AWS orgs and default org, current profile and org, and whether credentials/SSO are valid (calls WhoAmI). See README "Verifying your setup".

## [0.3.0] - 2025-02-11

### Added
- **Default provider shortcuts** — When `default_cloud` is set, all commands work without typing `aws`/`azure`:
  - `cloudctx login`, `cloudctx sync`, `cloudctx init` route to default cloud (AWS or Azure)
  - Root-level `cloudctx org add|rename|remove|clean-credentials` (when default is AWS); use `cloudctx aws org ...` when default is Azure
- **Login uses profile in context** — `cloudctx login` (no `--org`) logs in to the org of your current profile so `aws s3 ls` etc. work; pass `--org NAME` to log in to a specific org instead
- **Org colors in picker and list** — When you have multiple AWS orgs, each org gets a stable color so you can tell profiles apart at a glance
- **Org rename without guessing** — `cloudctx org rename` with no args: single org prompts for new name; multiple orgs prints "Your orgs: a, b, c" and usage
- **Documentation** — [Default provider and shortcuts](docs/DEFAULT-PROVIDER-AND-SHORTCUTS.md) for setting default and full shortcut list; README section "Setting the default provider"

### Changed
- **[sso] vs [manual] label** — Now based on whether the profile uses SSO (`sso_session` or `sso_account_id` in config), not `cloudctx_managed`. SSO profiles added by hand show [sso]; key-based show [manual]. `cloudctx_managed` is only used for sync/rename
- **sync --org all** — Syncs all orgs and reports which failed at the end instead of stopping on first failure
- **Org column in list/picker** — Shown when you have 2+ orgs or when the single org is not named "default" (so renamed orgs show clearly)
- **clean-credentials help** — Clarified as one-time cleanup; new renames don’t require it

### Fixed
- Profiles that should show as [manual] (e.g. key-based) no longer incorrectly show [sso] when they lack SSO keys

## [0.2.1] - 2024-12-18

### Added
- **Azure support** - Full Azure subscription switching via Azure CLI
  - `ctx azure` - Interactive subscription picker
  - `ctx azure login` - Azure authentication (disables Azure CLI's built-in picker)
  - `ctx azure whoami` - Show current identity
  - `ctx azure list` - List subscriptions
  - Friendly messages for `ctx azure init` and `ctx azure sync` (not needed for Azure)
- Set `default_cloud: azure` in config to use `ctx` directly for Azure
- **Commands and flags work interchangeably:**
  - `list` command or `-l` flag
  - `current` command or `-c` flag  
  - `version` command or `-v` flag
- **AWS: Read profiles from both files** - `~/.aws/config` AND `~/.aws/credentials`
- **AWS: Switch to any profile type** - SSO and credentials-based profiles both work
- Source indicator (`[sso]` or `[manual]`) in AWS profile list and picker
- `--sso` and `--manual` flags to filter AWS profile list

### Fixed
- **AWS: SSO sync now fetches ALL accounts** - Fixed pagination bug that only returned first ~20 accounts

## [0.1.6] - 2024-12-18

### Added
- `ctx` alias for quick access (installed via Homebrew)
- Root-level shortcuts: `cloudctx init`, `cloudctx login`, `cloudctx sync`, `cloudctx whoami`

### Fixed
- Sync now deletes ALL profile sections for clean slate (no more accumulation)
- Fixed duplicate profiles issue
- Ensure SSO session exists when setting context

### Changed
- Profiles no longer use `# cloudctx_managed` marker

## [0.1.5] - 2024-12-18

### Fixed
- Use `sso_session` reference in profiles for proper token caching
- Pin goreleaser to v2

## [0.1.4] - 2024-12-18

### Fixed
- Use `sso_session` reference in profiles for proper token caching
- Fix lint errors for unchecked FlagSet.Set return values

### Added  
- `make check` command to run lint+test+build before pushing

## [0.1.3] - 2024-12-18

### Added
- ROADMAP.md with future plans
- Improved examples with README

### Changed
- Polished README with badges and cleaner structure
- Enhanced CONTRIBUTING.md with provider implementation guide

## [0.1.2] - 2024-12-18

### Added
- Root command now uses default cloud - just type `cloudctx` instead of `cloudctx aws`
- Added `-c` and `-l` flags to root command for quick access

### Changed
- Simplified usage: `cloudctx`, `cloudctx -l`, `cloudctx <profile>`

## [0.1.1] - 2024-12-18

### Fixed
- Create `~/.aws` directory if it doesn't exist on fresh installs
- Add AWS CLI check with helpful error message if not installed
- Fixed lint errors for unchecked return values
- Include `go.sum` in repository for CI builds

## [0.1.0] - 2024-12-18

### Added
- Initial release
- Interactive AWS profile picker with fuzzy filtering
- AWS SSO integration with automatic profile sync
- Profile switching by updating `~/.aws/config` default section (like kubectx)
- `cloudctx aws init` - Configure SSO settings
- `cloudctx aws login` - SSO authentication
- `cloudctx aws sync` - Sync profiles from SSO
- `cloudctx aws whoami` - Show current AWS identity
- Pretty terminal output with tables and colors
- JSON output for scripting
- Cross-platform support (macOS, Linux, Windows)
- Homebrew installation support

[Unreleased]: https://github.com/devops-chris/cloudctx/compare/v0.4.1...HEAD
[0.4.1]: https://github.com/devops-chris/cloudctx/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/devops-chris/cloudctx/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/devops-chris/cloudctx/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/devops-chris/cloudctx/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/devops-chris/cloudctx/compare/v0.1.6...v0.2.1
[0.1.6]: https://github.com/devops-chris/cloudctx/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/devops-chris/cloudctx/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/devops-chris/cloudctx/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/devops-chris/cloudctx/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/devops-chris/cloudctx/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/devops-chris/cloudctx/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/devops-chris/cloudctx/releases/tag/v0.1.0

