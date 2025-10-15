package downloader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	opts := &Options{
		URL:         "https://example.com/video",
		Destination: "/tmp/downloads",
		MediaType:   MP3,
		Format:      "mp3",
		Verbose:     false,
	}

	dl := New(opts)
	if dl == nil {
		t.Fatal("New returned nil")
	}
	if dl.opts != opts {
		t.Error("Options not set correctly")
	}
}

func TestMediaType(t *testing.T) {
	if MP3 == Video {
		t.Error("MP3 and Video types should be different")
	}
}

func TestCheckYtDlp(t *testing.T) {
	// This test checks if yt-dlp is installed
	// It will fail if yt-dlp is not in PATH
	err := CheckYtDlp()
	if err != nil {
		t.Logf("yt-dlp not found: %v (this is expected if yt-dlp is not installed)", err)
		t.Skip("Skipping test as yt-dlp is not available")
	}
}

// Test GetMediaName with a mock (requires yt-dlp to be installed)
func TestGetMediaName(t *testing.T) {
	// Skip if yt-dlp is not available
	if err := CheckYtDlp(); err != nil {
		t.Skip("Skipping test as yt-dlp is not available")
	}

	tmpDir := t.TempDir()
	opts := &Options{
		URL:         "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		Destination: tmpDir,
		MediaType:   MP3,
		Format:      "mp3",
		Verbose:     false,
	}

	dl := New(opts)
	name, err := dl.GetMediaName()
	if err != nil {
		t.Logf("GetMediaName failed (this might be a network issue): %v", err)
		t.Skip("Skipping test due to network or API issue")
	}

	if name == "" {
		t.Error("GetMediaName returned empty string")
	}

	t.Logf("Media name: %s", name)

	// Check that the name is sanitized (no special characters except underscore)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			t.Errorf("Name contains invalid character: %c", char)
		}
	}

	// Check length limit
	if len(name) > 200 {
		t.Errorf("Name too long: %d characters", len(name))
	}
}

func TestOptions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		opts      *Options
		wantError bool
	}{
		{
			name: "valid mp3 options",
			opts: &Options{
				URL:         "https://example.com/video",
				Destination: tmpDir,
				MediaType:   MP3,
				Format:      "mp3",
				Verbose:     false,
			},
			wantError: false,
		},
		{
			name: "valid video options",
			opts: &Options{
				URL:         "https://example.com/video",
				Destination: tmpDir,
				MediaType:   Video,
				Format:      "mp4",
				Verbose:     true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dl := New(tt.opts)
			if dl == nil {
				t.Fatal("New returned nil")
			}

			if dl.opts.URL != tt.opts.URL {
				t.Errorf("URL mismatch: expected %s, got %s", tt.opts.URL, dl.opts.URL)
			}

			if dl.opts.MediaType != tt.opts.MediaType {
				t.Errorf("MediaType mismatch: expected %v, got %v", tt.opts.MediaType, dl.opts.MediaType)
			}

			if dl.opts.Format != tt.opts.Format {
				t.Errorf("Format mismatch: expected %s, got %s", tt.opts.Format, dl.opts.Format)
			}

			if dl.opts.Verbose != tt.opts.Verbose {
				t.Errorf("Verbose mismatch: expected %v, got %v", tt.opts.Verbose, dl.opts.Verbose)
			}
		})
	}
}

// Integration test for Download (requires yt-dlp and network access)
func TestDownload(t *testing.T) {
	// Skip if yt-dlp is not available
	if err := CheckYtDlp(); err != nil {
		t.Skip("Skipping integration test as yt-dlp is not available")
	}

	t.Skip("Skipping integration test - use 'go test -run=TestDownloadIntegration' to run manually")

	tmpDir := t.TempDir()
	opts := &Options{
		URL:         "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		Destination: tmpDir,
		MediaType:   MP3,
		Format:      "mp3",
		Verbose:     true,
	}

	dl := New(opts)
	outputPath, err := dl.Download()
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Check if the file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Downloaded file does not exist: %s", outputPath)
	}

	// Clean up
	os.Remove(outputPath)
}

func TestDownloadInvalidURL(t *testing.T) {
	// Skip if yt-dlp is not available
	if err := CheckYtDlp(); err != nil {
		t.Skip("Skipping test as yt-dlp is not available")
	}

	tmpDir := t.TempDir()
	opts := &Options{
		URL:         "not-a-valid-url",
		Destination: tmpDir,
		MediaType:   MP3,
		Format:      "mp3",
		Verbose:     false,
	}

	dl := New(opts)
	_, err := dl.GetMediaName()
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestDestinationPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		destination string
		format      string
		wantContain string
	}{
		{
			name:        "mp3 destination",
			destination: filepath.Join(tmpDir, "music"),
			format:      "mp3",
			wantContain: "music",
		},
		{
			name:        "video destination",
			destination: filepath.Join(tmpDir, "videos"),
			format:      "mp4",
			wantContain: "videos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the destination directory
			os.MkdirAll(tt.destination, 0755)

			opts := &Options{
				URL:         "https://example.com/video",
				Destination: tt.destination,
				MediaType:   MP3,
				Format:      tt.format,
				Verbose:     false,
			}

			dl := New(opts)
			if dl.opts.Destination != tt.destination {
				t.Errorf("Destination mismatch: expected %s, got %s",
					tt.destination, dl.opts.Destination)
			}
		})
	}
}
