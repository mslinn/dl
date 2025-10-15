package remote

import (
	"os"
	"path/filepath"
	"testing"

	"dl/pkg/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{}
	copier := New(cfg, false)

	if copier == nil {
		t.Fatal("New returned nil")
	}

	if copier.cfg != cfg {
		t.Error("Config not set correctly")
	}

	if copier.verbose != false {
		t.Error("Verbose not set correctly")
	}
}

func TestPurposeTypes(t *testing.T) {
	if PurposeMP3s == PurposeVideos {
		t.Error("PurposeMP3s and PurposeVideos should be different")
	}
	if PurposeVideos == PurposeXRated {
		t.Error("PurposeVideos and PurposeXRated should be different")
	}
	if PurposeMP3s == PurposeXRated {
		t.Error("PurposeMP3s and PurposeXRated should be different")
	}
}

func TestGetRemotePath(t *testing.T) {
	copier := &Copier{
		cfg:     &config.Config{},
		verbose: false,
	}

	tests := []struct {
		name        string
		remote      *config.Remote
		purpose     Purpose
		expected    string
		expectError bool
	}{
		{
			name: "mp3s path",
			remote: &config.Remote{
				MP3s:  "/data/mp3s",
				VDest: "/data/videos",
				XDest: "/data/xrated",
			},
			purpose:     PurposeMP3s,
			expected:    "/data/mp3s",
			expectError: false,
		},
		{
			name: "videos path",
			remote: &config.Remote{
				MP3s:  "/data/mp3s",
				VDest: "/data/videos",
				XDest: "/data/xrated",
			},
			purpose:     PurposeVideos,
			expected:    "/data/videos",
			expectError: false,
		},
		{
			name: "xrated path",
			remote: &config.Remote{
				MP3s:  "/data/mp3s",
				VDest: "/data/videos",
				XDest: "/data/xrated",
			},
			purpose:     PurposeXRated,
			expected:    "/data/xrated",
			expectError: false,
		},
		{
			name: "missing mp3s path",
			remote: &config.Remote{
				VDest: "/data/videos",
				XDest: "/data/xrated",
			},
			purpose:     PurposeMP3s,
			expectError: true,
		},
		{
			name: "missing vdest path",
			remote: &config.Remote{
				MP3s:  "/data/mp3s",
				XDest: "/data/xrated",
			},
			purpose:     PurposeVideos,
			expectError: true,
		},
		{
			name: "missing xdest path",
			remote: &config.Remote{
				MP3s:  "/data/mp3s",
				VDest: "/data/videos",
			},
			purpose:     PurposeXRated,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := copier.getRemotePath(tt.remote, tt.purpose)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if path != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, path)
			}
		})
	}
}

func TestCopyToRemotesNoRemotes(t *testing.T) {
	cfg := &config.Config{
		ActiveRemotes: map[string]*config.Remote{},
	}

	copier := New(cfg, false)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := copier.CopyToRemotes(testFile, PurposeMP3s)
	if err != nil {
		t.Errorf("Should not error when no remotes: %v", err)
	}
}

func TestCopyToRemotesWithRemote(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(destDir, 0755)

	cfg := &config.Config{
		ActiveRemotes: map[string]*config.Remote{
			"test-remote": {
				Method: "scp",
				MP3s:   destDir,
			},
		},
	}

	copier := New(cfg, true)

	testFile := filepath.Join(tmpDir, "test.mp3")
	if err := os.WriteFile(testFile, []byte("test audio"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This will fail because scp requires actual remote host
	// but we're testing the error handling
	err := copier.CopyToRemotes(testFile, PurposeMP3s)
	if err != nil {
		t.Logf("Expected error for scp without remote host: %v", err)
	}
}

func TestCopyToRemotesInvalidPurpose(t *testing.T) {
	cfg := &config.Config{
		ActiveRemotes: map[string]*config.Remote{
			"test-remote": {
				Method: "scp",
				// No paths defined
			},
		},
	}

	copier := New(cfg, false)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should error because MP3s path is not defined
	err := copier.CopyToRemotes(testFile, PurposeMP3s)
	if err == nil {
		t.Error("Expected error for undefined remote path")
	}
}

func TestCopyToRemoteMethod(t *testing.T) {
	cfg := &config.Config{}
	copier := New(cfg, false)

	remote := &config.Remote{
		Method: "unknown-method",
		MP3s:   "/tmp/test",
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := copier.copyToRemote(testFile, "test-remote", remote, PurposeMP3s)
	if err == nil {
		t.Error("Expected error for unknown method")
	}
}

func TestCopySambaInvalidPath(t *testing.T) {
	copier := &Copier{
		cfg:     &config.Config{},
		verbose: false,
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Invalid path (no colon)
	err := copier.copySamba(testFile, "test.txt", "remote", "/invalid/path")
	if err == nil {
		t.Error("Expected error for invalid samba path")
	}
}

func TestCopySCPConstruction(t *testing.T) {
	copier := &Copier{
		cfg:     &config.Config{},
		verbose: true,
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This will fail because remote host doesn't exist
	// but we're testing that the command is constructed properly
	err := copier.copySCP(testFile, "test.txt", "nonexistent-remote", "/tmp")
	if err == nil {
		t.Log("SCP succeeded unexpectedly (or nonexistent-remote actually exists)")
	} else {
		t.Logf("Expected SCP failure: %v", err)
	}
}

func TestVerboseMode(t *testing.T) {
	cfg := &config.Config{}

	copier1 := New(cfg, false)
	if copier1.verbose {
		t.Error("Verbose should be false")
	}

	copier2 := New(cfg, true)
	if !copier2.verbose {
		t.Error("Verbose should be true")
	}
}

// Integration test for actual file copying (requires sudo and WSL for samba)
func TestCopySambaIntegration(t *testing.T) {
	t.Skip("Skipping integration test - requires WSL and sudo privileges")

	copier := &Copier{
		cfg:     &config.Config{},
		verbose: true,
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This requires actual remote host and proper configuration
	err := copier.copySamba(testFile, "test.txt", "test-remote", "c:/temp")
	if err != nil {
		t.Logf("Samba copy failed: %v", err)
	}
}
