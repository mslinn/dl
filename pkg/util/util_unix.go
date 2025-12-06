//go:build unix

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// IsMountPoint checks if a path is a mount point
func IsMountPoint(path string) (bool, error) {
	// Get stat info for the path
	pathStat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Path doesn't exist, not a mount point
		}
		return false, fmt.Errorf("failed to stat path '%s': %w", path, err) // Permission denied, corrupted filesystem, etc.
	}

	// Get stat info for the parent
	parentPath := filepath.Dir(path)
	parentStat, err := os.Stat(parentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("parent directory '%s' does not exist for path '%s'", parentPath, path)
		}
		return false, fmt.Errorf("failed to stat parent directory '%s': %w", parentPath, err)
	}

	// If the device IDs are different, it's a mount point
	pathSys := pathStat.Sys()
	parentSys := parentStat.Sys()

	pathStatT, ok := pathSys.(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("failed to get syscall info for path '%s'", path)
	}
	parentStatT, ok := parentSys.(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("failed to get syscall info for parent '%s'", parentPath)
	}

	pathDev := pathStatT.Dev
	parentDev := parentStatT.Dev

	return pathDev != parentDev, nil
}
