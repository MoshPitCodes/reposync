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

package local

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository represents a local Git repository.
type Repository struct {
	Name     string
	Path     string
	Size     int64
	IsGitRepo bool
	Branch   string
}

// Scanner handles local filesystem repository discovery and operations.
type Scanner struct{}

// NewScanner creates a new local repository scanner.
func NewScanner() *Scanner {
	return &Scanner{}
}

// ScanDirectory recursively scans a directory for Git repositories.
func (s *Scanner) ScanDirectory(rootPath string) ([]Repository, error) {
	var repos []Repository

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip directories we can't access
		}

		// Skip hidden directories except .git
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
			return filepath.SkipDir
		}

		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repo, err := s.analyzeRepo(repoPath)
			if err != nil {
				return nil // Skip repos we can't analyze
			}

			repos = append(repos, *repo)
			return filepath.SkipDir // Don't descend into .git
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return repos, nil
}

// ScanMultipleDirectories scans multiple directories for repositories.
func (s *Scanner) ScanMultipleDirectories(paths []string) ([]Repository, error) {
	var allRepos []Repository

	for _, path := range paths {
		repos, err := s.ScanDirectory(path)
		if err != nil {
			// Log error but continue with other directories
			continue
		}
		allRepos = append(allRepos, repos...)
	}

	return allRepos, nil
}

// analyzeRepo extracts metadata from a Git repository.
func (s *Scanner) analyzeRepo(repoPath string) (*Repository, error) {
	repo := &Repository{
		Name:     filepath.Base(repoPath),
		Path:     repoPath,
		IsGitRepo: true,
	}

	// Get current branch
	branch, err := s.getCurrentBranch(repoPath)
	if err == nil {
		repo.Branch = branch
	}

	// Get repository size
	size, err := s.getDirectorySize(repoPath)
	if err == nil {
		repo.Size = size
	}

	return repo, nil
}

// getCurrentBranch retrieves the current branch name.
func (s *Scanner) getCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// getDirectorySize calculates the total size of a directory.
func (s *Scanner) getDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// CopyRepo copies a Git repository to the target directory.
func (s *Scanner) CopyRepo(sourcePath, targetDir string) error {
	repoName := filepath.Base(sourcePath)
	destPath := filepath.Join(targetDir, repoName)

	// Check if source exists and is a Git repository
	gitDir := filepath.Join(sourcePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("source is not a Git repository: %s", sourcePath)
	}

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("destination already exists: %s", destPath)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Use git clone for proper repository copying
	cmd := exec.Command("git", "clone", sourcePath, destPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Include git's error output in the error message
		errMsg := strings.TrimSpace(string(output))
		if errMsg != "" {
			return fmt.Errorf("git clone failed: %s", errMsg)
		}
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// CopyRepos copies multiple repositories with progress reporting.
func (s *Scanner) CopyRepos(repos []Repository, targetDir string, progressFn func(repo string, success bool, err error)) {
	for _, repo := range repos {
		err := s.CopyRepo(repo.Path, targetDir)
		progressFn(repo.Name, err == nil, err)
	}
}

// IsGitRepository checks if a directory is a Git repository.
func (s *Scanner) IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}
	return false
}

// GetRemoteURL retrieves the remote URL of a Git repository.
func (s *Scanner) GetRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetRepoStatus retrieves the status of a Git repository.
func (s *Scanner) GetRepoStatus(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--short")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	status := strings.TrimSpace(string(output))
	if status == "" {
		return "clean", nil
	}

	return "modified", nil
}

// FormatSize formats a byte size into a human-readable string.
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// RefreshRepo performs a git pull on an existing repository.
func (s *Scanner) RefreshRepo(repoPath string) error {
	// Verify the directory exists and is a git repository
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Run git pull
	cmd := exec.Command("git", "-C", repoPath, "pull")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Include git's error output in the error message
		errMsg := strings.TrimSpace(string(output))
		if errMsg != "" {
			return fmt.Errorf("git pull failed: %s", errMsg)
		}
		return fmt.Errorf("git pull failed: %w", err)
	}

	return nil
}
