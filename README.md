# cloudctx

A unified CLI for switching between cloud contexts. Think **kubectx** for cloud providers.

[![Release](https://img.shields.io/github/v/release/devops-chris/cloudctx)](https://github.com/devops-chris/cloudctx/releases)
[![License](https://img.shields.io/github/license/devops-chris/cloudctx)](LICENSE)

## Features

- Interactive profile/subscription picker with fuzzy filtering
- **AWS**: SSO integration with automatic profile sync, credentials file support
- **Azure**: Subscription switching via Azure CLI
- No shell integration required - works like kubectx (optional integration syncs `AWS_PROFILE` for prompts like Starship)
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

### AWS (default) — brand new to cloudctx

**One org (e.g. work only):**

```bash
ctx aws init            # Prompts: SSO URL, regions → creates config
ctx aws login           # Authenticate in browser
ctx aws sync            # Fetch profiles from SSO

# Daily use
ctx aws                 # Interactive profile picker
ctx aws prod            # Switch to profile matching "prod"
ctx aws -l              # List all profiles
```

**Multiple orgs (e.g. work + personal):** use `org add` instead of (or after) `init`:

```bash
ctx aws org add work    # First org: prompts for SSO URL and regions
ctx aws org add personal # Second org: same prompts, adds to config
ctx aws login --org work
ctx aws sync --org all
ctx aws                 # Picker shows all profiles from all orgs
```

- **init** and **org add** use the same basic flow (SSO URL + regions). `init` is “first-time, one org” and writes the simple config. `org add` adds a *named* org—use it for your first org if you want names from the start, or to add a second, third, etc. after `init`.
- Once you have one org, add more with **`ctx aws org add`** (or **`ctx aws org add <name>`**) as many times as you need.

### Azure

```bash
# First time setup
ctx azure login         # Authenticate (opens browser)

# Daily use
ctx azure               # Interactive subscription picker
ctx azure my-sub        # Switch to subscription matching "my-sub"
ctx azure -l            # List all subscriptions
```

### Shortcuts (default provider)

If you mostly use one cloud, set the **default provider** in config (see [Setting the default provider](#setting-the-default-provider) below). Then you can skip the cloud name:

```bash
ctx                     # Interactive picker (uses default cloud)
ctx prod                # Switch to matching profile/subscription
ctx -l                  # List all
ctx login               # Login (AWS SSO or Azure)
ctx sync                # Sync (AWS) or no-op (Azure)
ctx org rename default work   # When default is AWS: manage orgs
```

Full list of shortcuts and how to set the default: **[Default provider and shortcuts](docs/DEFAULT-PROVIDER-AND-SHORTCUTS.md)**.

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
ctx aws org add [name]    # Add another AWS org (e.g. second SSO portal)
```

**Multi-org:** If you have more than one AWS organization (e.g. work and personal SSO), add each with `ctx aws org add`, then use `--org` with login/sync: `ctx aws login --org work`, `ctx aws sync --org all`. List and picker show all profiles from all orgs. Rename an org (e.g. `default` → `work`) with **`ctx aws org rename default work`**.

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

### Shortcuts (default provider)

When a **default provider** is set, these commands run against that provider (no `aws`/`azure` needed):

| Command | Description |
|---------|-------------|
| `ctx` | Interactive picker |
| `ctx <name>` | Switch to profile/subscription |
| `ctx list` or `-l` | List all |
| `ctx current` or `-c` | Show current |
| `ctx login` | Login (AWS: org of current profile, or `--org NAME`; Azure: browser) |
| `ctx init` | Initialize (AWS SSO; no-op for Azure) |
| `ctx sync` | Sync (AWS profiles; no-op for Azure) |
| `ctx org add/rename/remove/clean-credentials` | Manage AWS orgs (only when default is AWS) |
| `ctx whoami` | Show identity |
| `ctx version` or `-v` | Show version |

See **[Default provider and shortcuts](docs/DEFAULT-PROVIDER-AND-SHORTCUTS.md)** for how to set the default and full details.

> **Note:** `-l`, `-c`, `-v` are shorthand for `list`, `current`, `version`. `ls` is an alias for `list`. Use one or the other, not both.

## Verifying your setup

Run **`cloudctx doctor`** to check that config, default provider, AWS orgs, current profile, and credentials/SSO are correct. Use it anytime you want to confirm things are configured right.

## How It Works

**AWS:** When you select a profile, cloudctx copies its settings to the `[default]` section in `~/.aws/config` (or `~/.aws/credentials` for key-based profiles). No environment variables needed.

**Azure:** Uses `az account set` to switch subscriptions directly via Azure CLI.

## Shell integration (Starship and other prompts)

By default cloudctx does **not** set `AWS_PROFILE` — it switches by rewriting `[default]` in `~/.aws/config`, so no shell setup is needed. But prompt tools like [Starship](https://starship.rs) read the `AWS_PROFILE` environment variable to show the active profile, and a CLI can't change an environment variable in the shell that launched it.

To make your prompt follow the active profile, add the integration to your shell rc file:

```bash
# ~/.zshrc
eval "$(ctx shell-init zsh)"

# ~/.bashrc
eval "$(ctx shell-init bash)"
```

Open a new shell. Now `ctx aws <profile>` (and the interactive picker) update `AWS_PROFILE` automatically, and Starship reflects the change immediately. This works the same way as `zoxide init`, `direnv hook`, etc. — a small wrapper function around `ctx`.

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

### Setting the default provider

The **default provider** is used when you run commands without specifying `aws` or `azure` (e.g. `cloudctx login`, `cloudctx sync`). Set it in your config file:

```yaml
default_cloud: aws   # or "azure" / "az"
```

- **`aws`** — Default. All shortcut commands (login, sync, list, org, etc.) run against AWS.
- **`azure`** or **`az`** — Shortcut commands run against Azure. For AWS-only commands (e.g. org), use `cloudctx aws org ...` explicitly.

You can override via environment: `CLOUDCTX_DEFAULT_CLOUD=azure`. See **[Default provider and shortcuts](docs/DEFAULT-PROVIDER-AND-SHORTCUTS.md)** for the full shortcut list and examples.

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
