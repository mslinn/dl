// Release tool for dl project maintainers
//
// This is a development tool for creating new releases of dl.
// It is NOT installed with 'go install' and should only be used by project maintainers.
//
// Build with: make build-release-tool
// Usage: ./release [OPTIONS] [VERSION]
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
)

type Options struct {
	skipTests bool
	debug     bool
}

func main() {
	opts := Options{}
	flag.BoolVar(&opts.skipTests, "s", false, "Skip running tests")
	flag.BoolVar(&opts.debug, "d", false, "Debug mode (additional output)")
	flag.Usage = usage
	flag.Parse()

	fmt.Println("==================================")
	fmt.Println("  dl Release Script")
	fmt.Println("==================================")
	fmt.Println()

	// Show current version
	showCurrentVersion()
	fmt.Println()

	// Get version from argument or prompt
	version := ""
	if flag.NArg() > 0 {
		version = flag.Arg(0)
	} else {
		nextVersion := getNextVersion()
		version = promptVersion(nextVersion)
	}

	// Validate version
	if err := validateVersion(version); err != nil {
		errorExit(err.Error())
	}
	success(fmt.Sprintf("Version format is valid: %s", version))

	// Run checks
	checkBranch()
	checkClean()
	checkTag(version)
	checkChangelog(version)

	// Run tests
	if !opts.skipTests {
		runTests()
	} else {
		warning("Skipping tests.")
	}

	// Update version files
	updateVersionFiles(version)

	// Confirmation
	fmt.Println()
	warning(fmt.Sprintf("Ready to create release v%s", version))
	if !confirmDefault("Proceed with release?", true) {
		errorExit("Release cancelled")
	}

	// Create and push tag
	createTag(version, opts.debug)

	// Run GoReleaser to create GitHub release and upload binaries
	runGoReleaser(version, opts.debug)

	fmt.Println()
	success(fmt.Sprintf("Release v%s completed successfully!", version))
	fmt.Println()

	// Display release URL
	repoURL, err := getRepoURL()
	if err == nil && repoURL != "" {
		info(fmt.Sprintf("View release at: https://github.com/%s/releases/tag/v%s", repoURL, version))
	}
	fmt.Println()
}

func usage() {
	nextVersion := getNextVersion()
	fmt.Fprintf(os.Stderr, "Release a new version of dl\n\n")
	fmt.Fprintf(os.Stderr, "Usage: release [OPTIONS] [VERSION]\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -d    Debug mode (additional output)\n")
	fmt.Fprintf(os.Stderr, "  -h    Display this help message\n")
	fmt.Fprintf(os.Stderr, "  -s    Skip running tests\n\n")
	fmt.Fprintf(os.Stderr, "VERSION: The version to release (e.g., %s)\n", nextVersion)
	os.Exit(0)
}

func info(msg string) {
	fmt.Printf("%sℹ%s  %s\n", colorBlue, colorReset, msg)
}

func success(msg string) {
	fmt.Printf("%s✓%s  %s\n", colorGreen, colorReset, msg)
}

func warning(msg string) {
	fmt.Printf("%s⚠%s  %s\n", colorYellow, colorReset, msg)
}

func errorMsg(msg string) {
	fmt.Printf("%s✗%s  %s\n", colorRed, colorReset, msg)
}

