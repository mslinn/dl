# dl - Media Downloader (Go Version)

Download videos and audio from various websites using yt-dlp, with automatic copying to remote destinations.

This is a complete rewrite of the Python version in Go, featuring a more modular architecture and enhanced functionality.

## Features

- Download audio (MP3) or video (MP4) from supported websites
- Automatic sanitization of filenames
- Support for multiple remote destinations (SCP and Samba/CIFS)
- Configurable local and remote paths via YAML configuration
- Verbose mode for debugging (`-v` or `--verbose`)
- Cross-platform support (Linux, macOS, Windows/WSL)

## Requirements

- Go 1.21 or higher
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) installed and in PATH
- For audio extraction: ffmpeg
- For remote copying:
  - SSH access for SCP method
  - Samba/CIFS support for Windows shares (WSL)

## Installation

### Install yt-dlp

```bash
# Using pip
pip install yt-dlp

# Or download binary from https://github.com/yt-dlp/yt-dlp
```

### Build from source

```bash
# Clone the repository or navigate to the dl directory
cd /path/to/dl

# Build the binary
go build -o dl ./cmd/dl

# Optionally, install to your PATH
sudo mv dl /usr/local/bin/
# Or add to your personal bin directory
mv dl ~/bin/
```

## Configuration

Create a configuration file at `~/dl.config` in YAML format:

```yaml
local:
  mp3s: ${HOME}/Music
  vdest: ${HOME}/Videos/staging
  xdest: ${HOME}/Videos/xrated

remotes:
  my-server:
    disabled: false
    method: scp                    # scp or samba
    mp3s: /data/media/mp3s
    vdest: /data/media/staging
    xdest: /data/secret/videos

  windows-share:
    disabled: false
    method: samba                  # For Windows shares (WSL only)
    mp3s: c:/media/mp3s           # Windows-style paths for samba
    vdest: c:/media/videos
    xdest: c:/secret/videos

  backup-server:
    disabled: true                 # This remote is disabled
    method: scp
    mp3s: /backup/music
    vdest: /backup/videos
```

### Configuration Options

#### Local Paths
- `mp3s`: Directory for downloaded audio files
- `vdest`: Directory for video files (used with `-k` flag)
- `xdest`: Directory for x-rated videos (used with `-x` flag)

#### Remote Configuration
- `disabled`: Set to `true` to temporarily disable a remote
- `method`: Transfer method (`scp` or `samba`)
- `mp3s`: Remote path for MP3 files
- `vdest`: Remote path for videos
- `xdest`: Remote path for x-rated content
- `other`: Alternative remote path (for custom destinations)

### Environment Variables

You can use environment variables in the configuration file:

```yaml
local:
  mp3s: ${HOME}/Music
  vdest: ${MEDIA_DIR}/staging
  xdest: ${STORAGE_DIR}/secret/videos
```

For WSL users, a special `$win_home` variable is automatically set to your Windows home directory.

## Usage

### Basic Usage

```bash
# Download audio (MP3) - default behavior
dl https://www.youtube.com/watch?v=VIDEO_ID

# Download audio with verbose output
dl -v https://www.youtube.com/watch?v=VIDEO_ID

# Download and keep video
dl -k https://www.youtube.com/watch?v=VIDEO_ID

# Download x-rated video to xdest
dl -x https://www.youtube.com/watch?v=VIDEO_ID

# Download video to custom directory
dl -V ~/Downloads https://www.youtube.com/watch?v=VIDEO_ID

# Use alternate config file
dl -c /path/to/config.yaml https://www.youtube.com/watch?v=VIDEO_ID
```

### Command-Line Options

```
Usage: dl [options] URL

Options:
  -c string
        Path to configuration file (default "~/dl.config")
  -d    Enable debug mode (alias for verbose)
  -k    Download and keep video
  -v    Enable verbose output
  -V string
        Download video to specified directory
  -x    Download x-rated video to xdest
```

## Project Structure

The Go version is organized into modular packages:

```
dl/
├── cmd/
│   └── dl/
│       └── main.go              # Main entry point
├── pkg/
│   ├── config/
│   │   ├── config.go            # Configuration loading and management
│   │   └── config_test.go       # Configuration tests
│   ├── downloader/
│   │   ├── downloader.go        # yt-dlp integration
│   │   └── downloader_test.go   # Downloader tests
│   ├── remote/
│   │   ├── remote.go            # Remote file copying (SCP/Samba)
│   │   └── remote_test.go       # Remote tests
│   └── util/
│       ├── util.go              # Utility functions
│       └── util_test.go         # Utility tests
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
└── README-GO.md                 # This file
```

## Testing

Run all tests:

```bash
go test ./...
```

Run tests for a specific package:

