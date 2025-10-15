# `dl` - Media Downloader

A command-line tool to download videos and audio from various websites using yt-dlp, with automatic copying to remote destinations via SCP or Samba/CIFS.

## Features

- Download audio (MP3) or video (MP4) from supported websites
- Automatic installation of yt-dlp if not found
- Automatic sanitization of filenames
- **Concurrent copying** to multiple remote destinations for faster performance
- Support for SCP and Samba/CIFS protocols
- Configurable local and remote paths via YAML configuration
- Verbose mode for debugging with full Python stack traces (`-v` or `--verbose`)
- Cross-platform support (Linux, macOS, Windows/WSL)

## Requirements

- Go 1.21 or higher
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - automatically installed if not found
- For audio extraction: ffmpeg
- For remote copying:
  - SSH access for SCP method
  - Samba/CIFS support for Windows shares (WSL)

## Installation

### Quick Install

```bash
go install github.com/mslinn/dl/cmd/dl@latest
```

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

### Upgrading from Python Version

If you previously had the Python version installed, remove old wrappers:

```bash
# Check for old Python wrapper
which -a dl

# Remove if found
rm ~/.local/bin/dl

# Verify the Go binary is active
which dl
file $(which dl)  # Should show: ELF 64-bit LSB executable
```

## Configuration

Create a configuration file at `~/dl.config` in YAML format.

### Configuration Example

```yaml
local:
  mp3s: ${HOME}/Music
  vdest: ${HOME}/Videos/staging
  xdest: ${HOME}/Videos/xrated

remotes:
  my-server:
    disabled: false
    method: scp                    # Transfer via SSH/SCP
    mp3s: /data/media/mp3s
    vdest: /data/media/staging
    xdest: /data/secret/videos

  windows-share:
    disabled: false
    method: samba                  # Windows shares (WSL only)
    mp3s: c:/media/mp3s           # Windows-style paths
    vdest: c:/media/videos
    xdest: c:/secret/videos

  backup-server:
    disabled: true                 # Disabled remote
    method: scp
    mp3s: /backup/music
```

### Configuration Options

#### Local Paths

- **`mp3s`**: Directory for downloaded audio files
- **`vdest`**: Directory for video files (used with `-k` flag)
- **`xdest`**: Directory for x-rated videos (used with `-x` flag)

#### Remote Configuration

- **`disabled`**: Set to `true` to temporarily disable a remote
- **`method`**: Transfer method (`scp` or `samba`)
  - For SCP: Use format `username@hostname` as the remote name
  - For Samba: Use just the hostname (e.g., `myserver`, not `user@myserver`)
- **`mp3s`**: Remote path for MP3 files
- **`vdest`**: Remote path for videos
- **`xdest`**: Remote path for x-rated content
- **`other`**: Alternative remote path (for custom destinations)

#### Environment Variables

You can use environment variables in the configuration file:

```yaml
local:
  mp3s: ${HOME}/Music
  vdest: ${MEDIA_DIR}/staging
  xdest: ${STORAGE_DIR}/secret/videos
```

For WSL users, a special `${win_home}` variable is automatically set to your Windows home directory.

## Usage

```text
dl - Download videos and audio from various websites

Version: 2.0.0

Usage: dl [options] URL

Downloads media from URLs using yt-dlp.
By default, downloads audio as MP3, unless -k, -x, or -V options are provided.

Options:
  -c, --config string    Path to configuration file (default "~/dl.config")
  -d, --debug            Enable debug mode (alias for verbose)
  -k, --keep-video       Download and keep video
  -v, --verbose          Enable verbose output
  -V, --video-dest string
                         Download video to specified directory
  -x, --xrated           Download x-rated video to xdest

Configuration:
  Edit ~/dl.config to configure local and remote destinations.
  See README-GO.md for configuration details.

Examples:
  dl https://www.youtube.com/watch?v=dQw4w9WgXcQ
  dl -v https://www.youtube.com/watch?v=dQw4w9WgXcQ
  dl -k https://www.youtube.com/watch?v=dQw4w9WgXcQ
  dl -V ~/Videos https://www.youtube.com/watch?v=dQw4w9WgXcQ
  dl -vV . https://www.youtube.com/watch?v=dQw4w9WgXcQ  # Combined flags
```

### Usage Examples

```bash
# Download audio (MP3) - default behavior
dl https://www.youtube.com/watch?v=VIDEO_ID

# Download audio with verbose output (shows Python stack traces on errors)
dl -v https://www.youtube.com/watch?v=VIDEO_ID

# Download and keep video
dl -k https://www.youtube.com/watch?v=VIDEO_ID

# Download x-rated video to xdest
dl -x https://www.youtube.com/watch?v=VIDEO_ID

# Download video to custom directory
dl -V ~/Downloads https://www.youtube.com/watch?v=VIDEO_ID

# Download video to current directory with verbose output
dl -vV . https://www.youtube.com/watch?v=VIDEO_ID

# Use alternate config file
dl -c /path/to/config.yaml https://www.youtube.com/watch?v=VIDEO_ID
```

