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

package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/MoshPitCodes/reposync/internal/github"
	"github.com/MoshPitCodes/reposync/internal/local"
)

// InlineProgressModel manages inline progress display during sync.
type InlineProgressModel struct {
	// Complex types first
	progressBar progress.Model
	spinner     spinner.Model

	// Slices (24 bytes each)
	repos   []string
	results []SyncResult

	// Strings (16 bytes each)
	targetDir   string
	mode        string // "github" or "local"
	currentRepo string

	// Time (24 bytes each)
	startTime time.Time
	endTime   time.Time

	// Ints (8 bytes each)
	current        int
	total          int
	pendingRepoIdx int

	// Bools (1 byte each, grouped together)
	running        bool
	complete       bool
	skipAll        bool
	refreshAll     bool
	waitingForUser bool
}

// NewInlineProgressModel creates a new inline progress model.
func NewInlineProgressModel() *InlineProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	p := progress.New(progress.WithDefaultGradient())

	return &InlineProgressModel{
		progressBar: p,
		spinner:     s,
		running:     false,
		complete:    false,
	}
}

// Start begins the sync process.
func (m *InlineProgressModel) Start(repos []string, targetDir, mode string) tea.Cmd {
	m.repos = repos
	m.targetDir = targetDir
	m.mode = mode
	m.current = 0
	m.total = len(repos)
	m.running = true
	m.complete = false
	m.results = []SyncResult{}
	m.startTime = time.Now()
	m.pendingRepoIdx = 0
	m.skipAll = false
	m.refreshAll = false
	m.waitingForUser = false

	return tea.Batch(
		m.spinner.Tick,
		m.syncNextRepo(),
	)
}

// syncNextRepo syncs the next repository in the queue.
func (m *InlineProgressModel) syncNextRepo() tea.Cmd {
	return func() tea.Msg {
		// Check if we're done
		if m.pendingRepoIdx >= len(m.repos) {
			return SyncCompleteMsg{Results: m.results}
		}

		if m.mode == "github" {
			return m.syncNextGitHubRepo()
		}
		return m.syncNextLocalRepo()
	}
}

// syncNextGitHubRepo synchronizes the next GitHub repository.
func (m *InlineProgressModel) syncNextGitHubRepo() tea.Msg {
	// Validate target directory
	if m.targetDir == "" {
		return SyncCompleteMsg{
			Results: []SyncResult{{Repo: "sync", Success: false, Error: fmt.Errorf("target directory not set")}},
		}
	}

	client, err := github.NewClient()
	if err != nil {
		return SyncCompleteMsg{
			Results: []SyncResult{{Repo: "sync", Success: false, Error: fmt.Errorf("failed to create GitHub client: %w", err)}},
		}
	}

	fullName := m.repos[m.pendingRepoIdx]
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		m.results = append(m.results, SyncResult{
			Repo:    fullName,
			Success: false,
			Error:   fmt.Errorf("invalid repository format (expected owner/repo, got %q)", fullName),
		})
		m.pendingRepoIdx++
		m.current++
		return m.syncNextRepo()()
	}

	owner, repoName := parts[0], parts[1]
	repoPath := filepath.Join(m.targetDir, repoName)

	// Check if repository already exists
	if _, err := os.Stat(repoPath); err == nil {
		// Repository exists
		if m.skipAll {
			// Skip this repository
			m.results = append(m.results, SyncResult{
				Repo:    repoName,
				Success: true,
				Error:   nil,
			})
			m.pendingRepoIdx++
			m.current++
			return m.syncNextRepo()()
		} else if m.refreshAll {
			// Refresh this repository
			err := client.RefreshRepo(repoPath)
			m.results = append(m.results, SyncResult{
				Repo:    repoName,
				Success: err == nil,
				Error:   err,
			})
			m.pendingRepoIdx++
			m.current++
			return m.syncNextRepo()()
		} else {
			// Prompt user
			return RepoExistsMsg{
				RepoName:  repoName,
				RepoPath:  repoPath,
				RepoIndex: m.pendingRepoIdx,
				Mode:      "github",
			}
		}
	}

	// Repository doesn't exist - clone it
	err = client.CloneRepo(owner, repoName, m.targetDir)
	m.results = append(m.results, SyncResult{
		Repo:    repoName,
		Success: err == nil,
		Error:   err,
	})
	m.pendingRepoIdx++
	m.current++

	return m.syncNextRepo()()
}

// syncNextLocalRepo synchronizes the next local repository.
func (m *InlineProgressModel) syncNextLocalRepo() tea.Msg {
	// Validate target directory
	if m.targetDir == "" {
		return SyncCompleteMsg{
			Results: []SyncResult{{Repo: "sync", Success: false, Error: fmt.Errorf("target directory not set")}},
		}
	}

	scanner := local.NewScanner()
	repoPath := m.repos[m.pendingRepoIdx]
	repoName := repoPath
	if parts := strings.Split(repoPath, "/"); len(parts) > 0 {
		repoName = parts[len(parts)-1]
	}

	destPath := filepath.Join(m.targetDir, repoName)

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		// Repository exists
		if m.skipAll {
			// Skip this repository
			m.results = append(m.results, SyncResult{
				Repo:    repoName,
				Success: true,
				Error:   nil,
			})
			m.pendingRepoIdx++
			m.current++
			return m.syncNextRepo()()
		} else if m.refreshAll {
			// Refresh this repository
			err := scanner.RefreshRepo(destPath)
			m.results = append(m.results, SyncResult{
				Repo:    repoName,
				Success: err == nil,
				Error:   err,
			})
			m.pendingRepoIdx++
			m.current++
			return m.syncNextRepo()()
		} else {
			// Prompt user
			return RepoExistsMsg{
				RepoName:  repoName,
				RepoPath:  destPath,
				RepoIndex: m.pendingRepoIdx,
				Mode:      "local",
			}
		}
	}

	// Repository doesn't exist - copy it
	err := scanner.CopyRepo(repoPath, m.targetDir)
	m.results = append(m.results, SyncResult{
		Repo:    repoName,
		Success: err == nil,
		Error:   err,
	})
	m.pendingRepoIdx++
	m.current++

	return m.syncNextRepo()()
}

