# Secrets Management Guide

How to securely store and use secrets (API keys, database passwords, etc.) in your applications.

---

## Overview

### The Problem
Your application needs secrets like database passwords and API keys, but you can't put them in your code or config files (that would be a security risk). Instead, we store them securely in AWS and automatically inject them into your app when it runs.

### How It Works

**The Flow:**

```
    YOU                        AWS                         KUBERNETES
    ===                        ===                         ==========

 1. lockr CLI  ─────────►  Parameter Store
    (write secrets)         (secure storage)
                                   │
                                   └──────────────────►  2. External Secrets
                                        (fetches)           Operator (ESO)
                                                                  │
                                                                  │ (creates)
                                                                  ▼
                                                            3. Kubernetes Secret
                                                                  │
                                                                  │ (injected)
                                                                  ▼
                                                            4. Your Container
                                                               (reads env vars)
```

**Components Explained:**

| Component | What It Is | Your Interaction |
|-----------|------------|------------------|
| **lockr** | CLI tool for managing secrets in AWS | You use this to add/update/list secrets |
| **AWS Parameter Store** | AWS service that stores encrypted key-value pairs | You don't interact directly - lockr handles it |
| **pt-secrets chart** | Helm chart that defines which secrets your app needs | You configure this in your values.yaml |
| **External Secrets Operator (ESO)** | Service running in Kubernetes that syncs secrets from AWS | Automatic - you never touch this |
| **Kubernetes Secret** | The actual secret object in K8s that your app reads | Created automatically by ESO |
| **pt-service chart** | Helm chart that deploys your app and wires up secrets | You configure this in your values.yaml |

### What You Need To Do

1. **Store your secrets** in AWS using the `lockr` CLI tool
2. **Configure your app** to pull those secrets (add a few lines to your values file)
3. **Deploy** - secrets automatically appear as environment variables in your container

---

## Step 1: Naming Your Secrets

### Path Convention

Follow PracticeTek naming guidelines - broadest scope on left, specific element on right:

```
/{scope}/{env}/{service}/[{component}/]{key}
```

| Segment | Required | Description |
|---------|----------|-------------|
| `scope` | Yes | Vertical, brand, or team that owns the secret (use standard abbreviations) |
| `env` | Yes | Environment: `dev`, `stg`, `prod` |
| `service` | Yes | The application or service name |
| `component` | No | Optional sub-context (e.g., `postgres`, `redis`, `stripe`) |
| `key` | Yes | The specific secret name |

### Scope Guidelines

Use the **standard PracticeTek abbreviations** for scope:

| Scope Type | Examples |
|------------|----------|
| **Vertical** | `ortho`, `chiro`, `vis`, `well`, `mkt`, `shared` |
| **Brand** | `tops`, `rev`, `ctc`, `acom`, `cs`, `pq`, `doc` |
| **Team** | `devops`, `eng`, `infra` |

> **Note:** The `pt-` prefix is implicit - don't need to include it in paths here.

### Standard Abbreviations Reference

**Verticals (broadest scope):**

| Vertical | Abbrev |
|----------|--------|
| Dental & Orthodontics | `ortho` |
| Chiropractic | `chiro` |
| Vision | `vis` |
| Wellness | `well` |
| Marketplace | `mkt` |
| Shared Services | `shared` |

**Brands (within verticals):**

| Vertical | Brand | Abbrev |
|----------|-------|--------|
| **chiro** | ChiroTouch Classic | `ct` |
| | ChiroTouch Cloud | `ctc` |
| | ACOM Health | `acom` |
| | ChiroSpring Desktop | `csd` |
| | ChiroSpring 360 | `cs360` |
| **vis** | RevolutionEHR | `rev` |
| **ortho** | Ora | `ora` |
| | OrthoMinds | `om` |
| | Oasys | `oa` |
| | EasyRX | `erx` |
| | MagicTouch | `mt` |
| | VisualDLP | `vdlp` |
| | TOPS | `tops` |
| **well** | ClinicSource | `cs` |
| | PracticeQ | `pq` |
| | IntakeQ | `iq` |
| **mkt** | Doctible | `doc` |
| | Gaidge | `gdge` |
| | GrowthPlug | `gp` |
| | Inception | `incp` |
| | ZingIT | `zing` |
| | PatientTrak | `ptrk` |
| **shared** | PracticePay | `pay` |

### Examples

**Simple (no component):**
```
/tops/prod/patient-api/jwt-secret
/rev/dev/scheduler-api/database-url
/ctc/stg/billing-svc/api-key
```

**With component context:**
```
/tops/prod/patient-api/postgres/username
/tops/prod/patient-api/postgres/password
/tops/prod/patient-api/redis/auth-token
/rev/prod/analytics-worker/snowflake/username
/rev/prod/analytics-worker/snowflake/password
```

