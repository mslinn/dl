# Contributing to dl

Thank you for your interest in contributing to dl!

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/[username]/dl.git
   cd dl
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the main binary:
   ```bash
   make build
   ```

4. Run tests:
   ```bash
   make test
   ```

## Project Structure

```
dl/
├── cmd/
│   ├── dl/          # Main application (installed by end users)
│   └── release/     # Release tool (developers only, NOT installed)
├── pkg/             # Shared packages
├── tools/           # Development tool tracking
├── scripts/         # Helper scripts and documentation
└── VERSION          # Version number file
```

## Important: Development Tools vs User Tools

### For End Users (cmd/dl)

The main `dl` application in `cmd/dl/` is what end users install via:

```bash
go install github.com/[username]/dl/cmd/dl@latest
```

This is the only command that should be available to end users.

### For Developers Only (cmd/release)

The `release` tool in `cmd/release/` is a development tool for project maintainers. It:

- Is **NOT** installed with `go install`
- Must be built manually with `make build-release-tool`
- Should never be distributed to end users
- Lives in a separate `cmd/` subdirectory

This separation ensures that `go install` only installs the main application, not development tools.

## Making Changes

### Code Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/my-feature
   ```

2. Make your changes

3. Run tests:
   ```bash
   make test
   ```

4. Format code:
   ```bash
   make fmt
   ```

5. Run checks:
   ```bash
   make check
   ```

### Adding Tests

All new features should include unit tests. Tests live alongside the code in `*_test.go` files:

```
pkg/
├── config/
│   ├── config.go
│   └── config_test.go
```

### Documentation

Update relevant documentation:
- `README-GO.md` for user-facing changes
- `scripts/README.md` for release process changes
- Code comments for API changes

## Releasing a New Version (Maintainers Only)

See [scripts/README.md](scripts/README.md) for detailed release instructions.

Quick summary:
1. Update `CHANGELOG.md`
2. Build release tool: `make build-release-tool`
3. Run release: `./release [version]`
4. Create GitHub release with generated binaries

## Code Style

- Follow standard Go conventions
- Use `go fmt` (or `make fmt`)
- Run `go vet` (or `make vet`)
- Add comments for exported functions
- Keep functions focused and testable

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with race detector
make test-race

# Run specific package tests
go test ./pkg/config -v
```

## Pull Request Process

1. Fork the repository
2. Create your feature branch
3. Commit your changes with clear messages
4. Push to your fork
5. Open a Pull Request

### PR Guidelines

- Include tests for new features
- Update documentation as needed
- Keep commits atomic and well-described
- Link related issues
- Ensure CI passes

## Questions?

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Be respectful and constructive in discussions

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
