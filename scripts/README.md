# Release Tool (Developers Only)

A Go-based release tool that automates the entire release process for dl.

**IMPORTANT:** This is a development tool for project maintainers only. It is NOT intended for end users and is NOT installed with `go install`. End users should simply run:

```bash
go install github.com/[username]/dl/cmd/dl@latest
```

## Building (Developers Only)

```bash
make build-release-tool
```

This creates a `release` binary in the project root.

## Usage

```bash
./release [OPTIONS] [VERSION]
```

### Options

- `-d` - Debug mode (additional output)
- `-h` - Display help message
- `-s` - Skip running tests

### Examples

```bash
# Interactive release (will prompt for version)
./release

# Release version 2.1.0
./release 2.1.0

# Release with skipped tests
./release -s 2.1.0

# Release in debug mode
./release -d 2.1.0
```

### What the Script Does

1. **Version Validation**: Ensures version follows semantic versioning (X.Y.Z)
2. **Branch Check**: Verifies you're on main, master, or golang branch
3. **Clean Working Directory**: Checks for uncommitted changes
4. **Tag Availability**: Ensures the version tag doesn't already exist
5. **Changelog Check**: Verifies CHANGELOG.md mentions the new version
6. **Run Tests**: Executes test suite (unless `-s` is specified)
7. **Update VERSION File**: Updates VERSION file and commits it
8. **Rebuild Binary**: Rebuilds with new version embedded
9. **Build All Platforms**: Creates binaries for Linux, macOS, Windows (x64 and ARM64)
10. **Create Git Tag**: Creates annotated tag (v{VERSION})
11. **Push to Remote**: Pushes commits and tag to origin
12. **Display Next Steps**: Shows instructions for GitHub release

### Version Auto-Increment

If no version is provided, the script will:
- Look for the latest git tag
- Increment the patch version (e.g., 2.0.0 → 2.0.1)
- Suggest this as the default

### Requirements

- Git repository with remote origin
- Go toolchain installed
- Write access to the repository
- Clean working directory (or willingness to commit changes)

### Pre-Release Checklist

Before running the release script:

1. ✅ Update CHANGELOG.md with new version and changes
2. ✅ Ensure all changes are committed or ready to commit
3. ✅ Tests pass: `make test`
4. ✅ Code builds: `make build`
5. ✅ Review changes: `git status` and `git log`

### Post-Release Steps

After the script completes:

1. Go to GitHub releases page
2. Click "Draft a new release"
3. Select the newly created tag
4. Copy relevant CHANGELOG.md entries to the release notes
5. Upload the built binaries (dl-linux-amd64, dl-darwin-amd64, etc.)
6. Publish the release
7. Announce on relevant channels

### Troubleshooting

**Tag already exists**
```bash
# Delete tag locally and remotely
git tag -d v2.1.0
git push origin :refs/tags/v2.1.0
```

**Tests failing**
```bash
# Fix tests first, or skip with -s flag (not recommended)
./scripts/release -s 2.1.0
```

**Wrong branch**
```bash
# Switch to correct branch
git checkout main
```

**Uncommitted changes**
- Script will prompt to commit them
- Or commit/stash manually first

### Script Output

The script provides color-coded output:
- 🔵 Blue (ℹ): Informational messages
- 🟢 Green (✓): Success messages
- 🟡 Yellow (⚠): Warnings
- 🔴 Red (✗): Errors

### Safety Features

- Validates version format
- Checks for clean working directory
- Verifies tag doesn't exist
- Confirms before creating release
- Runs tests before tagging
- Checks CHANGELOG.md is updated

### Customization

To customize the release process, edit `cmd/release/main.go`:
- Modify `checkBranch()` to add/remove allowed branches
- Update `runTests()` to change test command
- Adjust `buildReleases()` for different build targets
- Change `createTag()` for different tag message format

After making changes, rebuild:
```bash
make build-release-tool
```