// Update handles messages for the progress component.
func (m *InlineProgressModel) Update(msg tea.Msg) (*InlineProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case SyncProgressMsg:
		m.current = msg.Current
		m.currentRepo = msg.Repo
		return m, nil

	case SyncCompleteMsg:
		m.results = msg.Results
		m.current = m.total
		m.running = false
		m.complete = true
		m.endTime = time.Now()
		return m, nil

	case RepoExistsResponseMsg:
		// Handle user's response to repository exists dialog
		return m.handleRepoExistsResponse(msg)

	case spinner.TickMsg:
		if m.running {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// handleRepoExistsResponse processes the user's response to a repository exists prompt.
func (m *InlineProgressModel) handleRepoExistsResponse(msg RepoExistsResponseMsg) (*InlineProgressModel, tea.Cmd) {
	// Update flags based on action
	switch msg.Action {
	case ActionSkipAll:
		m.skipAll = true
	case ActionRefreshAll:
		m.refreshAll = true
	}

	// Get the repository info
	var repoName, repoPath string
	if m.mode == "github" {
		fullName := m.repos[m.pendingRepoIdx]
		parts := strings.Split(fullName, "/")
		if len(parts) == 2 {
			repoName = parts[1]
			repoPath = filepath.Join(m.targetDir, repoName)
		}
	} else {
		sourcePath := m.repos[m.pendingRepoIdx]
		if parts := strings.Split(sourcePath, "/"); len(parts) > 0 {
			repoName = parts[len(parts)-1]
			repoPath = filepath.Join(m.targetDir, repoName)
		}
	}

	// Handle the action
	switch msg.Action {
	case ActionSkip, ActionSkipAll:
		// Skip this repository
		m.results = append(m.results, SyncResult{
			Repo:    repoName,
			Success: true,
			Error:   nil,
		})
		m.pendingRepoIdx++
		m.current++
		return m, m.syncNextRepo()

	case ActionRefresh, ActionRefreshAll:
		// Refresh (git pull) this repository
		var err error
		if m.mode == "github" {
			client, clientErr := github.NewClient()
			if clientErr != nil {
				err = clientErr
			} else {
				err = client.RefreshRepo(repoPath)
			}
		} else {
			scanner := local.NewScanner()
			err = scanner.RefreshRepo(repoPath)
		}

		m.results = append(m.results, SyncResult{
			Repo:    repoName,
			Success: err == nil,
			Error:   err,
		})
		m.pendingRepoIdx++
		m.current++
		return m, m.syncNextRepo()
	}

	return m, nil
}

// View renders the inline progress bar.
func (m *InlineProgressModel) View() string {
	if !m.running && !m.complete {
		return ""
	}

	var b strings.Builder

	if m.running {
		// Show spinner and current operation
		percent := float64(m.current) / float64(m.total)
		percentText := fmt.Sprintf("%.0f%%", percent*100)

		b.WriteString(m.spinner.View() + " ")
		b.WriteString(progressBarStyle.Render(m.progressBar.ViewAs(percent)))
		b.WriteString(" " + progressTextStyle.Render(percentText))
		b.WriteString(" • ")
		b.WriteString(progressTextStyle.Render(fmt.Sprintf("%d/%d synced", m.current, m.total)))

		if m.currentRepo != "" {
			elapsed := time.Since(m.startTime)
			b.WriteString(" • ")
			b.WriteString(progressTextStyle.Render(formatDuration(elapsed)))
		}
	}

	if m.complete {
		// Show completion summary
		successCount := 0
		var failures []SyncResult
		for _, result := range m.results {
			if result.Success {
				successCount++
			} else {
				failures = append(failures, result)
			}
		}

		elapsed := m.endTime.Sub(m.startTime)
		if successCount == m.total {
			b.WriteString(RenderSuccess(fmt.Sprintf("✓ %d/%d synced • %s", successCount, m.total, formatDuration(elapsed))))
		} else {
			failCount := m.total - successCount
			b.WriteString(RenderWarning(fmt.Sprintf("⚠ %d succeeded, %d failed • %s", successCount, failCount, formatDuration(elapsed))))

			// Show error details for failed repos
			if len(failures) > 0 {
				b.WriteString("\n\n")
				b.WriteString(RenderError("Failed repositories:"))
				b.WriteString("\n")
				for _, failure := range failures {
					errMsg := "unknown error"
					if failure.Error != nil {
						errMsg = failure.Error.Error()
					}
					b.WriteString(RenderError(fmt.Sprintf("  • %s: %s", failure.Repo, errMsg)))
					b.WriteString("\n")
				}
			}
		}
	}

	return b.String()
}

// IsRunning returns whether the sync is currently running.
func (m *InlineProgressModel) IsRunning() bool {
	return m.running
}

// IsComplete returns whether the sync is complete.
func (m *InlineProgressModel) IsComplete() bool {
	return m.complete
}

// GetResults returns the sync results.
func (m *InlineProgressModel) GetResults() []SyncResult {
	return m.results
}

// Reset resets the progress model.
func (m *InlineProgressModel) Reset() {
	m.running = false
	m.complete = false
	m.current = 0
	m.total = 0
	m.currentRepo = ""
	m.results = []SyncResult{}
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", mins, secs)
	} else {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
}
