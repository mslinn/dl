//go:build windows

package util

// IsMountPoint checks if a path is a mount point
// On Windows, this always returns false since Samba mounting is not used
func IsMountPoint(path string) bool {
	return false
}
