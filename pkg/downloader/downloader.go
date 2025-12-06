package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// MediaType represents the type of media to download
type MediaType int

const (
	// MP3 represents MP3 audio format
	MP3 MediaType = iota
	// Video represents video format
	Video
)

// Options contains download configuration
type Options struct {
	URL         string
	Destination string
	MediaType   MediaType
	Format      string // mp3, mp4, etc.
	Verbose     bool
	XRated      bool
}

// Downloader handles downloading media using yt-dlp
type Downloader struct {
	opts *Options
}

// New creates a new Downloader
func New(opts *Options) *Downloader {
	return &Downloader{opts: opts}
}

// GetMediaName extracts and sanitizes the media title from the URL
func (d *Downloader) GetMediaName() (string, error) {
	// Run yt-dlp to get video info in JSON format
	args := []string{"--dump-json", "--no-playlist"}
	if d.opts.Verbose {
		args = append(args, "--verbose")
	} else {
		args = append(args, "--no-warnings")
	}
	args = append(args, d.opts.URL)

	cmd := exec.Command("yt-dlp", args...)

	if d.opts.Verbose {
		fmt.Printf("Running: yt-dlp %s\n", strings.Join(args, " "))
		cmd.Stderr = os.Stderr // Show stderr in verbose mode
	}

	output, err := cmd.Output()
	if err != nil {
		// If not verbose, capture and show the error output
		if !d.opts.Verbose {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintf(os.Stderr, "yt-dlp error: %s\n", string(exitErr.Stderr))
			}
		}
		return "", fmt.Errorf("failed to extract media info: %w", err)
	}

	// Parse JSON output
	var info map[string]interface{}
	if err := json.Unmarshal(output, &info); err != nil {
		return "", fmt.Errorf("failed to parse media info: %w", err)
	}

	// Get title
	title, ok := info["title"].(string)
	if !ok {
		return "no_name", nil
	}

	// Sanitize title: remove non-alphanumeric characters except spaces
	reg := regexp.MustCompile(`[^A-Za-z0-9 ]+`)
	name := reg.ReplaceAllString(title, "")
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")

	// Replace multiple underscores with single underscore
	multiUnderscore := regexp.MustCompile(`__+`)
	name = multiUnderscore.ReplaceAllString(name, "_")

	// Limit length to 200 characters
	if len(name) > 200 {
		name = name[:200]
	}

	return name, nil
}

// Download downloads the media according to the options
func (d *Downloader) Download() (string, error) {
	mediaName, err := d.GetMediaName()
	if err != nil {
		return "", err
	}

	if err := d.checkPrerequisites(); err != nil {
		return "", err
	}

	outputPath, args, err := d.buildYtDlpArgs(mediaName)
	if err != nil {
		return "", err
	}

	if err := d.executeYtDlp(args); err != nil {
		return "", err
	}

	actualPath := d.determineOutputPath(outputPath)
	return d.cleanupAndReturn(actualPath)
}

func (d *Downloader) checkPrerequisites() error {
	// Check for ffmpeg if downloading audio
	if d.opts.MediaType == MP3 {
		return CheckFfmpeg()
	}
	return nil
}

func (d *Downloader) buildYtDlpArgs(mediaName string) (outputPath string, args []string, err error) {
	switch d.opts.MediaType {
	case MP3:
		return d.buildMP3Args(mediaName)
	case Video:
		return d.buildVideoArgs(mediaName)
	default:
		return "", nil, fmt.Errorf("unsupported media type: %v", d.opts.MediaType)
	}
}

func (d *Downloader) buildMP3Args(mediaName string) (outputPath string, args []string, err error) {
	outputPath = filepath.Join(d.opts.Destination, mediaName)
	args = []string{
		"--extract-audio",
		"--audio-format", d.opts.Format,
		"--output", outputPath + ".%(ext)s",
	}

	args = appendVerboseArgs(args, d.opts.Verbose)
	args = append(args, d.opts.URL)

	fmt.Printf("Saving %s.%s\n", outputPath, d.opts.Format)

	return outputPath, args, nil
}

func (d *Downloader) buildVideoArgs(mediaName string) (outputPath string, args []string, err error) {
	outputPath = filepath.Join(d.opts.Destination, mediaName)
	args = []string{
		"--format", "mp4",
		"--merge-output-format", "mp4",
		"--output", outputPath + ".mp4",
	}

	args = appendVerboseArgs(args, d.opts.Verbose)
	args = append(args, d.opts.URL)

	fmt.Printf("Saving %s.mp4\n", outputPath)

	return outputPath, args, nil
}

func appendVerboseArgs(args []string, verbose bool) []string {
	if verbose {
		return append(args, "--verbose", "--progress")
	}
	return append(args, "--quiet", "--no-warnings")
}

func (d *Downloader) executeYtDlp(args []string) error {
	if d.opts.Verbose {
		fmt.Printf("Running: yt-dlp %s\n", strings.Join(args, " "))
	}

	cmd := exec.Command("yt-dlp", args...)

	// Capture output for better error diagnostics
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		// Show output in non-verbose mode if there were any messages
		if !d.opts.Verbose && stdout.String() != "" {
			fmt.Print(stdout.String())
		}
		return nil
	}

	return d.formatYtDlpError(err, args, stdout.String(), stderr.String())
}

