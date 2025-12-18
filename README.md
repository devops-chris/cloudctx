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

### AWS

```bash
ctx aws                   # Interactive profile picker
ctx aws <profile>         # Switch to profile
ctx aws list              # List profiles (or: ctx aws -l)
ctx aws current           # Show current (or: ctx aws -c)
ctx aws login             # SSO login
ctx aws sync              # Sync from SSO
ctx aws whoami            # Show identity
ctx aws init              # Configure SSO (first time)
```

Filter options:
```bash
ctx aws list --sso        # Only SSO-synced profiles
ctx aws list --manual     # Only manually created profiles
```

### Azure

```bash
ctx azure                 # Interactive subscription picker
ctx azure <subscription>  # Switch to subscription
ctx azure list            # List subscriptions (or: ctx azure -l)
ctx azure current         # Show current (or: ctx azure -c)
ctx azure login           # Azure login (opens browser)
ctx azure whoami          # Show identity
```

> **Note:** Azure doesn't need `init` or `sync` - subscriptions are fetched live.

### Shortcuts

Routes to `default_cloud` (default: aws):

```bash
ctx                       # Interactive picker
ctx <name>                # Switch to profile/subscription
ctx list                  # List all (or: ctx -l)
ctx current               # Show current (or: ctx -c)
ctx version               # Show version (or: ctx -v)
ctx login                 # Login
ctx whoami                # Show identity
```

> **Note:** `-l`, `-c`, `-v` are shortcuts for `list`, `current`, `version` commands.
> `ls` is an alias for `list`. Use one or the other, not both.

## How It Works

**AWS:** When you select a profile, cloudctx copies its settings to the `[default]` section in `~/.aws/config` (or `~/.aws/credentials` for key-based profiles). No environment variables needed.

**Azure:** Uses `az account set` to switch subscriptions directly via Azure CLI.

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