func errorExit(msg string) {
	errorMsg(msg)
	os.Exit(1)
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func runCommandVerbose(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getNextVersion() string {
	// Get version from git tags and increment
	output, err := runCommand("git", "describe", "--tags", "--abbrev=0")
	incrementedVersion := "2.0.0"
	if err == nil {
		latestTag := strings.TrimPrefix(output, "v")
		parts := strings.Split(latestTag, ".")
		if len(parts) == 3 {
			// Increment patch version
			var major, minor, patch int
			fmt.Sscanf(latestTag, "%d.%d.%d", &major, &minor, &patch)
			incrementedVersion = fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
		}
	}

	// Read VERSION file
	versionFileContent, err := os.ReadFile("VERSION")
	if err != nil {
		// VERSION file doesn't exist, use incremented version
		return incrementedVersion
	}

	versionFileVersion := strings.TrimSpace(string(versionFileContent))

	// Validate VERSION file format
	if err := validateVersion(versionFileVersion); err != nil {
		// Invalid format in VERSION file, use incremented version
		return incrementedVersion
	}

	// Compare versions and return the newer one
	if isNewerVersion(versionFileVersion, incrementedVersion) {
		return versionFileVersion
	}

	return incrementedVersion
}

// isNewerVersion returns true if v1 is newer than v2
func isNewerVersion(v1, v2 string) bool {
	var major1, minor1, patch1 int
	var major2, minor2, patch2 int

	fmt.Sscanf(v1, "%d.%d.%d", &major1, &minor1, &patch1)
	fmt.Sscanf(v2, "%d.%d.%d", &major2, &minor2, &patch2)

	if major1 != major2 {
		return major1 > major2
	}
	if minor1 != minor2 {
		return minor1 > minor2
	}
	return patch1 > patch2
}

func validateVersion(version string) error {
	matched, _ := regexp.MatchString(`^[0-9]+\.[0-9]+\.[0-9]+$`, version)
	if !matched {
		return fmt.Errorf("invalid version format: %s (expected: X.Y.Z)", version)
	}
	return nil
}

func checkBranch() {
	branch, err := runCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		errorExit("Failed to get current branch")
	}

	validBranches := []string{"main", "master", "golang"}
	valid := false
	for _, b := range validBranches {
		if branch == b {
			valid = true
			break
		}
	}

	if !valid {
		warning(fmt.Sprintf("You are on branch '%s', not main/master/golang", branch))
		if !confirm("Continue anyway?") {
			errorExit("Aborted")
		}
	}
	success(fmt.Sprintf("On branch: %s", branch))
}

func checkClean() {
	output, _ := runCommand("git", "status", "-s")
	if output != "" {
		warning("Working directory is not clean.")
		fmt.Println(output)
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Commit message (or press Enter for 'Pre-release commit'): ")
		commitMsg, _ := reader.ReadString('\n')
		commitMsg = strings.TrimSpace(commitMsg)
		if commitMsg == "" {
			commitMsg = "Pre-release commit"
		}

		info("Adding all changes...")
		if err := runCommandVerbose("git", "add", "-A"); err != nil {
			errorExit("Failed to add changes")
		}

		info("Committing changes...")
		if err := runCommandVerbose("git", "commit", "-m", commitMsg); err != nil {
			errorExit("Failed to commit changes")
		}

		info("Pushing changes to remote...")
		if err := runCommandVerbose("git", "push", "origin"); err != nil {
			errorExit("Failed to push changes")
		}

		success("Changes committed and pushed")
	} else {
		success("Working directory is clean")
	}
}

func checkTag(version string) {
	tag := fmt.Sprintf("v%s", version)
	_, err := runCommand("git", "rev-parse", tag)
	if err == nil {
		errorExit(fmt.Sprintf("Tag %s already exists", tag))
	}
	success(fmt.Sprintf("Tag %s is available", tag))
}

func checkChangelog(version string) {
	content, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		warning("CHANGELOG.md not found")
		if !confirm("Continue anyway?") {
			errorExit("Please create CHANGELOG.md")
		}
		return
	}

	if !strings.Contains(string(content), version) {
		warning(fmt.Sprintf("CHANGELOG.md does not mention version %s", version))
		if !confirm("Continue anyway?") {
			errorExit("Please update CHANGELOG.md before releasing")
		}
	} else {
		success(fmt.Sprintf("CHANGELOG.md mentions version %s", version))
	}
}

func runTests() {
	info("Running tests...")

	// Try make test first, fall back to go test
	err := runCommandVerbose("make", "test")
	if err != nil {
		// Try go test directly
		err = runCommandVerbose("go", "test", "./...")
	}

	if err != nil {
		errorExit("Tests failed. Fix issues before releasing.")
	}
	success("All tests passed")
}

func updateVersionFiles(version string) {
	info(fmt.Sprintf("Updating VERSION file to %s...", version))

	if err := os.WriteFile("VERSION", []byte(version+"\n"), 0644); err != nil {
		errorExit("Failed to write VERSION file")
	}
	success("VERSION file updated")

	// Rebuild with new version
	info("Rebuilding with new version...")
	err := runCommandVerbose("make", "build")
	if err != nil {
		// Try direct go build
		err = runCommandVerbose("go", "build", "-o", "dl", "./cmd/dl")
	}
	if err != nil {
		errorExit("Build failed")
	}
	success("Binary rebuilt with new version")

	// Commit VERSION file change
	runCommandVerbose("git", "add", "VERSION")
	if err := runCommandVerbose("git", "commit", "-m", fmt.Sprintf("Bump version to %s", version)); err != nil {
		errorExit("Failed to commit VERSION file")
	}
	if err := runCommandVerbose("git", "push", "origin"); err != nil {
		errorExit("Failed to push VERSION file")
	}
	success("VERSION file committed and pushed")
}

