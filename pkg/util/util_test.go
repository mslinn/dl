package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsWSL(t *testing.T) {
	// This test will pass or fail depending on the environment
	// We're just testing that the function runs without error
	result := IsWSL()
	t.Logf("IsWSL returned: %v", result)
}

func TestSambaParse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedDrive string
		expectedPath  string
		expectError   bool
	}{
		{
			name:          "valid windows path with forward slash",
			input:         "c:/path/to/file",
			expectedDrive: "c",
			expectedPath:  "/path/to/file",
			expectError:   false,
		},
		{
			name:          "valid windows path no leading slash",
			input:         "e:media/videos",
			expectedDrive: "e",
			expectedPath:  "media/videos",
			expectError:   false,
		},
		{
			name:        "invalid path no colon",
			input:       "invalid/path",
			expectError: true,
		},
		{
			name:        "invalid path multiple colons",
			input:       "c:/path:with:colons",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drive, path, err := SambaParse(tt.input)
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

			if drive != tt.expectedDrive {
				t.Errorf("Expected drive %s, got %s", tt.expectedDrive, drive)
			}
			if path != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, path)
			}
		})
	}
}

func TestExpandEnvVars(t *testing.T) {
	// Set a test environment variable
	testVar := "TEST_UTIL_VAR"
	testValue := "/test/path"
	os.Setenv(testVar, testValue)
	defer os.Unsetenv(testVar)

	result := os.ExpandEnv("$TEST_UTIL_VAR/music")
	expected := "/test/path/music"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy the file
	dstPath := filepath.Join(tmpDir, "dest.txt")
	if err := CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify the destination file exists and has the same content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("Content mismatch. Expected %s, got %s", content, dstContent)
	}
}

func TestCopyFileErrors(t *testing.T) {
	tmpDir := t.TempDir()

	// Test copying non-existent file
	err := CopyFile("/nonexistent/file.txt", filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("Expected error when copying non-existent file")
	}

	// Test copying to invalid destination
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	err = CopyFile(srcPath, "/invalid/path/dest.txt")
	if err == nil {
		t.Error("Expected error when copying to invalid destination")
	}
}

func TestRun(t *testing.T) {
	// Test successful command
	err := Run("echo 'test'", false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test with verbose
	err = Run("echo 'test verbose'", true)
	if err != nil {
		t.Errorf("Unexpected error with verbose: %v", err)
	}

	// Test failing command
	err = Run("nonexistent-command-xyz", false)
	if err == nil {
		t.Error("Expected error for non-existent command")
	}
}

func TestRunWithOutput(t *testing.T) {
	output, err := RunWithOutput("echo 'hello'")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if output != "hello" {
		t.Errorf("Expected 'hello', got '%s'", output)
	}

	// Test failing command
	_, err = RunWithOutput("nonexistent-command-xyz")
	if err == nil {
		t.Error("Expected error for non-existent command")
	}
}

func TestIsMountPoint(t *testing.T) {
	// Test with root directory (should be a mount point)
	result := IsMountPoint("/")
	t.Logf("IsMountPoint('/') returned: %v", result)

	// Test with /tmp (usually not a separate mount point, but might be)
	result = IsMountPoint("/tmp")
	t.Logf("IsMountPoint('/tmp') returned: %v", result)

	// Test with non-existent path
	result = IsMountPoint("/nonexistent/path/xyz")
	if result {
		t.Error("Non-existent path should not be a mount point")
	}
}

// Note: SambaMount and WinHome tests are skipped as they require
// specific system configuration (WSL, sudo privileges, actual remote hosts)
// These should be tested manually or in integration tests

func TestSambaMountInvalidPath(t *testing.T) {
	// This test will likely fail unless running with sudo and on WSL
	// It's here for documentation purposes
	t.Skip("Skipping SambaMount test as it requires sudo and WSL")

	_, err := SambaMount("nonexistent-host", "c", false)
	if err == nil {
		t.Error("Expected error for invalid host")
	}
}

func TestWinHomeNotWSL(t *testing.T) {
	if IsWSL() {
		t.Skip("Skipping test as we are running in WSL")
	}

	_, err := WinHome()
	if err == nil {
		t.Error("Expected error when not running in WSL")
	}
}
