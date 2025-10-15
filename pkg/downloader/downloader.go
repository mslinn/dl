package downloader

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// MediaType represents the type of media to download
type MediaType int

const (
	MP3 MediaType = iota
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

	// Check for ffmpeg if downloading audio
	if d.opts.MediaType == MP3 {
		if err := CheckFfmpeg(); err != nil {
			return "", err
		}
	}

	var outputPath string
	var args []string

	switch d.opts.MediaType {
	case MP3:
		outputPath = filepath.Join(d.opts.Destination, mediaName)
		args = []string{
			"--extract-audio",
			"--audio-format", d.opts.Format,
			"--output", outputPath + ".%(ext)s",
		}
		if d.opts.Verbose {
			args = append(args, "--verbose", "--progress")
		} else {
			args = append(args, "--quiet", "--no-warnings")
		}
		args = append(args, d.opts.URL)

		fmt.Printf("Saving %s.%s\n", outputPath, d.opts.Format)

	case Video:
		outputPath = filepath.Join(d.opts.Destination, mediaName)
		args = []string{
			"--format", "mp4",
			"--merge-output-format", "mp4",
			"--output", outputPath + ".mp4",
		}
		if d.opts.Verbose {
			args = append(args, "--verbose", "--progress")
		} else {
			args = append(args, "--quiet", "--no-warnings")
		}
		args = append(args, d.opts.URL)

		fmt.Printf("Saving %s.mp4\n", outputPath)
	}

	// Execute yt-dlp
	if d.opts.Verbose {
		fmt.Printf("Running: yt-dlp %s\n", strings.Join(args, " "))
	}

	cmd := exec.Command("yt-dlp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}

	// Determine the actual output filename
	actualPath := outputPath + "." + d.opts.Format
	if d.opts.MediaType == Video {
		actualPath = outputPath + ".mp4"
	}

	// Clean up any .webm files if video
	if d.opts.MediaType == Video {
		webmPath := outputPath + ".webm"
		if _, err := os.Stat(webmPath); err == nil {
			os.Remove(webmPath)
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
	fmt.Println("ffmpeg not found. ffmpeg is required for audio extraction.")
	fmt.Println()
	fmt.Println("Installation instructions:")
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
	fmt.Println("yt-dlp not found. Installing yt-dlp using pip...")

	// Try pip3 first, then pip
	installCmd := exec.Command("pip3", "install", "yt-dlp")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		// Try with pip if pip3 failed
		fmt.Println("Trying with pip instead of pip3...")
		installCmd = exec.Command("pip", "install", "yt-dlp")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install yt-dlp: %w\nPlease install manually: pip install yt-dlp or download from https://github.com/yt-dlp/yt-dlp", err)
		}
	}

	fmt.Println("yt-dlp installed successfully!")

	// Verify installation
	verifyCmd := exec.Command("yt-dlp", "--version")
	if err := verifyCmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp was installed but cannot be found in PATH. Try running the command again or restart your shell")
	}

	return nil
}
