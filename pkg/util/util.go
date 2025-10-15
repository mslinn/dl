package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
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
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// RunWithOutput executes a command and returns its output
func RunWithOutput(cmd string) (string, error) {
	command := exec.Command("sh", "-c", cmd)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
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
			return "", fmt.Errorf("failed to create mount point: %w", err)
		}
	}

	// Check if already mounted
	if !IsMountPoint(mountPoint) {
		cmd := fmt.Sprintf("sudo mount -t drvfs '\\\\%s\\%s' %s", remoteNode, remoteDrive, mountPoint)
		if err := Run(cmd, verbose); err != nil {
			return "", fmt.Errorf("failed to mount: %w", err)
		}
	}

	return mountPoint, nil
}

// IsMountPoint checks if a path is a mount point
func IsMountPoint(path string) bool {
	// Get stat info for the path
	pathStat, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Get stat info for the parent
	parentPath := filepath.Dir(path)
	parentStat, err := os.Stat(parentPath)
	if err != nil {
		return false
	}

	// If the device IDs are different, it's a mount point
	pathDev := pathStat.Sys().(*syscall.Stat_t).Dev
	parentDev := parentStat.Sys().(*syscall.Stat_t).Dev

	return pathDev != parentDev
}

// SambaParse parses Windows-style paths (e.g., "c:/path/to/file")
// Returns the drive letter and the path
func SambaParse(winPath string) (string, string, error) {
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
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return destFile.Sync()
}
