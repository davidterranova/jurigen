# Linting with golangci-lint

This project uses [golangci-lint](https://golangci-lint.run/) for code quality analysis.

## Quick Start

```bash
# Install golangci-lint
make lint-install

# Run linting
make lint

# Full development workflow (includes linting)
make dev
```

## Available Commands

- `make lint-install` - Install golangci-lint CLI tool
- `make lint` - Run golangci-lint on the codebase
- `make check-deps` - Check if golangci-lint is installed
- `make dev` - Full workflow: clean, swagger, generate, lint, test

## Configuration

Linting configuration is defined in `.golangci.yml`:

- **Enabled linters**: errcheck, govet, ineffassign, staticcheck, unused
- **Test files**: More lenient rules for `*_test.go` files
- **Testdata**: Excludes generated mock files from strict linting

## Current Issues Found

The linter currently identifies:
- Unchecked error returns (errcheck)
- Suspicious constructs (govet) 
- Unused assignments (ineffassign)
- Code quality issues (staticcheck)
- Unused variables/functions (unused)

## Integration with Development Workflow

Linting is automatically included in:
- `make dev` - Full development build process
- CI/CD pipelines (when configured)

## Fixing Lint Issues

Common fixes:
- Check error returns: `if err := fn(); err != nil { return err }`
- Remove unused variables/imports
- Fix suspicious constructs identified by govet
- Use proper context key types instead of strings
