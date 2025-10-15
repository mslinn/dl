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
	cmd := exec.Command("yt-dlp", "--dump-json", "--no-warnings", "--no-playlist", d.opts.URL)
	output, err := cmd.Output()
	if err != nil {
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
		if !d.opts.Verbose {
			args = append(args, "--quiet", "--no-warnings")
		} else {
			args = append(args, "--progress")
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
			args = append(args, "--progress")
		} else {
			args = append(args, "--quiet", "--no-warnings")
		}
		args = append(args, d.opts.URL)

		fmt.Printf("Saving %s.mp4\n", outputPath)
	}

	// Execute yt-dlp
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

// CheckYtDlp verifies that yt-dlp is installed and accessible
func CheckYtDlp() error {
	cmd := exec.Command("yt-dlp", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp not found: please install it (e.g., 'pip install yt-dlp' or download from https://github.com/yt-dlp/yt-dlp)")
	}
	return nil
}
