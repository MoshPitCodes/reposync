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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeWithPersisted(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name           string
		baseConfig     *Config
		persistedCfg   *PersistedConfig
		expectedTarget string
		expectedOwner  string
	}{
		{
			name: "env var takes precedence over persisted",
			baseConfig: &Config{
				TargetDir:   "/tmp/env-override",
				GitHubOwner: "env-owner",
			},
			persistedCfg: &PersistedConfig{
				TargetDir:    "~/persisted",
				DefaultOwner: "persisted-owner",
			},
			expectedTarget: "/tmp/env-override",
			expectedOwner:  "env-owner",
		},
		{
			name: "persisted config used when env not set",
			baseConfig: &Config{
				TargetDir:   "",
				GitHubOwner: "",
			},
			persistedCfg: &PersistedConfig{
				TargetDir:    "~/Development",
				DefaultOwner: "moshpitcodes",
			},
			expectedTarget: filepath.Join(homeDir, "Development"),
			expectedOwner:  "moshpitcodes",
		},
		{
			name: "default used when neither env nor persisted set",
			baseConfig: &Config{
				TargetDir:   "",
				GitHubOwner: "",
			},
			persistedCfg: &PersistedConfig{
				TargetDir:    "",
				DefaultOwner: "",
			},
			expectedTarget: filepath.Join(homeDir, "repos"),
			expectedOwner:  "",
		},
		{
			name: "tilde expansion works correctly",
			baseConfig: &Config{
				TargetDir: "",
			},
			persistedCfg: &PersistedConfig{
				TargetDir: "~/custom/path",
			},
			expectedTarget: filepath.Join(homeDir, "custom/path"),
			expectedOwner:  "",
		},
		{
			name: "nil persisted config handled gracefully",
			baseConfig: &Config{
				TargetDir:   "",
				GitHubOwner: "",
			},
			persistedCfg:   nil,
			expectedTarget: filepath.Join(homeDir, "repos"),
			expectedOwner:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := tt.baseConfig.MergeWithPersisted(tt.persistedCfg)

			assert.Equal(t, tt.expectedTarget, merged.TargetDir,
				"TargetDir mismatch")
			assert.Equal(t, tt.expectedOwner, merged.GitHubOwner,
				"GitHubOwner mismatch")
		})
	}
}

func TestExpandTilde(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde only",
			input:    "~",
			expected: homeDir,
		},
		{
			name:     "tilde with path",
			input:    "~/Development",
			expected: filepath.Join(homeDir, "Development"),
		},
		{
			name:     "tilde with nested path",
			input:    "~/foo/bar/baz",
			expected: filepath.Join(homeDir, "foo/bar/baz"),
		},
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "tilde not at start unchanged",
			input:    "path/~/file",
			expected: "path/~/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandTilde(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoad(t *testing.T) {
	// Save original env vars
	origTargetDir := os.Getenv("REPO_SYNC_TARGET_DIR")
	origOwner := os.Getenv("REPO_SYNC_GITHUB_OWNER")
	origSourceDirs := os.Getenv("REPO_SYNC_SOURCE_DIRS")

	// Restore after test
	defer func() {
		os.Setenv("REPO_SYNC_TARGET_DIR", origTargetDir)
		os.Setenv("REPO_SYNC_GITHUB_OWNER", origOwner)
		os.Setenv("REPO_SYNC_SOURCE_DIRS", origSourceDirs)
	}()

	t.Run("loads from environment variables", func(t *testing.T) {
		os.Setenv("REPO_SYNC_TARGET_DIR", "/tmp/test")
		os.Setenv("REPO_SYNC_GITHUB_OWNER", "testowner")
		os.Setenv("REPO_SYNC_SOURCE_DIRS", "/path1:/path2:/path3")

		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "/tmp/test", cfg.TargetDir)
		assert.Equal(t, "testowner", cfg.GitHubOwner)
		assert.Equal(t, []string{"/path1", "/path2", "/path3"}, cfg.SourceDirs)
	})

	t.Run("empty values when env vars not set", func(t *testing.T) {
		os.Unsetenv("REPO_SYNC_TARGET_DIR")
		os.Unsetenv("REPO_SYNC_GITHUB_OWNER")
		os.Unsetenv("REPO_SYNC_SOURCE_DIRS")

		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "", cfg.TargetDir)
		assert.Equal(t, "", cfg.GitHubOwner)
		assert.Nil(t, cfg.SourceDirs)
	})
}