func (d *Downloader) formatYtDlpError(err error, args []string, stdoutStr, stderrStr string) error {
	var errorMsg strings.Builder
	errorMsg.WriteString(fmt.Sprintf("yt-dlp execution failed: %v", err))
	errorMsg.WriteString(fmt.Sprintf("\nCommand: yt-dlp %s", strings.Join(args, " ")))

	if stdoutStr != "" {
		errorMsg.WriteString(fmt.Sprintf("\nstdout: %s", strings.TrimSpace(stdoutStr)))
	}
	if stderrStr != "" {
		errorMsg.WriteString(fmt.Sprintf("\nstderr: %s", strings.TrimSpace(stderrStr)))
	}

	// Provide common troubleshooting hints
	hint := d.getYtDlpErrorHint(stderrStr)
	if hint != "" {
		errorMsg.WriteString(fmt.Sprintf("\nHint: %s", hint))
	}

	return fmt.Errorf("%s", errorMsg.String())
}

func (d *Downloader) getYtDlpErrorHint(stderrStr string) string {
	switch {
	case strings.Contains(stderrStr, "network") || strings.Contains(stderrStr, "timeout"):
		return "Check your internet connection or try again later"
	case strings.Contains(stderrStr, "permission") || strings.Contains(stderrStr, "Permission denied"):
		return "Check file permissions in the destination directory"
	case strings.Contains(stderrStr, "space") || strings.Contains(stderrStr, "disk"):
		return "Check available disk space"
	case strings.Contains(stderrStr, "format") || strings.Contains(stderrStr, "unsupported"):
		return "The video format might not be supported or available"
	default:
		return ""
	}
}

func (d *Downloader) determineOutputPath(outputPath string) string {
	actualPath := outputPath + "." + d.opts.Format
	if d.opts.MediaType == Video {
		actualPath = outputPath + ".mp4"
	}
	return actualPath
}

func (d *Downloader) cleanupAndReturn(actualPath string) (string, error) {
	// Clean up any .webm files if video
	if d.opts.MediaType == Video {
		webmPath := strings.TrimSuffix(actualPath, ".mp4") + ".webm"
		if _, err := os.Stat(webmPath); err == nil {
			if err := os.Remove(webmPath); err != nil {
				return actualPath, fmt.Errorf("failed to remove .webm file: %w", err)
			}
		}
	}

	return actualPath, nil
}

// CheckFfmpeg verifies that ffmpeg is installed and accessible
// If not found, provides installation instructions
func CheckFfmpeg() error {
	// Check if ffmpeg is already installed
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err == nil {
		return nil // ffmpeg is already installed
	}

	// ffmpeg not found, provide installation instructions
	color.Red("ffmpeg not found. ffmpeg is required for audio extraction.")
	fmt.Println()
	color.Yellow("Installation instructions:")
	fmt.Println("  Ubuntu/Debian: sudo apt-get install ffmpeg")
	fmt.Println("  macOS:         brew install ffmpeg")
	fmt.Println("  Windows:       Download from https://ffmpeg.org/download.html")
	fmt.Println()

	return fmt.Errorf("ffmpeg is not installed")
}

// CheckYtDlp verifies that yt-dlp is installed and accessible
// If not found, attempts to install it using pip
func CheckYtDlp() error {
	// Check if yt-dlp is already installed
	cmd := exec.Command("yt-dlp", "--version")
	if err := cmd.Run(); err == nil {
		return nil // yt-dlp is already installed
	}

	// yt-dlp not found, attempt to install it
	color.Yellow("yt-dlp not found. Installing yt-dlp using pip...")

	// Try pip3 first, then pip
	installCmd := exec.Command("pip3", "install", "yt-dlp")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	pip3Err := installCmd.Run()
	var pip3Output string
	if pip3Err != nil {
		// Capture pip3 error output for better diagnostics
		if exitErr, ok := pip3Err.(*exec.ExitError); ok {
			pip3Output = string(exitErr.Stderr)
		}
	}

	if pip3Err != nil {
		// Try with pip if pip3 failed
		color.Yellow("pip3 failed, trying with pip instead...")
		color.Yellow("pip3 error: %v", pip3Err)
		if pip3Output != "" {
			color.Yellow("pip3 output: %s", pip3Output)
		}

		installCmd = exec.Command("pip", "install", "yt-dlp")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		var pipErr error
		var pipOutput string
		if pipErr = installCmd.Run(); pipErr != nil {
			// Capture pip error output for better diagnostics
			if exitErr, ok := pipErr.(*exec.ExitError); ok {
				pipOutput = string(exitErr.Stderr)
			}

			return fmt.Errorf("failed to install yt-dlp with both pip3 and pip:\n  pip3 error: %v\n  pip3 output: %s\n  pip error: %v\n  pip output: %s\n\nPlease install manually:\n  pip install yt-dlp\nOr download from: https://github.com/yt-dlp/yt-dlp", pip3Err, pip3Output, pipErr, pipOutput)
		}
	}

	color.Green("yt-dlp installed successfully!")

	// Verify installation
	verifyCmd := exec.Command("yt-dlp", "--version")
	if err := verifyCmd.Run(); err != nil {
		var verifyOutput string
		if exitErr, ok := err.(*exec.ExitError); ok {
			verifyOutput = string(exitErr.Stderr)
		}
		return fmt.Errorf("yt-dlp was installed but cannot be found in PATH: %v\n  stderr: %s\nTry running the command again or restart your shell", err, verifyOutput)
	}

	return nil
}
