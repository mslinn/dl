//go:build unix

package util

import (
	"os"
	"path/filepath"
	"syscall"
)

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
