# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2024-12-18

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

[Unreleased]: https://github.com/devops-chris/cloudctx/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/devops-chris/cloudctx/compare/v0.1.6...v0.2.0
[0.1.6]: https://github.com/devops-chris/cloudctx/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/devops-chris/cloudctx/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/devops-chris/cloudctx/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/devops-chris/cloudctx/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/devops-chris/cloudctx/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/devops-chris/cloudctx/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/devops-chris/cloudctx/releases/tag/v0.1.0

