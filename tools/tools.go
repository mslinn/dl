//go:build tools
// +build tools

// This file ensures development tools are tracked in go.mod
// but not included in regular builds or installs.
package tools

import (
	_ "dl/cmd/release"
)
