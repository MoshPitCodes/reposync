// Copyright 2024-2025 MoshPitCodes
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	// TargetDir is the default directory where repositories will be cloned/copied
	TargetDir string

	// GitHubOwner is the default GitHub user or organization name
	GitHubOwner string

	// SourceDirs is a list of local directories to scan for repositories
	SourceDirs []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{}

	// REPO_SYNC_TARGET_DIR: Default target directory for cloning/copying
	// Only set from env if explicitly provided (leave empty otherwise)
	cfg.TargetDir = os.Getenv("REPO_SYNC_TARGET_DIR")

	// REPO_SYNC_GITHUB_OWNER: Default GitHub owner/org
	cfg.GitHubOwner = os.Getenv("REPO_SYNC_GITHUB_OWNER")

	// REPO_SYNC_SOURCE_DIRS: Colon-separated list of source directories to scan
	sourceDirsEnv := os.Getenv("REPO_SYNC_SOURCE_DIRS")
	if sourceDirsEnv != "" {
		cfg.SourceDirs = strings.Split(sourceDirsEnv, ":")
	}

	return cfg, nil
}

// GetTargetDir returns the target directory, creating it if it doesn't exist.
func (c *Config) GetTargetDir() (string, error) {
	if err := os.MkdirAll(c.TargetDir, 0o755); err != nil {
		return "", err
	}
	return c.TargetDir, nil
}

// MergeWithPersisted merges the persisted config with this config.
// Environment variables take precedence over persisted values.
func (c *Config) MergeWithPersisted(p *PersistedConfig) *Config {
	merged := &Config{
		TargetDir:   c.TargetDir,
		GitHubOwner: c.GitHubOwner,
		SourceDirs:  c.SourceDirs,
	}

	// Use persisted values only if environment variables are not set
	// Priority: 1) Env vars, 2) Persisted config, 3) Defaults
	if merged.TargetDir == "" {
		if p != nil && p.TargetDir != "" {
			merged.TargetDir = expandTilde(p.TargetDir)
		} else {
			// Default if neither env nor persisted
			homeDir, _ := os.UserHomeDir()
			merged.TargetDir = filepath.Join(homeDir, "repos")
		}
	}

	if merged.GitHubOwner == "" && p != nil && p.DefaultOwner != "" {
		merged.GitHubOwner = p.DefaultOwner
	}

	if len(merged.SourceDirs) == 0 && p != nil && len(p.SourceDirs) > 0 {
		merged.SourceDirs = p.SourceDirs
	}

	return merged
}

// expandTilde expands ~ to the user's home directory.
func expandTilde(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if len(path) == 1 {
		return homeDir
	}

	if path[1] == '/' {
		return filepath.Join(homeDir, path[2:])
	}

	return path
}
