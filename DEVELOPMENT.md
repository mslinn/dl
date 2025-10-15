# Development Guide

This document provides detailed information for developers working on the `dl` project.

## Makefile

The project uses Make to automate common development tasks. All commands should be run from the project root.

### Available Commands

- `make help` - Show all available Makefile targets with descriptions
- `make build` - Build the main `dl` binary
- `make build-release-tool` - Build the `release` tool (maintainers only)
- `make build-all` - Cross-compile binaries for all supported platforms
- `make test` - Run all unit tests
- `make test-coverage` - Run tests with coverage report
- `make test-race` - Run tests with race condition detector
- `make clean` - Remove built binaries and artifacts
- `make install` - Install the binary to `/usr/local/bin` (requires sudo)
- `make run` - Build and run (shows help by default)
- `make fmt` - Format all Go code using `go fmt`
- `make vet` - Run `go vet` to check for common mistakes
- `make lint` - Run `golint` (requires golint to be installed)
- `make deps` - Download and tidy Go module dependencies
- `make check` - Run all checks (fmt, vet, test)
- `make version` - Display the current version from VERSION file

## Project Structure

```text
dl/
├── cmd/
│   ├── dl/                     # Main application entry point
│   │   └── main.go
│   └── release/                # Release tool (developers only)
│       └── main.go
├── pkg/
│   ├── config/                 # Configuration file parsing and management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── downloader/             # yt-dlp integration and media downloading
│   │   ├── downloader.go
│   │   └── downloader_test.go
│   ├── remote/                 # Remote file copying (SCP/Samba)
│   │   ├── remote.go
│   │   └── remote_test.go
│   └── util/                   # Utility functions (file sanitization, etc.)
│       ├── util.go
│       └── util_test.go
├── scripts/                    # Helper scripts and release documentation
├── tools/                      # Development tool tracking (Go tool dependencies)
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── Makefile                    # Build automation
├── VERSION                     # Version number (e.g., "2.0.0")
├── CHANGELOG.md                # Version history and changes
├── CONTRIBUTING.md             # Contributor guidelines
├── README.md                   # User documentation
└── dl.config                   # Example configuration file
```


## Building

### Build from Source

```bash
# Clone the repository
git clone https://github.com/mslinn/dl.git
cd dl

# Build and install
make build
sudo make install

# Or just build locally
make build
./dl -h
```


### Build the Main Binary

```bash
$ make build # Build for current platform


# The binary will be created as ./dl
./dl -h
```


### Build for Specific Platforms

```bash
# Build for all supported platforms
make build-all

# This creates:
# - dl-linux-amd64
# - dl-linux-arm64
# - dl-darwin-amd64
# - dl-darwin-arm64
# - dl-windows-amd64.exe
```

### Manual Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(cat VERSION)" -o dl-linux ./cmd/dl

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(cat VERSION)" -o dl-mac-arm64 ./cmd/dl

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(cat VERSION)" -o dl.exe ./cmd/dl
```

### Version Injection

The version number is automatically injected from the `VERSION` file at build time using `-ldflags`:

```bash
go build -ldflags "-X main.version=$(cat VERSION)" -o dl ./cmd/dl
```

## Running

### Run Directly with `go run`

```bash
# Run without building a binary
go run ./cmd/dl -h

# Download audio example
go run ./cmd/dl https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Download with verbose output
go run ./cmd/dl -v https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

### Run Built Binary

```bash
# Build first
make build

# Then run
./dl -h
./dl https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

### Run After Install

```bash
# Install to system path
sudo make install

# Now available globally
dl -h
dl https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

## Testing

### Run All Tests

```bash
# Run all tests with verbose output
make test

# Or directly with go
go test ./...
```

### Test Specific Packages

```bash
# Test configuration package
go test ./pkg/config -v

# Test downloader package
go test ./pkg/downloader -v

# Test remote copying
go test ./pkg/remote -v

# Test utilities
go test ./pkg/util -v
```

### Test with Coverage

```bash
# Show coverage for all packages
make test-coverage

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test with Race Detector

```bash
# Check for race conditions (important for concurrent code)
make test-race

