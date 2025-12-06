package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// IsWSL detects if the program is running under WSL
func IsWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

// Run executes a command and optionally prints output
func Run(cmd string, verbose bool) error {
	if verbose {
		fmt.Printf("Executing: %s\n", cmd)
	}

	command := exec.Command("sh", "-c", cmd)

	if verbose {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	} else {
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
	}

	if err := command.Run(); err != nil {
		// Capture output for better error diagnostics
		var stderrStr string
		if stderrBuffer, ok := command.Stderr.(*bytes.Buffer); ok {
			stderrStr = stderrBuffer.String()
		}

		if stderrStr != "" {
			return fmt.Errorf("command '%s' failed: %w\nstderr: %s", cmd, err, stderrStr)
		}
		return fmt.Errorf("command '%s' failed: %w", cmd, err)
	}

	return nil
}

// RunWithOutput executes a command and returns its output
func RunWithOutput(cmd string) (string, error) {
	command := exec.Command("sh", "-c", cmd)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command '%s' failed: %w\noutput: %s", cmd, err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

// SambaMount mounts a remote Samba share if not already mounted
// Returns the local mount point path
func SambaMount(remoteNode, remoteDrive string, verbose bool) (string, error) {
	slash := ""
	if !strings.HasPrefix(remoteDrive, "/") {
		slash = "/"
	}
	mountPoint := fmt.Sprintf("/mnt/%s%s%s", remoteNode, slash, remoteDrive)

	// Create mount point if it doesn't exist
	if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
		cmd := fmt.Sprintf("sudo mkdir -p %s", mountPoint)
		if err := Run(cmd, verbose); err != nil {
			return "", fmt.Errorf("failed to create mount point '%s': %w", mountPoint, err)
		}
	}

	// Check if already mounted
	isMount, err := IsMountPoint(mountPoint)
	if err != nil {
		return "", fmt.Errorf("failed to check if '%s' is a mount point: %w", mountPoint, err)
	}

	if !isMount {
		cmd := fmt.Sprintf("sudo mount -t drvfs '\\\\%s\\%s' %s", remoteNode, remoteDrive, mountPoint)
		if err := Run(cmd, verbose); err != nil {
			return "", fmt.Errorf("failed to mount '%s' to '%s': %w", remoteDrive, mountPoint, err)
		}
	}

	return mountPoint, nil
}

// SambaParse parses Windows-style paths (e.g., "c:/path/to/file")
// Returns the drive letter and the path
func SambaParse(winPath string) (drive, path string, err error) {
	// Check for multiple colons (invalid Windows path)
	if strings.Count(winPath, ":") != 1 {
		return "", "", fmt.Errorf("invalid Windows path '%s': must contain exactly one colon", winPath)
	}

	parts := strings.SplitN(winPath, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid Windows path '%s': expected format 'drive:/path'", winPath)
	}
	return parts[0], parts[1], nil
}

// WinHome returns the Windows home directory when running under WSL
func WinHome() (string, error) {
	if !IsWSL() {
		return "", fmt.Errorf("not running in WSL")
	}

	// Get Windows home directory
	cmd := exec.Command("cmd.exe", "/c", "echo %UserProfile%")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Windows home: %w", err)
	}
	winPath := strings.TrimSpace(string(output))

	// Convert to WSL path
	cmd = exec.Command("/usr/bin/wslpath", winPath)
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to convert path: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file '%s': %w", src, err)
	}
	defer sourceFile.Close()

	// Create destination with exclusive access to avoid partial writes
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file '%s': %w", dst, err)
	}

	// Copy the file content
	bytesCopied, err := io.Copy(destFile, sourceFile)
	if err != nil {
		// Clean up incomplete destination file
		os.Remove(dst)
		return fmt.Errorf("failed to copy file from '%s' to '%s' (copied %d bytes): %w", src, dst, bytesCopied, err)
	}

	// Ensure data is written to disk
	if err := destFile.Sync(); err != nil {
		// Clean up incomplete destination file
		os.Remove(dst)
		return fmt.Errorf("failed to sync file '%s' to disk: %w", dst, err)
	}

	// Close destination file
	if err := destFile.Close(); err != nil {
		return fmt.Errorf("failed to close destination file '%s': %w", dst, err)
	}

	return nil
}