```bash
go test ./pkg/config
go test ./pkg/downloader
go test ./pkg/remote
go test ./pkg/util
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Note: Some integration tests are skipped by default as they require yt-dlp, network access, or specific system configuration. To run integration tests manually, remove the `t.Skip()` calls in the test files.

## Architecture

### Modular Design

The Go implementation is significantly more modular than the Python version:

1. **Config Package**: Handles all configuration file parsing and environment variable expansion
2. **Downloader Package**: Manages yt-dlp interaction and media file downloading
3. **Remote Package**: Handles copying files to remote destinations (SCP and Samba)
4. **Util Package**: Provides utility functions for WSL detection, command execution, and file operations

### Key Improvements over Python Version

1. **Better separation of concerns**: Each package has a single, well-defined responsibility
2. **Comprehensive testing**: Unit tests for all packages with good coverage
3. **Type safety**: Go's static typing catches errors at compile time
4. **Better error handling**: Explicit error returns and proper error propagation
5. **No Python dependencies**: Standalone binary with no runtime dependencies
6. **Improved CLI**: More intuitive flag parsing and help messages
7. **Verbose mode**: New `-v`/`--verbose` flag for debugging

## Differences from Python Version

1. **Command-line flags**:
   - Python used `-v` for "keep video", Go uses `-k`
   - New `-v`/`--verbose` flag for debugging output
   - Python's `-d` (debug) is now an alias for `-v` (verbose)

2. **Error handling**: More robust error checking and reporting

3. **Configuration**: Same YAML format, fully compatible

4. **No colorama dependency**: Uses standard output (colors can be added later if needed)

5. **Cleaner code structure**: Better organized with clear package boundaries

## Troubleshooting

### yt-dlp not found

```
Error: yt-dlp not found: please install it
```

Install yt-dlp: `pip install yt-dlp` or download from [yt-dlp releases](https://github.com/yt-dlp/yt-dlp/releases)

### Config file not found

```
Error loading config: failed to read config file
```

Create `~/dl.config` or specify a different path with `-c`

### Permission denied (Samba)

```
failed to mount: command failed
```

Samba mounting requires sudo privileges on WSL. Run with appropriate permissions or use SCP method instead.

### Remote copy failed

```
Warning: failed to copy to remote
```

Check that:
- Remote host is accessible via SSH (for SCP)
- SSH keys are configured for password-less login
- Remote directories exist and are writable
- For Samba: WSL is running and mount point is accessible

## Building for Different Platforms

```bash
# Linux (current platform)
go build -o dl ./cmd/dl

# macOS
GOOS=darwin GOARCH=amd64 go build -o dl-mac ./cmd/dl

# Windows
GOOS=windows GOARCH=amd64 go build -o dl.exe ./cmd/dl

# Linux ARM (Raspberry Pi, etc.)
GOOS=linux GOARCH=arm64 go build -o dl-arm64 ./cmd/dl
```

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `go test ./...`
2. Code is formatted: `go fmt ./...`
3. No linting errors: `golint ./...` (if installed)
4. Add tests for new features

## License

Same license as the original Python version.

## Version Management

The version number is stored in the `VERSION` file at the root of the project. The build process automatically embeds this version into the binary using Go's `//go:embed` directive.

### Manual Version Update

To update the version manually:
1. Edit the `VERSION` file
2. Update `CHANGELOG.md` with your changes
3. Rebuild: `make build`

You can check the current version with:
```bash
make version
# Or after building:
./dl -h
```

### Release Tool (Developers Only)

**Note:** The release tool is for project maintainers only. End users should install dl using `go install` (see Installation section above).

For developers releasing a new version, use the provided Go-based release tool:

```bash
# First, build the release tool
make build-release-tool

# Interactive release (prompts for version)
./release

# Release with specific version
./release 2.1.0

# Skip tests during release
./release -s 2.1.0

# Debug mode
./release -d 2.1.0

# Help
./release -h
```

The release script will:
1. Validate the version format (semantic versioning)
2. Check you're on the correct branch (main/master/golang)
3. Ensure working directory is clean
4. Verify the tag doesn't already exist
5. Check that CHANGELOG.md mentions the new version
6. Run tests (unless `-s` is specified)
7. Update the VERSION file
8. Rebuild the binary with the new version
9. Commit and push the VERSION file change
10. Build binaries for all platforms
11. Create and push a git tag
12. Display next steps for GitHub release

This automates the entire release process and ensures consistency.

## Version History

- **2.0.0** - Initial Go implementation with modular architecture
- Added comprehensive unit tests
- Added `-v`/`--verbose` flag
- Improved error handling and reporting
- Version number stored in VERSION file

## See Also

- [yt-dlp documentation](https://github.com/yt-dlp/yt-dlp)
- [Original Python version](README.md)