**Shared/infrastructure secrets:**
```
/shared/prod/datadog/api-key
/devops/prod/terraform/backend-key
/infra/prod/certificates/wildcard-cert
```

---

## Step 2: Adding Secrets with lockr

### Install lockr

```bash
brew install devops-chris/tap/lockr
```

### Prerequisites

You need to be logged into AWS before using lockr. If you haven't already:
```bash
cloudctx aws login   # Log in via SSO
cloudctx aws sync    # Sync available profiles
cloudctx             # Select your AWS profile
```

### Write a Secret

```bash
# Simple secret (no component)
lockr write /tops/prod/patient-api/jwt-secret "your-jwt-secret..."

# With component context (database credentials)
lockr write /tops/prod/patient-api/postgres/username "app_user"
lockr write /tops/prod/patient-api/postgres/password "super-secret"

# With description tag
lockr write /tops/prod/patient-api/stripe/secret-key "sk_live_..." \
  --tag "Description=Stripe production API key"

# From a file
lockr write /tops/prod/patient-api/tls-cert --file ./cert.pem

# From stdin (for piping)
echo "secret-value" | lockr write /tops/prod/patient-api/api-key --value -
```

### Verify Your Secret

```bash
# Read a specific secret
lockr read /tops/prod/patient-api/jwt-secret

# List all secrets for your service
lockr list /tops/prod/patient-api

# List secrets for a specific component
lockr list /tops/prod/patient-api/postgres
```

---

## Step 3: Configuring Your App to Use Secrets

Once your secrets are stored in AWS, you need to tell your application where to find them. This is done by adding configuration to your app's **values file** (the YAML file that defines how your app is deployed).

### Option A: Via pt-service (Recommended)

If your app uses the `pt-service` Helm chart (most apps do), add an `externalSecretsEnv` section to your values file:

```yaml
# values.yaml for your app
enable:
  externalSecretsEnv: true           # Turn on secrets feature

externalSecretsEnv:
  name: patient-api-secrets          # A name for this group of secrets
  backend: aws-parameter-store       # Where to fetch from (don't change this)
  refreshInterval: 1h                # How often to check for updated secrets
  sources:
    # Map AWS paths to environment variable names
    - path: /tops/prod/patient-api
      secrets:
        - secret: jwt-secret         # The secret name in AWS
          alias: JWT_SECRET          # The env var name your app will see

    - path: /tops/prod/patient-api/postgres
      secrets:
        - secret: username
          alias: DB_USERNAME         # Your app reads process.env.DB_USERNAME
        - secret: password
          alias: DB_PASSWORD         # Your app reads process.env.DB_PASSWORD
```

**What happens:** When your app starts, `JWT_SECRET`, `DB_USERNAME`, and `DB_PASSWORD` will be available as environment variables - just like if you had set them manually.

### Option B: Standalone (if not using pt-service)

If your app doesn't use pt-service, contact DevOps for help setting up secrets. The configuration is similar but requires a separate deployment step.

### What Happens Behind the Scenes

You don't need to understand this, but if you're curious:

1. Your values file tells Kubernetes "I need these secrets from AWS"
2. A background service automatically fetches them from AWS Parameter Store
3. The secrets are stored securely in Kubernetes
4. When your container starts, they're injected as environment variables

### Reading Secrets in Your Code

Your application just reads environment variables like normal:

```javascript
// Node.js
const dbPassword = process.env.DB_PASSWORD;
const jwtSecret = process.env.JWT_SECRET;
```

```python
# Python
import os
db_password = os.environ['DB_PASSWORD']
jwt_secret = os.environ['JWT_SECRET']
```

```go
// Go
dbPassword := os.Getenv("DB_PASSWORD")
jwtSecret := os.Getenv("JWT_SECRET")
```

### Advanced: Manual Deployment Reference

If you're not using pt-service and need to reference secrets manually in a deployment:

```yaml
# deployment.yaml
spec:
  containers:
    - name: patient-api
      envFrom:
        - secretRef:
            name: patient-api-secrets
      # Or individual env vars:
      env:
        - name: DATABASE_USER
          valueFrom:
            secretKeyRef:
              name: patient-api-secrets
              key: DB_USERNAME
```

---

## Complete Example (using pt-service)

### 1. Store secrets with lockr

```bash
# API keys (simple, no component)
lockr write /tops/prod/patient-api/jwt-secret "your-jwt-signing-key"
lockr write /tops/prod/patient-api/api-key "internal-api-key"

# Database credentials (with component context)
lockr write /tops/prod/patient-api/postgres/host "db.internal.example.com"
lockr write /tops/prod/patient-api/postgres/username "patient_api_user"
lockr write /tops/prod/patient-api/postgres/password "super-secret-password"

# Redis credentials (with component context)
lockr write /tops/prod/patient-api/redis/url "redis://cache.internal:6379"
lockr write /tops/prod/patient-api/redis/password "redis-auth-token"
```

### 2. Configure your pt-service values

