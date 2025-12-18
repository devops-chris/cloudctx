# Contributing to cloudctx

Thank you for your interest in contributing!

## Development Setup

```bash
# Clone
git clone https://github.com/devops-chris/cloudctx.git
cd cloudctx

# Install dependencies
go mod download

# Build
go build -o cloudctx .

# Run tests
go test ./...

# Run linter
golangci-lint run
```

## Making Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `go test ./...`
5. Run linter: `golangci-lint run`
6. Commit with clear messages
7. Push and create a Pull Request

## Adding a New Cloud Provider

1. Create a new package: `internal/<provider>/`
2. Implement the `provider.Provider` interface:
   ```go
   type Provider interface {
       Name() string
       Login() error
       Sync() error
       ListContexts() ([]Context, error)
       SetContext(name string) error
       CurrentContext() (*Context, error)
       WhoAmI() (*Identity, error)
   }
   ```
3. Add commands in `cmd/<provider>.go`
4. Update documentation and examples

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add comments for exported functions
- Keep functions focused and small
- Handle errors explicitly

## Commit Messages

Use clear, descriptive commit messages:
- `feat: Add Azure subscription switching`
- `fix: Handle missing ~/.aws directory`
- `docs: Update installation instructions`
- `refactor: Simplify profile sync logic`

## Questions?

Open an issue for questions or discussions.
