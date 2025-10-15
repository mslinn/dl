# Releasing

This document describes the release process for the `dl` project. This is intended for project maintainers only.

## Prerequisites

Before creating a release, ensure:

1. **Write access** to the repository
2. **All tests pass**: Run `make test` and verify all tests succeed
3. **Master branch is up to date**: `git pull origin master`
4. **CHANGELOG.md is updated** with all changes for this release
5. **VERSION file** reflects the correct version number
6. **No uncommitted changes**: `git status` should show a clean working tree

## Creating a Release

### Method 1: Using the Release Tool (Recommended)

The project includes a dedicated release tool that automates most of the process.

#### Build the Release Tool

```bash
make build-release-tool
```

This creates a `./release` executable in the project root.

#### Run the Release Tool

**Interactive Mode (prompts for version):**
```bash
./release
```

**Specify Version:**
```bash
./release 2.1.0
```

**See All Options:**
```bash
./release -h
```

#### What the Release Tool Does

The release tool automates:

1. **Version Validation**: Ensures the version follows semantic versioning (e.g., 2.1.0)
2. **Test Execution**: Runs `make test` to ensure all tests pass
3. **VERSION File Update**: Updates the VERSION file with the new version number
4. **Cross-Platform Builds**: Builds binaries for all supported platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
5. **Git Tagging**: Creates an annotated git tag (e.g., `v2.1.0`)
6. **Push to Remote**: Pushes the tag and commits to the remote repository
7. **Release Preparation**: Prepares artifacts for GitHub release

### Method 2: Manual Release Process

If you prefer manual control or need to troubleshoot, follow these steps:

#### 1. Update Version and Changelog

```bash
# Update the VERSION file
echo "2.1.0" > VERSION

# Edit CHANGELOG.md and add your changes
vim CHANGELOG.md

# Commit the changes
git add VERSION CHANGELOG.md
git commit -m "Prepare release v2.1.0"
```

#### 2. Run Tests

```bash
make test
```

Ensure all tests pass before proceeding.

#### 3. Build for All Platforms

```bash
make build-all
```

This creates binaries for:
- `dl-linux-amd64`
- `dl-linux-arm64`
- `dl-darwin-amd64`
- `dl-darwin-arm64`
- `dl-windows-amd64.exe`

#### 4. Create and Push Git Tag

```bash
# Create annotated tag
git tag -a v2.1.0 -m "Release version 2.1.0"

# Push commits and tag to remote
git push origin master
git push origin v2.1.0
```

#### 5. Create GitHub Release

