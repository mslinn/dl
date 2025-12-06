package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultMethodSCP is the default method for remote transfers
	DefaultMethodSCP = "scp"
)

// Config represents the application configuration
type Config struct {
	ConfigPath    string
	Local         *LocalConfig       `yaml:"local"`
	Remotes       map[string]*Remote `yaml:"remotes"`
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
	expandedPath, err := expandConfigPath(configPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", expandedPath, err)
	}

	cfg, err := unmarshalConfig(data, expandedPath)
	if err != nil {
		return nil, err
	}

	if err := expandEnvVariables(cfg); err != nil {
		return nil, err
	}

	cfg.ActiveRemotes = findActiveRemotes(cfg.Remotes)

	return cfg, nil
}

func expandConfigPath(configPath string) (string, error) {
	expandedPath := os.ExpandEnv(configPath)
	if strings.HasPrefix(expandedPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		expandedPath = filepath.Join(home, expandedPath[2:])
	}
	return expandedPath, nil
}

func unmarshalConfig(data []byte, configPath string) (*Config, error) {
	var cfg Config
	cfg.ConfigPath = configPath

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func expandEnvVariables(cfg *Config) error {
	// Expand environment variables in local paths
	if cfg.Local != nil {
		cfg.Local.MP3s = expandEnvVars(cfg.Local.MP3s)
		cfg.Local.VDest = expandEnvVars(cfg.Local.VDest)
		cfg.Local.XDest = expandEnvVars(cfg.Local.XDest)
	}

	// Expand environment variables in remote paths
	return expandRemoteEnvVariables(cfg.Remotes)
}

func expandRemoteEnvVariables(remotes map[string]*Remote) error {
	if remotes == nil {
		return nil
	}

	for _, remote := range remotes {
		remote.MP3s = expandEnvVars(remote.MP3s)
		remote.VDest = expandEnvVars(remote.VDest)
		remote.XDest = expandEnvVars(remote.XDest)
		if remote.Other != "" {
			remote.Other = expandEnvVars(remote.Other)
		}
		// Default method to scp if not specified
		if remote.Method == "" {
			remote.Method = DefaultMethodSCP
		}
	}

	return nil
}

func findActiveRemotes(remotes map[string]*Remote) map[string]*Remote {
	activeRemotes := make(map[string]*Remote)
	for name, remote := range remotes {
		if !remote.Disabled {
			activeRemotes[name] = remote
		}
	}
	return activeRemotes
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