# Or for specific package
go test -race ./pkg/remote
```

### Test Filtering

```bash
# Run only tests matching a pattern
go test ./... -run TestDownload

# Run tests with short mode (skip long-running tests)
go test -short ./...
```

## Code Quality

### Format Code

```bash
# Format all Go files
make fmt

# Check formatting without modifying files
gofmt -l .
```

### Static Analysis

```bash
# Run go vet (checks for common mistakes)
make vet

# Run golint (style checker)
make lint

# Run all checks together
make check
```

### Pre-Commit Checklist

Before committing code, ensure:

```bash
make fmt      # Format code
make vet      # Check for issues
make test     # Run all tests
```

Or run them all at once:

```bash
make check
```

## Dependencies

### Managing Dependencies

```bash
# Download all dependencies
go mod download

# Add a new dependency (automatically updates go.mod)
go get github.com/example/package

# Update dependencies
make deps

# Or manually
go mod tidy
```

### Viewing Dependencies

```bash
# List all direct dependencies
go list -m all

# View dependency graph
go mod graph

# Check for available updates
go list -u -m all
```

## Creating Releases

### Version Management

The version is stored in the `VERSION` file at the project root. To update:

```bash
# Edit the VERSION file
echo "2.1.0" > VERSION

# Check current version
make version

# Build with new version
make build
./dl -h  # Should show new version
```

### Using the Release Tool

**Note:** The release tool is for maintainers only and is not installed with `go install`.

```bash
# Build the release tool
make build-release-tool

# Interactive release (prompts for version)
./release

# Release with specific version
./release 2.1.0

# See all options
./release -h
```

The release tool automates:
1. Version validation (semantic versioning)
2. Running tests to ensure everything passes
3. Updating the VERSION file
4. Building binaries for all platforms (Linux, macOS, Windows)
5. Creating git tags
6. Pushing tags to remote repository
7. Preparing artifacts for GitHub releases

### Manual Release Process

If you prefer to create releases manually:

```bash
# 1. Update VERSION file
echo "2.1.0" > VERSION

# 2. Update CHANGELOG.md with changes

# 3. Run tests
make test

# 4. Build for all platforms
make build-all

# 5. Create git tag
git add VERSION CHANGELOG.md
git commit -m "Release v2.1.0"
git tag -a v2.1.0 -m "Release version 2.1.0"

# 6. Push to remote
git push origin master
git push origin v2.1.0
```

### GitHub Release

After pushing the tag, create a GitHub release:

1. Go to the repository's Releases page
2. Click "Draft a new release"
3. Select the tag you just pushed
4. Add release notes from CHANGELOG.md
5. Attach the built binaries from `make build-all`
6. Publish the release

## Development Tips

### Quick Development Cycle

```bash
# Make changes to code
vim pkg/downloader/downloader.go

# Format and test
make fmt && make test

# Test specific package
go test ./pkg/downloader -v

# Build and run
make build && ./dl -v [URL]
```

### Debugging

```bash
# Run with verbose output
./dl -v [URL]

# Use delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug ./cmd/dl -- -v [URL]
```

### Integration Testing

Some tests require external dependencies (yt-dlp, network access):

```bash
# Skip integration tests (if implemented)
go test -short ./...

# Run all tests including integration
go test ./...
```

### Testing Remote Copying

To test SCP/Samba functionality:

```bash
# Create a test config
cp dl.config ~/dl-test.config
# Edit to point to test destinations

# Run with test config
./dl -c ~/dl-test.config -v [URL]
```

## Troubleshooting

### Build Issues

```bash
# Clean and rebuild
make clean
make deps
make build
```

### Test Failures

```bash
# Run tests with verbose output
go test -v ./...

# Run specific failing test
go test ./pkg/config -v -run TestLoadConfig
```

### Module Issues

```bash
# Reset modules
rm -rf go.sum
go mod tidy
go mod download
```

## Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [yt-dlp Documentation](https://github.com/yt-dlp/yt-dlp)
- [Project README](README.md)
- [Contributing Guidelines](CONTRIBUTING.md)
