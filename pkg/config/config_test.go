package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	configContent := `local:
  mp3s: /tmp/music
  vdest: /tmp/videos
  xdest: /tmp/xrated

remotes:
  test-remote:
    disabled: false
    method: scp
    mp3s: /data/music
    vdest: /data/videos
    xdest: /data/xrated
  disabled-remote:
    disabled: true
    mp3s: /other/music
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.config")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load the config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test local config
	if cfg.Local == nil {
		t.Fatal("Local config is nil")
	}
	if cfg.Local.MP3s != "/tmp/music" {
		t.Errorf("Expected MP3s to be /tmp/music, got %s", cfg.Local.MP3s)
	}
	if cfg.Local.VDest != "/tmp/videos" {
		t.Errorf("Expected VDest to be /tmp/videos, got %s", cfg.Local.VDest)
	}
	if cfg.Local.XDest != "/tmp/xrated" {
		t.Errorf("Expected XDest to be /tmp/xrated, got %s", cfg.Local.XDest)
	}

	// Test remotes
	if cfg.Remotes == nil {
		t.Fatal("Remotes is nil")
	}
	if len(cfg.Remotes) != 2 {
		t.Errorf("Expected 2 remotes, got %d", len(cfg.Remotes))
	}

	testRemote, ok := cfg.Remotes["test-remote"]
	if !ok {
		t.Fatal("test-remote not found")
	}
	if testRemote.Disabled {
		t.Error("test-remote should not be disabled")
	}
	if testRemote.Method != "scp" {
		t.Errorf("Expected method to be scp, got %s", testRemote.Method)
	}
	if testRemote.MP3s != "/data/music" {
		t.Errorf("Expected MP3s to be /data/music, got %s", testRemote.MP3s)
	}

	// Test active remotes
	if len(cfg.ActiveRemotes) != 1 {
		t.Errorf("Expected 1 active remote, got %d", len(cfg.ActiveRemotes))
	}
	if _, ok := cfg.ActiveRemotes["test-remote"]; !ok {
		t.Error("test-remote should be in active remotes")
	}
	if _, ok := cfg.ActiveRemotes["disabled-remote"]; ok {
		t.Error("disabled-remote should not be in active remotes")
	}
}

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "expand environment variable",
			input:    "$TEST_VAR/music",
			envVar:   "TEST_VAR",
			envValue: "/home/user",
			expected: "/home/user/music",
		},
		{
			name:     "expand with braces",
			input:    "${TEST_VAR}/music",
			envVar:   "TEST_VAR",
			envValue: "/home/user",
			expected: "/home/user/music",
		},
		{
			name:     "no expansion needed",
			input:    "/tmp/music",
			envVar:   "",
			envValue: "",
			expected: "/tmp/music",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			result := expandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetPaths(t *testing.T) {
	cfg := &Config{
		Local: &LocalConfig{
			MP3s:  "/tmp/music",
			VDest: "/tmp/videos",
			XDest: "/tmp/xrated",
		},
	}

	mp3s, err := cfg.GetMP3sPath()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if mp3s != "/tmp/music" {
		t.Errorf("Expected /tmp/music, got %s", mp3s)
	}

	vdest, err := cfg.GetVDestPath()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if vdest != "/tmp/videos" {
		t.Errorf("Expected /tmp/videos, got %s", vdest)
	}

	xdest, err := cfg.GetXDestPath()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if xdest != "/tmp/xrated" {
		t.Errorf("Expected /tmp/xrated, got %s", xdest)
	}
}

func TestGetPathsErrors(t *testing.T) {
	cfg := &Config{
		Local: nil,
	}

	_, err := cfg.GetMP3sPath()
	if err == nil {
		t.Error("Expected error for nil local config")
	}

	_, err = cfg.GetVDestPath()
	if err == nil {
		t.Error("Expected error for nil local config")
	}

	_, err = cfg.GetXDestPath()
	if err == nil {
		t.Error("Expected error for nil local config")
	}
}

func TestGetActiveRemoteNames(t *testing.T) {
	cfg := &Config{
		ActiveRemotes: map[string]*Remote{
			"remote1": {},
			"remote2": {},
		},
	}

	names := cfg.GetActiveRemoteNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	// Check that both names are present (order doesn't matter)
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}
	if !nameMap["remote1"] || !nameMap["remote2"] {
		t.Errorf("Expected remote1 and remote2, got %v", names)
	}
}

func TestDefaultMethod(t *testing.T) {
	configContent := `remotes:
  test-remote:
    disabled: false
    mp3s: /data/music
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.config")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	remote := cfg.Remotes["test-remote"]
	if remote.Method != "scp" {
		t.Errorf("Expected default method to be scp, got %s", remote.Method)
	}
}