**Note on Verbose Mode (`-v`):**
- Shows the exact yt-dlp commands being executed
- Displays full error output including Python tracebacks if yt-dlp fails
- Shows download progress and detailed debug information
- Enables yt-dlp's `--verbose` flag for maximum diagnostic output

## Architecture

### Modular Design

The Go implementation uses a modular package structure:

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
├── Makefile                     # Build automation
└── VERSION                      # Version file
```

### Key Features

1. **Automatic yt-dlp Installation**: If yt-dlp is not found, the program automatically installs it using pip
2. **Concurrent Remote Copying**: Files are copied to all configured remotes simultaneously using goroutines
3. **Better Separation of Concerns**: Each package has a single, well-defined responsibility
4. **Comprehensive Testing**: Unit tests for all packages with good coverage
5. **Type Safety**: Go's static typing catches errors at compile time
6. **Better Error Handling**: Explicit error returns and proper error propagation
7. **No Python Dependencies**: Standalone binary with no runtime dependencies (except yt-dlp)
8. **Improved CLI**: More intuitive flag parsing and help messages
9. **Verbose Mode**: New `-v`/`--verbose` flag for debugging with full Python stack traces

## Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/config
go test ./pkg/downloader

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...
```

**Note:** Some integration tests are skipped by default as they require yt-dlp, network access, or specific system configuration.

## Differences from Python Version

1. **Command-line flags**:
   - Python used `-v` for "keep video", Go uses `-k`
   - New `-v`/`--verbose` flag for debugging output
   - Python's `-d` (debug) is now an alias for `-v` (verbose)

2. **Automatic yt-dlp installation**: The Go version installs yt-dlp if not found

3. **Concurrent copying**: Remote copies happen in parallel for better performance

4. **Error handling**: More robust error checking and reporting with full stack traces in verbose mode

5. **Configuration**: Same YAML format, fully compatible with Python version

6. **Cleaner code structure**: Better organized with clear package boundaries

## Troubleshooting

### yt-dlp not found

If yt-dlp is not installed, the program will automatically install it:

```
yt-dlp not found. Installing yt-dlp using pip...
[pip installation output]
yt-dlp installed successfully!
```

If automatic installation fails, install manually:
```bash
pip install yt-dlp
# Or download from: https://github.com/yt-dlp/yt-dlp/releases
```

### Config file not found

```
Error loading config: failed to read config file
```

**Solution:** Create `~/dl.config` or specify a different path with `-c`

### Permission denied (Samba)

```
failed to mount: command failed
```

**Solution:** Samba mounting requires sudo privileges on WSL. Ensure your user has sudo access, or use SCP method instead.

### Remote copy failed

```
Warning: failed to copy to remote
```

**Check:**
- Remote host is accessible via SSH (for SCP)
- SSH keys are configured for password-less login
- Remote directories exist and are writable
- For Samba: WSL is running and mount point is accessible
- For Samba: Remote name is just the hostname, not `username@hostname`

### Verbose mode shows Python errors

When using `-v`, you may see detailed Python output from yt-dlp. This is intentional and helps diagnose issues.

Example:
```
[debug] Command-line config: ['--dump-json', '--no-playlist', '--verbose', 'URL']
[debug] Python 3.13.3 (CPython x86_64 64bit)
ERROR: [XHamster] Unknown algorithm ID: 5
```

This detailed output helps identify problems with specific extractors or sites.

## Building for Different Platforms

```bash
# Linux (current platform)
make build

# Build for all platforms
make build-all

# Manual cross-compilation
GOOS=darwin GOARCH=amd64 go build -o dl-mac ./cmd/dl
GOOS=windows GOARCH=amd64 go build -o dl.exe ./cmd/dl
GOOS=linux GOARCH=arm64 go build -o dl-arm64 ./cmd/dl
```

## Development

### Build and Install Locally

```bash
# Format code
make fmt

# Run linter
make vet

# Run tests
make test

# Build
make build

# Install to system path
sudo make install
```

### Version Management

The version number is stored in the `VERSION` file at the root of the project. The build process injects this version into the binary at compile time.

To update the version:

1. Edit the `VERSION` file
2. Update `CHANGELOG.md` with your changes
3. Rebuild: `make build`

Check the current version:
```bash
make version
# Or after building:
./dl -h
```

### Release Process (Maintainers Only)

Use the provided release tool:

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

The release script automates:
- Version validation
- Running tests
- Updating VERSION file
- Building binaries for all platforms
- Creating and pushing git tags
- Preparing GitHub release

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `make test`
2. Code is formatted: `make fmt`
3. No linting errors: `make vet`
4. Add tests for new features
5. Update documentation

## License

The program is available as open source under the terms of the MIT License.

## Additional Information

- [yt-dlp documentation](https://github.com/yt-dlp/yt-dlp)
- [GitHub repository](https://github.com/mslinn/dl)

## Version History

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.

- **2.0.0** - Initial Go implementation with modular architecture
  - Added automatic yt-dlp installation
  - Added concurrent remote copying
  - Added comprehensive unit tests
  - Added `-v`/`--verbose` flag with full stack traces
  - Improved error handling and reporting
  - Version number stored in VERSION file