func runGoReleaser(version string, debug bool) {
	// Check for GITHUB_TOKEN
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		warning("GITHUB_TOKEN environment variable not set")
		info("Attempting to use GitHub CLI (gh) for authentication...")

		// Try to get token from gh
		ghToken, err := runCommand("gh", "auth", "token")
		if err != nil || ghToken == "" {
			errorExit("Failed to get GitHub token. Please set GITHUB_TOKEN or run 'gh auth login'")
		}
		os.Setenv("GITHUB_TOKEN", ghToken)
		success("Using GitHub CLI token")
	} else {
		success("Found GITHUB_TOKEN environment variable")
	}

	// Check if goreleaser is installed and version
	info("Checking for goreleaser...")
	needsInstall := false
	output, err := runCommand("goreleaser", "--version")
	if err != nil {
		needsInstall = true
	} else {
		// Check if it's v2 or later
		if !strings.Contains(output, "goreleaser version v2") && !strings.Contains(output, "goreleaser version 2") {
			warning("Found older version of goreleaser, upgrading to v2...")
			needsInstall = true
		}
	}

	if needsInstall {
		info("Installing goreleaser v2...")
		if err := runCommandVerbose("go", "install", "github.com/goreleaser/goreleaser/v2@latest"); err != nil {
			errorExit("Failed to install goreleaser v2")
		}
	}
	success("goreleaser v2 is available")

	// Run goreleaser
	fmt.Println()
	info("Running goreleaser to create GitHub release...")

	args := []string{"release", "--clean"}
	if debug {
		args = append(args, "--debug")
	}

	if err := runCommandVerbose("goreleaser", args...); err != nil {
		errorExit("goreleaser failed. The tag has been pushed but the release was not created.")
	}

	success("GitHub release created with binaries uploaded")
}

func getRepoURL() (string, error) {
	repoURL, err := runCommand("git", "config", "--get", "remote.origin.url")
	if err != nil {
		return "", err
	}

	// Extract repo path from git URL
	repoURL = strings.TrimSuffix(repoURL, ".git")
	repoURL = strings.TrimPrefix(repoURL, "git@github.com:")
	repoURL = strings.TrimPrefix(repoURL, "https://github.com/")

	return repoURL, nil
}

func createTag(version string, debug bool) {
	tag := fmt.Sprintf("v%s", version)
	tagMessage := fmt.Sprintf("Release %s", tag)

	if debug {
		tagMessage += "\n\n[debug]"
		warning("Debug mode enabled")
	}

	info(fmt.Sprintf("Creating tag %s...", tag))
	if err := runCommandVerbose("git", "tag", "-a", tag, "-m", tagMessage); err != nil {
		errorExit("Failed to create tag")
	}
	success(fmt.Sprintf("Tag %s created", tag))

	info("Pushing tag to origin...")
	if err := runCommandVerbose("git", "push", "origin", tag); err != nil {
		errorExit("Failed to push tag")
	}
	success("Tag pushed to origin")
}

func showCurrentVersion() {
	latestTag, err := runCommand("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		latestTag = "none"
	}
	info(fmt.Sprintf("Most recent version tag: %s", latestTag))

	versionFile, err := os.ReadFile("VERSION")
	if err != nil {
		info("VERSION file contains: unknown")
	} else {
		info(fmt.Sprintf("VERSION file contains: %s", strings.TrimSpace(string(versionFile))))
	}
}

func promptVersion(defaultVersion string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("What version number should this release have (accept the default with Enter) [%s] ", defaultVersion)
	version, _ := reader.ReadString('\n')
	version = strings.TrimSpace(version)
	if version == "" {
		version = defaultVersion
	}
	return version
}

func confirm(prompt string) bool {
	return confirmDefault(prompt, false)
}

func confirmDefault(prompt string, defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)

	suffix := "(y/N)"
	if defaultYes {
		suffix = "(Y/n)"
	}

	fmt.Printf("%s %s ", prompt, suffix)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	// If empty response, use default
	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}
