# cloudctx

A unified CLI for switching between cloud contexts. Think `kubectx` but for cloud providers.

Supports AWS profiles (via SSO), with Azure and GCP support planned.

## Features

- Interactive profile picker with fuzzy filtering
- AWS SSO integration with automatic profile sync
- No shell integration required - just works like kubectx
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

Download the latest release from [GitHub Releases](https://github.com/devops-chris/cloudctx/releases).

## Quick Start

```bash
# First time: Configure AWS SSO
cloudctx aws init

# Login to AWS SSO
cloudctx aws login

# Sync profiles from SSO
cloudctx aws sync

# Select a profile (interactive)
cloudctx aws

# That's it! AWS CLI now uses the selected profile
aws s3 ls
```

## Usage

### AWS Commands

```bash
# Interactive profile picker (default)
cloudctx aws

# Set specific profile
cloudctx aws my-account:admin

# Set profile matching pattern
cloudctx aws prod    # Shows picker if multiple matches

# Show current profile
cloudctx aws -c
cloudctx aws --current

# List all profiles
cloudctx aws -l
cloudctx aws --list

# Login to AWS SSO
cloudctx aws login

# Sync profiles from SSO
cloudctx aws sync

# Show current identity (like aws sts get-caller-identity)
cloudctx aws whoami
cloudctx aws whoami --json
```

## Configuration

cloudctx uses `~/.config/cloudctx/config.yaml`:

```yaml
# Default cloud provider
default_cloud: aws

# AWS settings
aws:
  sso_start_url: https://your-org.awsapps.com/start
  sso_region: us-east-1
  default_region: us-east-1
```

### Environment Variables

All settings can be overridden with environment variables:

| Variable | Description |
|----------|-------------|
| `CLOUDCTX_DEFAULT_CLOUD` | Default cloud provider (aws, azure) |
| `CLOUDCTX_AWS_SSO_START_URL` | AWS SSO portal URL |
| `CLOUDCTX_AWS_SSO_REGION` | AWS SSO region |
| `CLOUDCTX_AWS_DEFAULT_REGION` | Default AWS region for profiles |

## How It Works

### AWS Profile Sync

When you run `cloudctx aws sync`:

1. Reads your SSO access token from `~/.aws/sso/cache/`
2. Calls AWS SSO APIs to list your accounts and roles
3. Creates/updates profiles in `~/.aws/config`

Profiles are named as `account-name:role-name` (lowercase, spaces replaced with dashes).

### Profile Selection

When you select a profile, cloudctx copies its settings to the `[default]` section in `~/.aws/config`. This is the same approach used by `kubectx` - no environment variables or shell integration needed.

The AWS CLI automatically uses the `[default]` profile when no `AWS_PROFILE` is set.

Note: If you have `AWS_PROFILE` set in your environment, it will override the default. cloudctx will warn you if this is the case.

## Prerequisites

- AWS CLI v2 (for SSO login flow)
- Valid AWS SSO access

## Roadmap

- Azure subscription switching
- GCP project switching
- Profile groups and favorites
- Integration with aws-vault
- Team-based access patterns

## Contributing

Contributions welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

