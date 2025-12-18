# Contributing to cloudctx

Thank you for your interest in contributing to cloudctx!

## Development Setup

1. Clone the repository:
```bash
git clone https://github.com/devops-chris/cloudctx.git
cd cloudctx
```

2. Install dependencies:
```bash
go mod download
```

3. Build:
```bash
go build -o cloudctx .
```

4. Run tests:
```bash
go test ./...
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

1. Create a new package in `internal/` (e.g., `internal/azure/`)
2. Implement the `provider.Provider` interface
3. Add commands in `cmd/` (e.g., `cmd/azure.go`)
4. Update documentation

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add comments for exported functions
- Keep functions focused and small

## Questions?

Open an issue for any questions or discussions.

