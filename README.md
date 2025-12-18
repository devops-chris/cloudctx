# cloudctx

A unified CLI for switching between cloud contexts. Think **kubectx** for cloud providers.

[![Release](https://img.shields.io/github/v/release/devops-chris/cloudctx)](https://github.com/devops-chris/cloudctx/releases)
[![License](https://img.shields.io/github/license/devops-chris/cloudctx)](LICENSE)

## Features

- Interactive profile/subscription picker with fuzzy filtering
- **AWS**: SSO integration with automatic profile sync, credentials file support
- **Azure**: Subscription switching via Azure CLI
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

### AWS (default)

```bash
# First time setup
ctx aws init            # Configure SSO
ctx aws login           # Authenticate  
ctx aws sync            # Fetch profiles from SSO

# Daily use
ctx aws                 # Interactive profile picker
ctx aws prod            # Switch to profile matching "prod"
ctx aws -l              # List all profiles
```

### Azure

```bash
# First time setup
ctx azure login         # Authenticate (opens browser)

# Daily use
ctx azure               # Interactive subscription picker
ctx azure my-sub        # Switch to subscription matching "my-sub"
ctx azure -l            # List all subscriptions
```

### Shortcuts

If you mostly use one cloud, set `default_cloud` in config and skip the cloud name:

```bash
ctx                     # Interactive picker (uses default cloud)
ctx prod                # Switch to matching profile/subscription
ctx -l                  # List all
```

> **Note:** `ctx` is an alias for `cloudctx`, installed automatically via Homebrew.

## Usage

### AWS Profile Switching

```bash
ctx                       # Interactive profile picker
ctx <profile>             # Set specific profile
ctx prod                  # Fuzzy match (picker if multiple)
ctx -c                    # Show current profile
ctx -l                    # List all profiles
ctx -l --sso              # List only SSO-synced profiles
ctx -l --manual           # List only manually created profiles
```

The list and picker show both SSO-synced and manually created profiles. Each profile is tagged with its source (`[sso]` or `[manual]`) so you can tell them apart.

### AWS Setup & Auth

```bash
ctx init                  # Configure SSO settings
ctx login                 # SSO authentication
ctx sync                  # Sync profiles from SSO
ctx whoami                # Show current identity
ctx whoami --json
```

### Azure Subscription Switching

```bash
ctx azure                 # Interactive subscription picker
ctx azure <subscription>  # Set specific subscription
ctx azure -c              # Show current subscription
ctx azure -l              # List all subscriptions
ctx azure login           # Azure login (opens browser)
ctx azure whoami          # Show current identity
ctx azure whoami --json
```

> **Note:** Azure doesn't need a `sync` command - subscriptions are fetched live from Azure CLI.

## How It Works

When you select a profile, cloudctx copies its settings to the `[default]` section in `~/.aws/config`. This is the same approach used by kubectx - no environment variables or shell integration needed.

## Configuration

Configuration file: `~/.config/cloudctx/config.yaml`

```yaml
default_cloud: aws  # or "azure"

aws:
  sso_start_url: https://your-org.awsapps.com/start
  sso_region: us-east-1
  default_region: us-east-1

azure:
  default_location: eastus
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLOUDCTX_DEFAULT_CLOUD` | Default cloud provider (`aws` or `azure`) |
| `CLOUDCTX_AWS_SSO_START_URL` | AWS SSO portal URL |
| `CLOUDCTX_AWS_SSO_REGION` | AWS SSO region |
| `CLOUDCTX_AWS_DEFAULT_REGION` | Default region for profiles |
| `CLOUDCTX_AZURE_DEFAULT_LOCATION` | Default Azure location |

## Prerequisites

**For AWS:**
- [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) (required for SSO login)

**For Azure:**
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) (`brew install azure-cli`)

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features:
- GCP project switching
- Profile favorites and groups
- aws-vault integration

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE).
