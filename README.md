# cloudctx

A unified CLI for switching between cloud contexts. Think **kubectx** for cloud providers.

[![Release](https://img.shields.io/github/v/release/devops-chris/cloudctx)](https://github.com/devops-chris/cloudctx/releases)
[![License](https://img.shields.io/github/license/devops-chris/cloudctx)](LICENSE)

## Features

- Interactive profile picker with fuzzy filtering
- AWS SSO integration with automatic profile sync
- No shell integration required - works like kubectx
- Pretty terminal output with tables and colors
- JSON output for scripting
- Cross-platform (macOS, Linux, Windows)

## Installation

### Homebrew (macOS/Linux)

```bash
brew install devops-chris/tap/cloudctx
```

### Go

```bash
go install github.com/devops-chris/cloudctx@latest
```

### Manual

Download from [GitHub Releases](https://github.com/devops-chris/cloudctx/releases).

## Quick Start

```bash
# First time setup
ctx init                # Configure SSO
ctx login               # Authenticate
ctx sync                # Fetch profiles

# Daily use
ctx                     # Interactive picker
ctx prod                # Switch to profile matching "prod"
```

> **Note:** `ctx` is an alias for `cloudctx`, installed automatically via Homebrew.

## Usage

### Profile Switching

```bash
ctx                       # Interactive profile picker
ctx <profile>             # Set specific profile
ctx prod                  # Fuzzy match (picker if multiple)
ctx -c                    # Show current profile
ctx -l                    # List all profiles
```

### Setup & Auth

```bash
ctx init                  # Configure SSO settings
ctx login                 # SSO authentication
ctx sync                  # Sync profiles from SSO
ctx whoami                # Show current identity
ctx whoami --json
```

## How It Works

When you select a profile, cloudctx copies its settings to the `[default]` section in `~/.aws/config`. This is the same approach used by kubectx - no environment variables or shell integration needed.

## Configuration

Configuration file: `~/.config/cloudctx/config.yaml`

```yaml
default_cloud: aws

aws:
  sso_start_url: https://your-org.awsapps.com/start
  sso_region: us-east-1
  default_region: us-east-1
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLOUDCTX_DEFAULT_CLOUD` | Default cloud provider |
| `CLOUDCTX_AWS_SSO_START_URL` | AWS SSO portal URL |
| `CLOUDCTX_AWS_SSO_REGION` | AWS SSO region |
| `CLOUDCTX_AWS_DEFAULT_REGION` | Default region for profiles |

## Prerequisites

- [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) (required for SSO login)
- AWS SSO access

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features:
- Azure subscription switching
- GCP project switching
- Profile favorites and groups
- aws-vault integration

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE).
