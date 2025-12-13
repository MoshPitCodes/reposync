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

// Mode messages

// SwitchModeMsg is sent to switch between different view modes.
type SwitchModeMsg struct {
	Mode ViewMode
}

// SelectOwnerMsg is sent when an owner is selected.
type SelectOwnerMsg struct {
	Owner string
	IsOrg bool
}

// Data messages

// ReposLoadedMsg is sent when repositories are successfully loaded.
type ReposLoadedMsg struct {
	Items []ListItem
}

// OrgsLoadedMsg is sent when organizations are successfully loaded.
type OrgsLoadedMsg struct {
	Orgs []string
}

// LoadErrorMsg is sent when an error occurs during data loading.
type LoadErrorMsg struct {
	Err error
}

// Settings messages

// SettingsOpenMsg is sent to open the settings overlay.
type SettingsOpenMsg struct{}

// SettingsCloseMsg is sent to close the settings overlay.
type SettingsCloseMsg struct {
	Save bool
}

// Sync messages

// StartSyncMsg is sent to start the sync process.
type StartSyncMsg struct {
	Repos []string
}

// SyncProgressMsg is sent to update sync progress.
type SyncProgressMsg struct {
	Current int
	Total   int
	Repo    string
	Success bool
	Err     error
}

// SyncCompleteMsg is sent when sync is complete.
type SyncCompleteMsg struct {
	Results []SyncResult
}

// SyncResult represents the result of syncing a single repository.
type SyncResult struct {
	Repo    string
	Success bool
	Error   error
}

// Owner selector messages

// ToggleOwnerSelectorMsg is sent to toggle the owner selector dropdown.
type ToggleOwnerSelectorMsg struct{}

// CloseOwnerSelectorMsg is sent to close the owner selector dropdown.
type CloseOwnerSelectorMsg struct{}

// Help messages

// ToggleHelpMsg is sent to toggle the help overlay.
type ToggleHelpMsg struct{}

// Repository exists messages

// ExistsAction represents the action to take when a repository exists.
type ExistsAction int

const (
	ActionSkip ExistsAction = iota
	ActionRefresh
	ActionSkipAll
	ActionRefreshAll
)

// RepoExistsMsg is sent when a repository already exists during sync.
type RepoExistsMsg struct {
	RepoName  string
	RepoPath  string
	RepoIndex int
	Mode      string // "github" or "local"
}

// RepoExistsResponseMsg is sent in response to a repository exists prompt.
type RepoExistsResponseMsg struct {
	Action    ExistsAction
	RepoIndex int
}

// Template workflow messages

// TemplateRepoSelectedMsg is sent when a template repository is selected.
type TemplateRepoSelectedMsg struct {
	Owner     string // For GitHub templates
	Repo      string // For GitHub templates
	LocalPath string // For local templates (mutually exclusive with Owner/Repo)
	IsLocal   bool   // True if this is a local template
}

// TemplateTreeNode represents a file or folder in the template repository tree.
type TemplateTreeNode struct {
	Path     string
	Name     string
	IsDir    bool
	SHA      string
	Size     int64
	Children []*TemplateTreeNode
	Expanded bool
	Selected bool
}

// TemplateTreeLoadedMsg is sent when the repository tree is fetched.
type TemplateTreeLoadedMsg struct {
	Root *TemplateTreeNode
	Err  error
}

// TemplateTargetsSelectedMsg is sent when target local repos are chosen.
type TemplateTargetsSelectedMsg struct {
	TargetPaths []string
}

// TemplateConflictMsg is sent when a file conflict is detected during sync.
type TemplateConflictMsg struct {
	FilePath       string
	TargetRepoPath string
	TemplateSize   int64
	LocalSize      int64
}

// TemplateConflictAction represents the user's choice for handling a conflict.
type TemplateConflictAction int

const (
	ConflictOverwrite TemplateConflictAction = iota
	ConflictSkip
	ConflictOverwriteAll
	ConflictSkipAll
)

// TemplateConflictResponseMsg is sent in response to a conflict prompt.
type TemplateConflictResponseMsg struct {
	Action   TemplateConflictAction
	FilePath string
}

// TemplateSyncProgressMsg reports sync progress.
type TemplateSyncProgressMsg struct {
	Current     int
	Total       int
	CurrentFile string
	TargetRepo  string
}

// TemplateSyncCompleteMsg is sent when template sync finishes.
type TemplateSyncCompleteMsg struct {
	Synced  int
	Skipped int
	Errors  int
}

// TemplateStepChangeMsg is sent when the template workflow step changes.
type TemplateStepChangeMsg struct {
	Step int
}
