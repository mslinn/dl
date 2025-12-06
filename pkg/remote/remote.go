package remote

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"dl/pkg/config"
	"dl/pkg/util"

	"github.com/fatih/color"
)

// Purpose defines the type of media being copied
type Purpose int

const (
	// PurposeMP3s indicates copying MP3 files
	PurposeMP3s Purpose = iota
	// PurposeVideos indicates copying video files
	PurposeVideos
	// PurposeXRated indicates copying adult content
	PurposeXRated
)

// Copier handles copying files to remote destinations
type Copier struct {
	cfg     *config.Config
	verbose bool
}

// New creates a new Copier
func New(cfg *config.Config, verbose bool) *Copier {
	return &Copier{
		cfg:     cfg,
		verbose: verbose,
	}
}

// CopyToRemotes copies a file to all active remote destinations concurrently
func (c *Copier) CopyToRemotes(localPath string, purpose Purpose) error {
	if len(c.cfg.ActiveRemotes) == 0 {
		if c.verbose {
			fmt.Println("No active remotes configured")
		}
		return nil
	}

	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	// Use mutex to protect the errors slice
	var mu sync.Mutex
	var errors []error
	var successCount int

	// Launch a goroutine for each remote destination
	for remoteName, remote := range c.cfg.ActiveRemotes {
		wg.Add(1)
		go func(name string, r *config.Remote) {
			defer wg.Done()
			c.processRemoteCopy(localPath, name, r, purpose, &mu, &errors, &successCount)
		}(remoteName, remote)
	}

	// Wait for all copies to complete
	wg.Wait()

	// Report summary
	fmt.Printf("\nRemote copy summary: %d succeeded, %d failed out of %d remotes\n", successCount, len(errors), len(c.cfg.ActiveRemotes))

	// Return error only if all copies failed
	if len(errors) == len(c.cfg.ActiveRemotes) && len(errors) > 0 {
		return fmt.Errorf("failed to copy to any remote destination (all %d attempts failed)", len(errors))
	}

	return nil
}

// processRemoteCopy handles the copy operation for a single remote destination
func (c *Copier) processRemoteCopy(localPath, remoteName string, remote *config.Remote, purpose Purpose, mu *sync.Mutex, errors *[]error, successCount *int) {
	if err := c.copyToRemote(localPath, remoteName, remote, purpose); err != nil {
		diagnosticMsg := c.getErrorDiagnostic(err)
		fmt.Printf("WARNING: Failed to copy to remote '%s': %s%s\n", remoteName, err.Error(), diagnosticMsg)

		mu.Lock()
		*errors = append(*errors, err)
		mu.Unlock()
	} else {
		fmt.Printf("SUCCESS: Successfully copied to remote '%s'\n", remoteName)

		mu.Lock()
		*successCount++
		mu.Unlock()
	}
}

// getErrorDiagnostic provides specific error diagnosis for common issues
func (c *Copier) getErrorDiagnostic(err error) string {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "permission denied") || strings.Contains(errMsg, "Permission denied"):
		return " - Check SSH keys or Samba permissions"
	case strings.Contains(errMsg, "No such file or directory") || strings.Contains(errMsg, "No route to host"):
		return " - Check network connectivity and remote host availability"
	case strings.Contains(errMsg, "Connection refused") || strings.Contains(errMsg, "Connection timed out"):
		return " - Check if SSH service is running on remote host"
	case strings.Contains(errMsg, "failed to mount"):
		return " - Check if Samba share is accessible and WSL is configured properly"
	case strings.Contains(errMsg, "command failed"):
		return " - Check if required tools (scp/mount) are installed"
	default:
		return ""
	}
}

// copyToRemote copies a file to a single remote destination
func (c *Copier) copyToRemote(localPath, remoteName string, remote *config.Remote, purpose Purpose) error {
	// Get the appropriate remote path based on purpose
	remotePath, err := c.getRemotePath(remote, purpose)
	if err != nil {
		return fmt.Errorf("configuration error for remote '%s': %w", remoteName, err)
	}

	filename := filepath.Base(localPath)

	switch remote.Method {
	case "samba":
		return c.copySamba(localPath, filename, remoteName, remotePath)
	case "scp":
		return c.copySCP(localPath, filename, remoteName, remotePath)
	default:
		return fmt.Errorf("unknown copy method '%s' for remote '%s'. Supported methods: 'samba', 'scp'", remote.Method, remoteName)
	}
}

// getRemotePath returns the appropriate path based on the purpose
func (c *Copier) getRemotePath(remote *config.Remote, purpose Purpose) (string, error) {
	switch purpose {
	case PurposeMP3s:
		if remote.MP3s == "" {
			return "", fmt.Errorf("remote does not define mp3s path")
		}
		return remote.MP3s, nil
	case PurposeVideos:
		if remote.VDest == "" {
			return "", fmt.Errorf("remote does not define vdest path")
		}
		return remote.VDest, nil
	case PurposeXRated:
		if remote.XDest == "" {
			return "", fmt.Errorf("remote does not define xdest path")
		}
		return remote.XDest, nil
	default:
		return "", fmt.Errorf("unknown purpose")
	}
}

// copySamba copies a file using Samba/CIFS mount
func (c *Copier) copySamba(localPath, filename, remoteName, remotePath string) error {
	remoteDrive, localPathOnRemote, err := util.SambaParse(remotePath)
	if err != nil {
		return fmt.Errorf("invalid Samba path '%s': %w", remotePath, err)
	}

	sambaRoot, err := util.SambaMount(remoteName, remoteDrive, c.verbose)
	if err != nil {
		return fmt.Errorf("failed to mount Samba share '%s' (drive %s): %w", remoteName, remoteDrive, err)
	}

	targetPath := filepath.Join(sambaRoot, localPathOnRemote, filename)
	color.Cyan("Copying to %s using samba\n", targetPath)

	if err := util.CopyFile(localPath, targetPath); err != nil {
		return fmt.Errorf("failed to copy file to '%s': %w", targetPath, err)
	}

	return nil
}

// copySCP copies a file using scp
func (c *Copier) copySCP(localPath, filename, remoteName, remotePath string) error {
	target := fmt.Sprintf("%s:%s/%s", remoteName, remotePath, filename)
	fmt.Printf("Copying to %s using scp\n", target)

	cmd := fmt.Sprintf("scp %s %s:%s", localPath, remoteName, remotePath)
	if err := util.Run(cmd, c.verbose); err != nil {
		return fmt.Errorf("failed to copy file using SCP from '%s' to '%s': %w", localPath, target, err)
	}

	return nil
}
