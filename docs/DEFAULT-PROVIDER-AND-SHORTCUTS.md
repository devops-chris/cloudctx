# Default Provider and Shortcuts

When you use one cloud most of the time, you can set a **default provider** and use **shortcuts** so you don't have to type `aws` or `azure` every time.

---

## Setting the default provider

The default provider is used when you run commands **without** specifying a cloud (e.g. `cloudctx login` instead of `cloudctx aws login`).

### Config file (recommended)

Edit your config file (default: `~/.config/cloudctx/config.yaml`):

```yaml
default_cloud: aws   # or "azure"
```

Supported values:

| Value   | Meaning |
|--------|---------|
| `aws`  | AWS is default (also used when omitted or empty) |
| `azure` or `az` | Azure is default |

Example with other settings:

```yaml
default_cloud: aws

aws:
  # ... your AWS orgs, etc.

azure:
  default_location: eastus
```

### Environment variable

You can override the config file with:

```bash
export CLOUDCTX_DEFAULT_CLOUD=aws    # or azure
```

Environment variables take precedence over the config file.

---

## Shortcuts (root-level commands)

These commands run **against the default provider** when you don't specify `aws` or `azure`. If your default is AWS, they behave like `cloudctx aws ...`; if your default is Azure, they behave like `cloudctx azure ...`.

| Shortcut        | With default AWS        | With default Azure       |
|----------------|-------------------------|--------------------------|
| `cloudctx`     | Interactive profile picker | Interactive subscription picker |
| `cloudctx <name>` | Switch to profile       | Switch to subscription   |
| `cloudctx list` or `-l` | List AWS profiles    | List Azure subscriptions  |
| `cloudctx current` or `-c` | Show current profile | Show current subscription |
| `cloudctx login` | AWS SSO login (org of current profile unless you pass `--org`) | Azure login (browser)    |
| `cloudctx init`  | Initialize AWS SSO     | No-op (Azure needs no init) |
| `cloudctx sync`  | Sync AWS profiles from SSO | No-op (Azure fetches live) |
| `cloudctx whoami` | Show AWS identity     | Show Azure identity      |
| `cloudctx org add \| rename \| remove \| clean-credentials` | Manage AWS orgs | **Not available** — use `cloudctx aws org ...` for AWS org commands |

### Notes

- **`-l`** and **`-c`** are shorthand for the `list` and `current` commands. **`-v`** shows version. Use one or the other (e.g. `cloudctx -l` or `cloudctx list`), not both together.
- **Org commands** are AWS-only. When your default is Azure, `cloudctx org rename` (etc.) will error and tell you to use `cloudctx aws org ...` instead.
- You can **always** use the explicit form: `cloudctx aws sync` or `cloudctx azure login` no matter what your default is.

### Login (AWS)

When you run `cloudctx login` without `--org`, cloudctx logs you in to the **org of the profile currently in context**. That way the next `aws s3 ls` (or any AWS call) works. If you pass `--org NAME`, that org is used instead.

---

## Examples

**Default is AWS (config: `default_cloud: aws`):**

```bash
cloudctx              # picker
cloudctx prod         # switch to profile matching "prod"
cloudctx login        # AWS SSO login
cloudctx sync         # sync AWS profiles
cloudctx sync --org all
cloudctx org rename default work
```

**Default is Azure (config: `default_cloud: azure`):**

```bash
cloudctx              # Azure subscription picker
cloudctx my-sub       # switch subscription
cloudctx login        # Azure login
cloudctx whoami       # Azure identity
# For AWS org commands you must specify aws:
cloudctx aws org rename default work
```

---

## Summary

1. Set **`default_cloud`** in `~/.config/cloudctx/config.yaml` (or `CLOUDCTX_DEFAULT_CLOUD`) to `aws` or `azure`.
2. Use the **shortcuts** (`cloudctx`, `cloudctx login`, `cloudctx sync`, etc.) — they run against the default provider.
3. Use **`cloudctx aws ...`** or **`cloudctx azure ...`** when you want to target a provider explicitly, regardless of default.