1. Go to the [Releases page](https://github.com/mslinn/dl/releases)
2. Click **"Draft a new release"**
3. Select the tag you just pushed (`v2.1.0`)
4. Set the release title (e.g., "v2.1.0")
5. Copy the relevant section from CHANGELOG.md to the release description
6. Attach the built binaries:
   - `dl-linux-amd64`
   - `dl-linux-arm64`
   - `dl-darwin-amd64`
   - `dl-darwin-arm64`
   - `dl-windows-amd64.exe`
7. Click **"Publish release"**

## Supported Platforms

Releases include binaries for the following platforms:

| Platform | Architecture | Binary Name |
|----------|-------------|-------------|
| Linux | amd64 | dl-linux-amd64 |
| Linux | arm64 | dl-linux-arm64 |
| macOS | amd64 (Intel) | dl-darwin-amd64 |
| macOS | arm64 (Apple Silicon) | dl-darwin-arm64 |
| Windows | amd64 | dl-windows-amd64.exe |

## Version Numbering

The project follows [Semantic Versioning](https://semver.org/):

**MAJOR.MINOR.PATCH**

- **MAJOR**: Incompatible API changes or major rewrites
- **MINOR**: New features, backwards-compatible
- **PATCH**: Bug fixes, backwards-compatible

### Examples

- `2.0.0` - Major version (initial Go rewrite)
- `2.1.0` - Minor version (new feature added)
- `2.1.1` - Patch version (bug fix)

## Changelog

Before releasing, update `CHANGELOG.md` with:

- **New Features**: List new functionality
- **Bug Fixes**: Document fixed issues
- **Breaking Changes**: Highlight incompatible changes
- **Deprecations**: Note deprecated features
- **Performance**: Mention performance improvements
- **Documentation**: Note documentation updates

### Changelog Example

```markdown
## [2.1.0] - 2024-10-15

### Added
- Support for batch downloading multiple URLs
- New `--playlist` flag to download entire playlists

### Fixed
- Fixed race condition in concurrent remote copying
- Corrected filename sanitization for special characters

### Changed
- Improved error messages for yt-dlp failures
- Updated verbose mode to show more diagnostic information
```

## Verify the Release

After publishing the release:

1. **Check the Releases page**: Verify the release appears at https://github.com/mslinn/dl/releases
2. **Verify binaries**: Confirm all platform binaries are attached and downloadable
3. **Test installation**: Test `go install` with the new version:
   ```bash
   go install github.com/mslinn/dl/cmd/dl@v2.1.0
   ```
4. **Test binary**: Download and run a binary to verify it works correctly
5. **Check version**: Ensure the binary reports the correct version:
   ```bash
   dl -h  # Should show "Version: 2.1.0"
   ```

## Rollback

If you discover a critical issue after releasing:

### Delete the Git Tag

```bash
# Delete local tag
git tag -d v2.1.0

# Delete remote tag
git push origin :refs/tags/v2.1.0
```

### Delete the GitHub Release

1. Go to the [Releases page](https://github.com/mslinn/dl/releases)
2. Click on the problematic release
3. Click **"Delete"**
4. Confirm deletion

### Create a Patch Release

After fixing the issue:

```bash
# Increment patch version
echo "2.1.1" > VERSION

# Update CHANGELOG.md with the fix
vim CHANGELOG.md

# Follow the normal release process
./release 2.1.1
```

## Troubleshooting

### Tests Fail During Release

**Solution**: Fix the failing tests before releasing. Never release with failing tests.

```bash
# Run tests with verbose output
make test

# Or run specific package tests
go test ./pkg/config -v
```

### Build Fails for Specific Platform

**Solution**: Check the Go version and cross-compilation setup.

```bash
# Verify Go version (requires 1.21+)
go version

# Test build for specific platform
GOOS=darwin GOARCH=arm64 go build -o test-binary ./cmd/dl
```

### Tag Already Exists

**Solution**: Delete the existing tag or use a different version number.

```bash
# Delete local and remote tag
git tag -d v2.1.0
git push origin :refs/tags/v2.1.0

# Then create the new tag
git tag -a v2.1.0 -m "Release version 2.1.0"
git push origin v2.1.0
```

### Version Mismatch in Binary

**Solution**: Ensure the VERSION file is correct and rebuild.

The version is injected at build time from the VERSION file:

```bash
# Check VERSION file
cat VERSION

# Rebuild with correct version
make clean
make build

# Verify version in binary
./dl -h
```

### GitHub Release Upload Fails

**Solution**: Manually upload binaries through the web interface.

1. Create the release without binaries
2. Edit the release
3. Drag and drop the binary files to upload them

## Post-Release Checklist

After successfully releasing:

- [ ] Verify release appears on GitHub
- [ ] Test `go install` with new version
- [ ] Announce release (if applicable)
- [ ] Update any external documentation
- [ ] Close related GitHub issues/PRs
- [ ] Consider posting to relevant communities

## Additional Resources

- [Semantic Versioning](https://semver.org/)
- [GitHub Releases Documentation](https://docs.github.com/en/repositories/releasing-projects-on-github)
- [Go Release Best Practices](https://go.dev/doc/modules/release-workflow)
- [DEVELOPMENT.md](DEVELOPMENT.md) - Build and development instructions
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
