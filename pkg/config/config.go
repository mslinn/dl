package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	ConfigPath    string
	Local         *LocalConfig         `yaml:"local"`
	Remotes       map[string]*Remote   `yaml:"remotes"`
	ActiveRemotes map[string]*Remote
}

// LocalConfig represents local file paths
type LocalConfig struct {
	MP3s  string `yaml:"mp3s"`
	VDest string `yaml:"vdest"`
	XDest string `yaml:"xdest"`
}

// Remote represents a remote destination
type Remote struct {
	Disabled bool   `yaml:"disabled"`
	Method   string `yaml:"method"`
	MP3s     string `yaml:"mp3s"`
	Other    string `yaml:"other"`
	VDest    string `yaml:"vdest"`
	XDest    string `yaml:"xdest"`
}

// Load reads and parses the configuration file
func Load(configPath string) (*Config, error) {
	expandedPath := os.ExpandEnv(configPath)
	if strings.HasPrefix(expandedPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		expandedPath = filepath.Join(home, expandedPath[2:])
	}

	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", expandedPath, err)
	}

	var cfg Config
	cfg.ConfigPath = expandedPath

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand environment variables in local paths
	if cfg.Local != nil {
		cfg.Local.MP3s = expandEnvVars(cfg.Local.MP3s)
		cfg.Local.VDest = expandEnvVars(cfg.Local.VDest)
		cfg.Local.XDest = expandEnvVars(cfg.Local.XDest)
	}

	// Expand environment variables in remote paths
	if cfg.Remotes != nil {
		for _, remote := range cfg.Remotes {
			remote.MP3s = expandEnvVars(remote.MP3s)
			remote.VDest = expandEnvVars(remote.VDest)
			remote.XDest = expandEnvVars(remote.XDest)
			if remote.Other != "" {
				remote.Other = expandEnvVars(remote.Other)
			}
			// Default method to scp if not specified
			if remote.Method == "" {
				remote.Method = "scp"
			}
		}
	}

	// Find active remotes
	cfg.ActiveRemotes = make(map[string]*Remote)
	for name, remote := range cfg.Remotes {
		if !remote.Disabled {
			cfg.ActiveRemotes[name] = remote
		}
	}

	return &cfg, nil
}

// expandEnvVars expands environment variables in a string
func expandEnvVars(s string) string {
	return os.ExpandEnv(s)
}

// GetMP3sPath returns the MP3s path from local config
func (c *Config) GetMP3sPath() (string, error) {
	if c.Local == nil || c.Local.MP3s == "" {
		return "", fmt.Errorf("mp3s path not defined in config")
	}
	return c.Local.MP3s, nil
}

// GetVDestPath returns the video destination path from local config
func (c *Config) GetVDestPath() (string, error) {
	if c.Local == nil || c.Local.VDest == "" {
		return "", fmt.Errorf("vdest path not defined in config")
	}
	return c.Local.VDest, nil
}

// GetXDestPath returns the x-rated video destination path from local config
func (c *Config) GetXDestPath() (string, error) {
	if c.Local == nil || c.Local.XDest == "" {
		return "", fmt.Errorf("xdest path not defined in config")
	}
	return c.Local.XDest, nil
}

// GetActiveRemoteNames returns a list of active remote names
func (c *Config) GetActiveRemoteNames() []string {
	names := make([]string, 0, len(c.ActiveRemotes))
	for name := range c.ActiveRemotes {
		names = append(names, name)
	}
	return names
}