```yaml
# values-prod.yaml
enable:
  externalSecretsEnv: true

externalSecretsEnv:
  name: patient-api-secrets
  backend: aws-parameter-store
  refreshInterval: 1h
  sources:
    # API keys
    - path: /tops/prod/patient-api
      secrets:
        - secret: jwt-secret
          alias: JWT_SECRET
        - secret: api-key
          alias: API_KEY

    # Database credentials
    - path: /tops/prod/patient-api/postgres
      secrets:
        - secret: host
          alias: DB_HOST
        - secret: username
          alias: DB_USERNAME
        - secret: password
          alias: DB_PASSWORD

    # Redis credentials
    - path: /tops/prod/patient-api/redis
      secrets:
        - secret: url
          alias: REDIS_URL
        - secret: password
          alias: REDIS_PASSWORD

# Your app config
image:
  repository: your-registry/patient-api
  tag: v1.0.0
```

### 3. Deploy

```bash
helm upgrade --install patient-api pt-service -f values-prod.yaml
```

### 4. Verify it worked

After deployment, check that your secrets are available:

```bash
# See if your app has the env vars (replace patient-api with your app name)
kubectl exec deploy/patient-api -- env | grep DB_
```

You should see output like:
```
DB_HOST=db.internal.example.com
DB_USERNAME=patient_api_user
DB_PASSWORD=super-secret-password
```

> **Note:** When you deploy, secrets are automatically fetched from AWS *before* your application starts. This happens automatically whether you deploy via ArgoCD, Helm, or any other method.

---

## File-Mounted Secrets (Advanced)

Most secrets work fine as environment variables. But sometimes you need a secret as a **file** on disk - for example, TLS certificates or SSH keys that libraries expect to read from a file path.

For these cases, use `externalSecretsFileMounts` instead:

```yaml
enable:
  externalSecretsFileMounts: true

externalSecretsFileMounts:
  name: patient-api-certs
  backend: aws-parameter-store
  sources:
    - path: /tops/prod/patient-api/tls
      secrets:
        - secret: cert
          alias: tls.crt
        - secret: key
          alias: tls.key

# Then mount in your container
volumeMounts:
  - name: certs
    mountPath: /etc/ssl/certs
    
volumes:
  - name: certs
    secret:
      secretName: patient-api-certs
```

---

## Best Practices

### Naming
- Use lowercase with hyphens: `api-key` not `API_KEY` or `apiKey`
- Be descriptive: `stripe-webhook-secret` not `secret1`
- Include context: `database-url` not just `url`

### Security
- Never commit secrets to git
- Use different secrets per environment (dev/staging/prod)
- Rotate secrets regularly
- Use `refreshInterval` to pick up rotated secrets automatically

### Organization
- Use component sub-paths to group related secrets: `/scope/env/service/component/*`
- Keep the same structure across environments (just change `env`)
- Use consistent naming - if `prod` has `/tops/prod/patient-api/postgres/password`, `dev` should have `/tops/dev/patient-api/postgres/password`
- Follow PracticeTek naming conventions - use standard abbreviations for brands/verticals

---

## Troubleshooting

### My app says the environment variable is missing

1. **Check the secret exists in AWS:**
   ```bash
   lockr read /tops/prod/patient-api/your-secret-name
   ```
   If this fails, the secret wasn't created - go back to Step 2.

2. **Check your values file:**
   - Is `enable.externalSecretsEnv: true` set?
   - Does the `path` match exactly where you stored the secret?
   - Does the `alias` match the env var name your app is looking for?

3. **Check the sync status:**
   ```bash
   kubectl get externalsecret <your-secret-name>
   ```
   The `STATUS` column should say `SecretSynced`. If it says `SecretSyncedError`, there's a problem fetching from AWS.

### Permission denied errors

This usually means the Kubernetes cluster doesn't have permission to read from that AWS path. **Contact DevOps** - they'll need to update the cluster's IAM permissions.

### Secret was updated but app still has old value

Secrets refresh automatically based on `refreshInterval` (default: every 2 minutes). You can either:
- Wait for the next refresh cycle
- Restart your pod to pick up changes immediately: `kubectl rollout restart deployment/<your-app>`

### Still stuck?

Reach out to the DevOps team with:
- Your app name and environment
- The secret path you're trying to use
- Any error messages you're seeing

---

## Quick Reference

| Task | Command |
|------|---------|
| Write secret | `lockr write /scope/env/svc/key "value"` |
| Write with component | `lockr write /scope/env/svc/component/key "value"` |
| Read secret | `lockr read /scope/env/svc/key` |
| List service secrets | `lockr list /scope/env/svc` |
| List component secrets | `lockr list /scope/env/svc/component` |
| Delete secret | `lockr delete /scope/env/svc/key` |
| Check sync status | `kubectl get externalsecret` |

**Example:** `lockr write /tops/prod/patient-api/postgres/password "secret123"`

