// Package config handles loading and validating Tailscale auth keys.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration.
type Config struct {
	AuthKey string
}

// Options configures how the auth key is loaded.
type Options struct {
	KeyFilePath string // Path to auth key file on USB
	EnvVarName  string // Env var name (default: "TS_AUTHKEY")
}

// ValidateAuthKey checks whether key is a valid Tailscale auth key.
// Valid keys start with "tskey-auth-" and contain only lowercase characters
// with no whitespace.
func ValidateAuthKey(key string) error {
	if key == "" {
		return fmt.Errorf("auth key is empty")
	}
	if strings.ContainsAny(key, " \t\n\r") {
		return fmt.Errorf("auth key contains whitespace")
	}
	if !strings.HasPrefix(key, "tskey-auth-") {
		return fmt.Errorf("auth key has invalid prefix, must start with 'tskey-auth-'")
	}
	if key != strings.ToLower(key) {
		return fmt.Errorf("auth key must be lowercase")
	}
	return nil
}

// Load reads the auth key from the environment variable or file and returns
// a validated Config. The environment variable takes precedence over the file.
// The key is stored in memory only and never written to disk.
func Load(opts Options) (*Config, error) {
	envName := opts.EnvVarName
	if envName == "" {
		envName = "TS_AUTHKEY"
	}

	var key string

	// Environment variable takes precedence over file.
	if envVal := os.Getenv(envName); envVal != "" {
		key = envVal
	} else if opts.KeyFilePath != "" {
		data, err := os.ReadFile(opts.KeyFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("auth key file not found: %s: %w", opts.KeyFilePath, err)
			}
			return nil, fmt.Errorf("reading auth key file: %w", err)
		}
		key = strings.TrimSpace(string(data))
	} else {
		return nil, fmt.Errorf("no auth key source configured: set %s or provide a key file path", envName)
	}

	if key == "" {
		return nil, fmt.Errorf("auth key is empty")
	}

	if err := ValidateAuthKey(key); err != nil {
		return nil, fmt.Errorf("invalid auth key format: %w", err)
	}

	return &Config{AuthKey: key}, nil
}
