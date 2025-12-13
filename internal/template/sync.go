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

// Package template provides template sync functionality.
package template

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/MoshPitCodes/reposync/internal/github"
)

// ConflictAction represents the action to take when a file conflict occurs.
type ConflictAction int

const (
	// ActionOverwrite replaces the existing file.
	ActionOverwrite ConflictAction = iota
	// ActionSkip keeps the existing file.
	ActionSkip
	// ActionOverwriteAll replaces all conflicting files.
	ActionOverwriteAll
	// ActionSkipAll keeps all existing files.
	ActionSkipAll
)

// SyncResult represents the result of syncing a single file.
type SyncResult struct {
	FilePath   string
	TargetRepo string
	Success    bool
	Skipped    bool
	Error      error
}

// SyncEngine handles template synchronization.
type SyncEngine struct {
	// GitHub client for fetching remote files
	githubClient *github.Client

	// Template source information (GitHub)
	templateOwner  string
	templateRepo   string
	templateBranch string

	// Template source information (Local)
	localTemplatePath string
	isLocal           bool

	// Batch conflict actions
	overwriteAll bool
	skipAll      bool
}

// NewSyncEngine creates a sync engine for GitHub templates.
func NewSyncEngine(client *github.Client, owner, repo, branch string) *SyncEngine {
	return &SyncEngine{
		githubClient:   client,
		templateOwner:  owner,
		templateRepo:   repo,
		templateBranch: branch,
		isLocal:        false,
	}
}

// NewLocalSyncEngine creates a sync engine for local templates.
func NewLocalSyncEngine(localPath string) *SyncEngine {
	return &SyncEngine{
		localTemplatePath: localPath,
		isLocal:           true,
	}
}

// SetOverwriteAll sets the overwrite all flag.
func (e *SyncEngine) SetOverwriteAll(val bool) {
	e.overwriteAll = val
}

// SetSkipAll sets the skip all flag.
func (e *SyncEngine) SetSkipAll(val bool) {
	e.skipAll = val
}

// ShouldOverwriteAll returns whether all conflicts should be overwritten.
func (e *SyncEngine) ShouldOverwriteAll() bool {
	return e.overwriteAll
}

// ShouldSkipAll returns whether all conflicts should be skipped.
func (e *SyncEngine) ShouldSkipAll() bool {
	return e.skipAll
}

// CheckConflict checks if a file already exists at the target path.
func (e *SyncEngine) CheckConflict(filePath, targetRepoPath string) (bool, error) {
	destPath := filepath.Join(targetRepoPath, filePath)
	_, err := os.Stat(destPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check file: %w", err)
	}
	return true, nil
}

// SyncFile downloads/copies a file from the template and writes it to the target.
func (e *SyncEngine) SyncFile(filePath, targetRepoPath string) error {
	destPath := filepath.Join(targetRepoPath, filePath)

	// Create parent directories
	parentDir := filepath.Dir(destPath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	var content []byte
	var err error

	if e.isLocal {
		// Read from local template
		sourcePath := filepath.Join(e.localTemplatePath, filePath)
		content, err = os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read source file %s: %w", sourcePath, err)
		}
	} else {
		// Fetch from GitHub
		content, err = e.githubClient.GetFileContent(
			e.templateOwner,
			e.templateRepo,
			filePath,
			e.templateBranch,
		)
		if err != nil {
			return fmt.Errorf("failed to fetch file from GitHub: %w", err)
		}
	}

	// Write to destination
	if err := os.WriteFile(destPath, content, 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

// CopyLocalFile copies a file from local template to target.
func (e *SyncEngine) CopyLocalFile(filePath, targetRepoPath string) error {
	sourcePath := filepath.Join(e.localTemplatePath, filePath)
	destPath := filepath.Join(targetRepoPath, filePath)

	// Create parent directories
	parentDir := filepath.Dir(destPath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	// Open source file
	src, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Get source file info for permissions
	srcInfo, err := src.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create destination file
	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// SyncProgress represents progress information for the sync operation.
type SyncProgress struct {
	Current     int
	Total       int
	CurrentFile string
	TargetRepo  string
}

// ConflictInfo represents information about a file conflict.
type ConflictInfo struct {
	FilePath   string
	TargetRepo string
}

// SyncFiles syncs multiple files to multiple targets with callbacks.
// progressFn is called for each file synced.
// conflictFn is called when a conflict is detected and returns the action to take.
func (e *SyncEngine) SyncFiles(
	files []string,
	targets []string,
	progressFn func(progress SyncProgress),
	conflictFn func(conflict ConflictInfo) ConflictAction,
) (results []SyncResult) {
	results = make([]SyncResult, 0)
	total := len(files) * len(targets)
	current := 0

	for _, targetRepo := range targets {
		for _, filePath := range files {
			current++

			// Report progress
			if progressFn != nil {
				progressFn(SyncProgress{
					Current:     current,
					Total:       total,
					CurrentFile: filePath,
					TargetRepo:  targetRepo,
				})
			}

			result := SyncResult{
				FilePath:   filePath,
				TargetRepo: targetRepo,
			}

			// Check for conflict
			hasConflict, err := e.CheckConflict(filePath, targetRepo)
			if err != nil {
				result.Error = err
				results = append(results, result)
				continue
			}

			if hasConflict {
				// Determine action
				var action ConflictAction

				if e.overwriteAll {
					action = ActionOverwrite
				} else if e.skipAll {
					action = ActionSkip
				} else if conflictFn != nil {
					action = conflictFn(ConflictInfo{
						FilePath:   filePath,
						TargetRepo: targetRepo,
					})

					// Update batch flags
					if action == ActionOverwriteAll {
						e.overwriteAll = true
						action = ActionOverwrite
					} else if action == ActionSkipAll {
						e.skipAll = true
						action = ActionSkip
					}
				} else {
					// Default to skip if no callback
					action = ActionSkip
				}

				if action == ActionSkip {
					result.Skipped = true
					result.Success = true
					results = append(results, result)
					continue
				}
			}

			// Sync the file
			if e.isLocal {
				err = e.CopyLocalFile(filePath, targetRepo)
			} else {
				err = e.SyncFile(filePath, targetRepo)
			}

			if err != nil {
				result.Error = err
			} else {
				result.Success = true
			}

			results = append(results, result)
		}
	}

	return results
}

// GetSyncSummary returns a summary of sync results.
func GetSyncSummary(results []SyncResult) (synced, skipped, errors int) {
	for _, r := range results {
		if r.Error != nil {
			errors++
		} else if r.Skipped {
			skipped++
		} else if r.Success {
			synced++
		}
	}
	return
}
