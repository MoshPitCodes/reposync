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
