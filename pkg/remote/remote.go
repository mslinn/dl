package remote

import (
	"fmt"
	"path/filepath"
	"sync"

	"dl/pkg/config"
	"dl/pkg/util"
	"github.com/fatih/color"
)

// Purpose defines the type of media being copied
type Purpose int

const (
	PurposeMP3s Purpose = iota
	PurposeVideos
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

	// Launch a goroutine for each remote destination
	for remoteName, remote := range c.cfg.ActiveRemotes {
		wg.Add(1)
		go func(name string, r *config.Remote) {
			defer wg.Done()
			if err := c.copyToRemote(localPath, name, r, purpose); err != nil {
				fmt.Printf("Warning: failed to copy to %s: %v\n", name, err)
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}(remoteName, remote)
	}

	// Wait for all copies to complete
	wg.Wait()

	// Return error only if all copies failed
	if len(errors) == len(c.cfg.ActiveRemotes) && len(errors) > 0 {
		return fmt.Errorf("failed to copy to any remote")
	}

	return nil
}

// copyToRemote copies a file to a single remote destination
func (c *Copier) copyToRemote(localPath, remoteName string, remote *config.Remote, purpose Purpose) error {
	// Get the appropriate remote path based on purpose
	remotePath, err := c.getRemotePath(remote, purpose)
	if err != nil {
		return err
	}

	filename := filepath.Base(localPath)

	switch remote.Method {
	case "samba":
		return c.copySamba(localPath, filename, remoteName, remotePath)
	case "scp":
		return c.copySCP(localPath, filename, remoteName, remotePath)
	default:
		return fmt.Errorf("unknown copy method: %s", remote.Method)
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
		return err
	}

	sambaRoot, err := util.SambaMount(remoteName, remoteDrive, c.verbose)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(sambaRoot, localPathOnRemote, filename)
	color.Cyan("Copying to %s using samba\n", targetPath)

	return util.CopyFile(localPath, targetPath)
}

// copySCP copies a file using scp
func (c *Copier) copySCP(localPath, filename, remoteName, remotePath string) error {
	target := fmt.Sprintf("%s:%s/%s", remoteName, remotePath, filename)
	fmt.Printf("Copying to %s using scp\n", target)

	cmd := fmt.Sprintf("scp %s %s:%s", localPath, remoteName, remotePath)
	return util.Run(cmd, c.verbose)
}
