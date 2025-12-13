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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PersistedConfig represents configuration stored in the config file.
type PersistedConfig struct {
	TargetDir       string   `json:"target_dir,omitempty"`
	SourceDirs      []string `json:"source_dirs,omitempty"`
	DefaultOwner    string   `json:"default_owner,omitempty"`
	RecentOwners    []string `json:"recent_owners,omitempty"`
	RecentTemplates []string `json:"recent_templates,omitempty"`
}

// ConfigStore handles persistent storage of configuration.
type ConfigStore struct {
	path string
}

// NewConfigStore creates a new ConfigStore with the default config path.
func NewConfigStore() (*ConfigStore, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config directory: %w", err)
	}

	path := filepath.Join(configDir, "reposync", "config.json")

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &ConfigStore{path: path}, nil
}

// Load reads the persisted configuration from disk.
func (s *ConfigStore) Load() (*PersistedConfig, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &PersistedConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg PersistedConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func (s *ConfigStore) Save(cfg *PersistedConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Path returns the path to the config file.
func (s *ConfigStore) Path() string {
	return s.path
}

// AddRecentOwner adds an owner to the recent owners list.
func (p *PersistedConfig) AddRecentOwner(owner string) {
	// Remove if already exists
	for i, o := range p.RecentOwners {
		if o == owner {
			p.RecentOwners = append(p.RecentOwners[:i], p.RecentOwners[i+1:]...)
			break
		}
	}

	// Add to front
	p.RecentOwners = append([]string{owner}, p.RecentOwners...)

	// Keep only last 10
	if len(p.RecentOwners) > 10 {
		p.RecentOwners = p.RecentOwners[:10]
	}
}

// AddRecentTemplate adds a template to the recent templates list.
// Template format: "owner/repo" for GitHub or "local:path" for local templates.
func (p *PersistedConfig) AddRecentTemplate(template string) {
	// Remove if already exists
	for i, t := range p.RecentTemplates {
		if t == template {
			p.RecentTemplates = append(p.RecentTemplates[:i], p.RecentTemplates[i+1:]...)
			break
		}
	}

	// Add to front
	p.RecentTemplates = append([]string{template}, p.RecentTemplates...)

	// Keep only last 10
	if len(p.RecentTemplates) > 10 {
		p.RecentTemplates = p.RecentTemplates[:10]
	}
}
