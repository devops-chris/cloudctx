# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/devops-chris/cloudctx/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/devops-chris/cloudctx/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/devops-chris/cloudctx/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/devops-chris/cloudctx/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/devops-chris/cloudctx/releases/tag/v0.1.0

