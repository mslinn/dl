# `dl` [![GitHub newest release](https://img.shields.io/github/v/release/mslinn/dl)](https://github.com/mslinn/dl/releases/latest)

A Command-Line Interface to `yt-dlp` for Downloading Files

Download videos and mp3s from many websites.
Uses Samba and `scp` to copy downloaded media to other computers.


## Features

- Download audio (MP3) or video (MP4) from supported websites
- Automatic installation of `yt-dlp` if not found
- Automatic sanitization of filenames
- **Concurrent copying** to multiple remote destinations for faster performance
- Support for SCP and Samba/CIFS protocols
- Configurable local and remote paths via YAML configuration
- Verbose mode for debugging with full Python stack traces (`-v` or `--verbose`)
- Cross-platform support (Linux, macOS, Windows/WSL)


## Requirements

- Go 1.21 or higher
- [`yt-dlp`](https://github.com/yt-dlp/yt-dlp) - **automatically installed** if not found
- `ffmpeg` - **automatically checked** with installation instructions if not found (required for audio extraction)
- For remote copying:
  - SSH access for SCP method
  - Samba/CIFS support for Windows shares (WSL)


## Installation

### Quick Install

```bash
go install github.com/mslinn/dl/cmd/dl@latest
```


### Upgrading from the Python Version

If you previously had the Python version installed, remove the old wrapper file:

```bash
$ which -a dl # Check for old Python wrapper
$ rm ~/.local/bin/dl
$ hash -r  # clear command cache
$ which dl # Verify the Go binary is active
$ file $(which dl)  # Should show: ELF 64-bit LSB executable
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
    mp3s: c:/media/mp3s            # Windows-style paths
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
  See DEVELOPMENT.md for build and development details.

Examples:
  $ dl https://www.youtube.com/watch?v=dQw4w9WgXcQ
  $ dl -v https://www.youtube.com/watch?v=dQw4w9WgXcQ
  $ dl -k https://www.youtube.com/watch?v=dQw4w9WgXcQ
  $ dl -V ~/Videos https://www.youtube.com/watch?v=dQw4w9WgXcQ
  $ dl -vV . https://www.youtube.com/watch?v=dQw4w9WgXcQ  # Combined flags
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

- Shows the exact `yt-dlp` commands being executed
- Displays full error output including Python tracebacks if `yt-dlp` fails
- Shows download progress and detailed debug information
- Enables `yt-dlp`'s `--verbose` flag for maximum diagnostic output


## Architecture

### Modular Design

The Go implementation uses a modular package structure:

```text
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

1. **Automatic `yt-dlp` Installation**: If `yt-dlp` is not found, the program automatically installs it using pip
2. **Concurrent Remote Copying**: Files are copied to all configured remotes simultaneously using goroutines
3. **Better Separation of Concerns**: Each package has a single, well-defined responsibility
4. **Comprehensive Testing**: Unit tests for all packages with good coverage
5. **Type Safety**: Go's static typing catches errors at compile time
6. **Better Error Handling**: Explicit error returns and proper error propagation
7. **No Python Dependencies**: Standalone binary with no runtime dependencies (except yt-dlp)
8. **Improved CLI**: More intuitive flag parsing and help messages
9. **Verbose Mode**: New `-v`/`--verbose` flag for debugging with full Python stack traces

## Differences from Python Version

1. **Command-line flags**:
   - Python used `-v` for "keep video", Go uses `-k`
   - New `-v`/`--verbose` flag for debugging output
   - Python's `-d` (debug) is now an alias for `-v` (verbose)

2. **Automatic `yt-dlp` installation**: The Go version installs `yt-dlp` if not found

3. **Concurrent copying**: Remote copies happen in parallel for better performance

4. **Error handling**: More robust error checking and reporting with full stack traces in verbose mode

5. **Configuration**: Same YAML format, fully compatible with Python version

6. **Cleaner code structure**: Better organized with clear package boundaries


## Troubleshooting

### `yt-dlp` not found

The program automatically installs `yt-dlp` if not found:

```text
yt-dlp not found. Installing yt-dlp using pip...
[pip installation output]
yt-dlp installed successfully!
```

If automatic installation fails, install manually:

```bash
pip install yt-dlp
# Or download from: https://github.com/yt-dlp/yt-dlp/releases
```

### `ffmpeg` not found

When downloading audio (MP3), the program checks for `ffmpeg` and provides platform-specific installation instructions:

```text
ffmpeg not found. ffmpeg is required for audio extraction.

Installation instructions:
  Ubuntu/Debian: sudo apt-get install ffmpeg
  macOS:         brew install ffmpeg
  Windows:       Download from https://ffmpeg.org/download.html
```

Follow the appropriate command for your platform to install ffmpeg.

### Config file not found

```text
Error loading config: failed to read config file
```

**Solution:** Create `~/dl.config` or specify a different path with `-c`

### Permission denied (Samba)

```text
failed to mount: command failed
```

**Solution:** Samba mounting requires sudo privileges on WSL. Ensure your user has sudo access, or use SCP method instead.

### Remote copy failed

```text
Warning: failed to copy to remote
```

**Check:**

- Remote host is accessible via SSH (for SCP)
- SSH keys are configured for password-less login
- Remote directories exist and are writable
- For Samba: WSL is running and mount point is accessible
- For Samba: Remote name is just the hostname, not `username@hostname`

### Verbose mode shows Python errors

When using `-v`, you may see detailed Python output from `yt-dlp`.
This is intentional and helps diagnose issues.

Example:

```text
[debug] Command-line config: ['--dump-json', '--no-playlist', '--verbose', 'URL']
[debug] Python 3.13.3 (CPython x86_64 64bit)
ERROR: [XHamster] Unknown algorithm ID: 5
```

This detailed output helps identify problems with specific extractors or sites.


## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

For developers, see [DEVELOPMENT.md](DEVELOPMENT.md) for build instructions, testing, and development workflows.

For maintainers, see [RELEASING.md](RELEASING.md) for the release process.


## License

The program is available as open source under the terms of the MIT License.


## Additional Information

- [`yt-dlp` documentation](https://github.com/yt-dlp/yt-dlp)
- [GitHub repository](https://github.com/mslinn/dl)


## Version History

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.
