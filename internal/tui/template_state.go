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

import "strings"

// TemplateWorkflowStep represents the current step in the template sync workflow.
type TemplateWorkflowStep int

const (
	// StepSelectTemplate is the step where user selects a template repository.
	StepSelectTemplate TemplateWorkflowStep = iota
	// StepBrowseTree is the step where user browses and selects files from template.
	StepBrowseTree
	// StepSelectTargets is the step where user selects target local repositories.
	StepSelectTargets
	// StepSyncing is the step where the sync operation is in progress.
	StepSyncing
	// StepComplete is the step shown after sync completes.
	StepComplete
)

// String returns the string representation of the workflow step.
func (s TemplateWorkflowStep) String() string {
	switch s {
	case StepSelectTemplate:
		return "Select Template"
	case StepBrowseTree:
		return "Browse Files"
	case StepSelectTargets:
		return "Select Targets"
	case StepSyncing:
		return "Syncing"
	case StepComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}

// TemplateSyncProgress tracks the progress of a template sync operation.
type TemplateSyncProgress struct {
	Current     int
	Total       int
	CurrentFile string
	TargetRepo  string
	Synced      int
	Skipped     int
	Errors      int
}

// TemplateSyncState holds all state for the template sync workflow.
type TemplateSyncState struct {
	// Current step in the workflow
	Step TemplateWorkflowStep

	// Template source type (GitHub or Local)
	IsLocal bool

	// GitHub template repository information
	TemplateOwner  string
	TemplateRepo   string
	TemplateBranch string

	// Local template path
	LocalTemplatePath string

	// Tree data
	TreeRoot *TemplateTreeNode

	// Selected files/folders for sync (paths)
	SelectedPaths []string

	// Target local repository paths
	TargetRepos []string

	// Conflict handling state
	OverwriteAll bool
	SkipAll      bool

	// Sync progress and results
	SyncProgress TemplateSyncProgress

	// Sync results (deprecated, use SyncProgress)
	SyncedCount  int
	SkippedCount int
	ErrorCount   int
}

// NewTemplateSyncState creates a new template sync state initialized to the first step.
func NewTemplateSyncState() *TemplateSyncState {
	return &TemplateSyncState{
		Step:          StepSelectTemplate,
		SelectedPaths: make([]string, 0),
		TargetRepos:   make([]string, 0),
	}
}

// Reset clears the state and returns to the first step.
func (s *TemplateSyncState) Reset() {
	s.Step = StepSelectTemplate
	s.IsLocal = false
	s.TemplateOwner = ""
	s.TemplateRepo = ""
	s.TemplateBranch = ""
	s.LocalTemplatePath = ""
	s.TreeRoot = nil
	s.SelectedPaths = make([]string, 0)
	s.TargetRepos = make([]string, 0)
	s.OverwriteAll = false
	s.SkipAll = false
	s.SyncedCount = 0
	s.SkippedCount = 0
	s.ErrorCount = 0
}

// SetTemplate sets the template repository information (GitHub).
func (s *TemplateSyncState) SetTemplate(owner, repo, branch string) {
	s.IsLocal = false
	s.TemplateOwner = owner
	s.TemplateRepo = repo
	s.TemplateBranch = branch
	s.LocalTemplatePath = ""
}

// SetLocalTemplate sets the local template path.
func (s *TemplateSyncState) SetLocalTemplate(path string) {
	s.IsLocal = true
	s.LocalTemplatePath = path
	s.TemplateOwner = ""
	s.TemplateRepo = ""
	s.TemplateBranch = ""
}

// GetTemplateFullName returns the "owner/repo" format or local path.
func (s *TemplateSyncState) GetTemplateFullName() string {
	if s.IsLocal {
		return s.LocalTemplatePath
	}
	if s.TemplateOwner == "" || s.TemplateRepo == "" {
		return ""
	}
	return s.TemplateOwner + "/" + s.TemplateRepo
}

// GetTemplateDisplayName returns a user-friendly display name.
func (s *TemplateSyncState) GetTemplateDisplayName() string {
	if s.IsLocal {
		// Show just the last directory name for brevity
		parts := strings.Split(s.LocalTemplatePath, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1] + " (local)"
		}
		return s.LocalTemplatePath
	}
	if s.TemplateOwner == "" || s.TemplateRepo == "" {
		return ""
	}
	return s.TemplateOwner + "/" + s.TemplateRepo
}

// IsTargetSameAsTemplate checks if a target repo is the same as the template.
// This prevents syncing a local template to itself.
func (s *TemplateSyncState) IsTargetSameAsTemplate(targetPath string) bool {
	if !s.IsLocal {
		return false // GitHub templates can sync to any local repo
	}
	// Normalize paths for comparison
	return normalizePath(targetPath) == normalizePath(s.LocalTemplatePath)
}

// normalizePath normalizes a path by removing trailing slashes.
func normalizePath(path string) string {
	return strings.TrimSuffix(path, "/")
}

// HasTemplate returns true if a template is selected (GitHub or local).
func (s *TemplateSyncState) HasTemplate() bool {
	if s.IsLocal {
		return s.LocalTemplatePath != ""
	}
	return s.TemplateOwner != "" && s.TemplateRepo != ""
}

// HasSelectedFiles returns true if any files are selected for sync.
func (s *TemplateSyncState) HasSelectedFiles() bool {
	return len(s.SelectedPaths) > 0
}

// HasTargetRepos returns true if any target repositories are selected.
func (s *TemplateSyncState) HasTargetRepos() bool {
	return len(s.TargetRepos) > 0
}

// NextStep advances to the next workflow step.
func (s *TemplateSyncState) NextStep() {
	if s.Step < StepComplete {
		s.Step++
	}
}

// PrevStep goes back to the previous workflow step.
func (s *TemplateSyncState) PrevStep() {
	if s.Step > StepSelectTemplate {
		s.Step--
	}
}
