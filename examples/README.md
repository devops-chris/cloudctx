# Examples

Configuration and integration examples for cloudctx.

## Files

- `config.yaml` - Example cloudctx configuration (single AWS org)
- `multi-org-config.yaml` - Example with multiple AWS organizations (e.g. work + personal SSO). Use with `cloudctx --config examples/multi-org-config.yaml aws -l`, or copy to `~/.config/cloudctx/config.yaml`. You can also add orgs from the CLI: `cloudctx aws org add`

## Shell Integration (Optional)

While cloudctx works without any shell setup, you can add aliases for convenience:

```bash
# ~/.zshrc or ~/.bashrc

# Short alias
alias ctx='cloudctx'

# Even shorter for the lazy
alias c='cloudctx'
```

## CI/CD Usage

In CI/CD pipelines, you typically don't need cloudctx since you'd use:
- IAM roles (EKS IRSA, EC2 instance profiles)
- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- OIDC federation (GitHub Actions, GitLab CI)

cloudctx is designed for interactive developer workflows.

