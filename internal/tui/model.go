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

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MoshPitCodes/reposync/internal/config"
	"github.com/MoshPitCodes/reposync/internal/github"
	"github.com/MoshPitCodes/reposync/internal/local"
)

// Model is the main Bubble Tea model with unified single-view architecture.
type Model struct {
	// Pointers (8 bytes each)
	config           *config.Config
	store            *config.ConfigStore
	tabs             *TabBarModel
	list             *ListModel
	settings         *SettingsModel
	progress         *InlineProgressModel
	ownerSelector    *OwnerSelectorModel
	repoExistsDialog *RepoExistsDialogModel
	githubClient     *github.Client

	// Slices (24 bytes)
	orgs []string

	// Strings (16 bytes each)
	owner    string
	username string

	// Ints (8 bytes each)
	width  int
	height int

	// Enum (platform-dependent, typically 4-8 bytes)
	mode ViewMode

	// Bools (1 byte each, grouped together)
	showSettings bool
	showHelp     bool
	syncing      bool
	quitting     bool
}

// NewModel creates a new unified model starting in Personal mode.
func NewModel(cfg *config.Config) (Model, error) {
	store, err := config.NewConfigStore()
	if err != nil {
		return Model{}, err
	}

	// Load persisted config and merge
	persistedCfg, err := store.Load()
	if err != nil {
		persistedCfg = &config.PersistedConfig{}
	}
	mergedCfg := cfg.MergeWithPersisted(persistedCfg)

	// Initialize GitHub client and get username
	client, err := github.NewClient()
	if err != nil {
		return Model{}, err
	}

	username, err := client.GetCurrentUser()
	if err != nil {
		return Model{}, err
	}

	// Determine initial owner
	owner := username
	if mergedCfg.GitHubOwner != "" {
		owner = mergedCfg.GitHubOwner
	}

	// Create list model and set compact mode from persisted config
	list := NewListModel()
	list.SetCompactMode(persistedCfg.CompactMode)

	return Model{
		config:           mergedCfg,
		store:            store,
		mode:             ModePersonal,
		owner:            owner,
		username:         username,
		orgs:             []string{},
		tabs:             NewTabBarModel(),
		list:             list,
		settings:         NewSettingsModel(store),
		progress:         NewInlineProgressModel(),
		ownerSelector:    NewOwnerSelectorModel(username),
		repoExistsDialog: NewRepoExistsDialogModel(),
		showSettings:     false,
		showHelp:         false,
		syncing:          false,
		quitting:         false,
		githubClient:     client,
	}, nil
}

// NewGitHubModel creates a model that starts in GitHub mode with a specific owner.
func NewGitHubModel(cfg *config.Config, owner string) (Model, error) {
	model, err := NewModel(cfg)
	if err != nil {
		return model, err
	}

	model.owner = owner
	model.mode = ModePersonal
	model.tabs.SetActive(ModePersonal)

	return model, nil
}

// NewLocalModel creates a model that starts in Local mode.
func NewLocalModel(cfg *config.Config) (Model, error) {
	model, err := NewModel(cfg)
	if err != nil {
		return model, err
	}

	model.mode = ModeLocal
	model.tabs.SetActive(ModeLocal)

	return model, nil
}

// Init initializes the model and loads initial data.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadOrgs(),
		m.loadRepositories(),
	)
}

// loadOrgs loads the user's organizations.
func (m *Model) loadOrgs() tea.Cmd {
	return func() tea.Msg {
		orgs, err := m.githubClient.ListUserOrgs()
		if err != nil {
			return LoadErrorMsg{Err: err}
		}
		return OrgsLoadedMsg{Orgs: orgs}
	}
}

// loadRepositories loads repositories based on the current mode.
func (m *Model) loadRepositories() tea.Cmd {
	return func() tea.Msg {
		switch m.mode {
		case ModePersonal:
			repos, err := m.githubClient.ListUserRepos(m.username)
			if err != nil {
				return LoadErrorMsg{Err: err}
			}
			return ReposLoadedMsg{Items: FromGitHubRepos(repos)}

		case ModeOrganization:
			repos, err := m.githubClient.ListOrgRepos(m.owner)
			if err != nil {
				return LoadErrorMsg{Err: err}
			}
			return ReposLoadedMsg{Items: FromGitHubRepos(repos)}

		case ModeLocal:
			scanner := local.NewScanner()
			repos, err := scanner.ScanMultipleDirectories(m.config.SourceDirs)
			if err != nil {
				return LoadErrorMsg{Err: err}
			}
			return ReposLoadedMsg{Items: FromLocalRepos(repos)}
		}
		return nil
	}
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	var cmds []tea.Cmd

	// Handle window size
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		m.settings.SetSize(msg.Width, msg.Height)
		return m, nil
	}

	// Handle settings overlay
	if m.showSettings {
		return m.updateSettings(msg)
	}

	// Handle repository exists dialog
	if m.repoExistsDialog.IsVisible() {
		return m.updateRepoExistsDialog(msg)
	}

	// Handle owner selector
	if m.ownerSelector.IsExpanded() {
		return m.updateOwnerSelector(msg)
	}

	// Handle help overlay
	if m.showHelp {
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "?" || msg.String() == "esc" {
				m.showHelp = false
			}
		}
		return m, nil
	}

	// Handle global key messages
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "c":
			if !m.syncing {
				m.showSettings = true
				return m, nil
			}

		case "o":
			if m.mode != ModeLocal && !m.syncing {
				m.ownerSelector.Toggle()
				return m, nil
			}

		case "enter":
			if !m.syncing {
				return m.startSync()
			}
		}
	}

	// Handle custom messages
	switch msg := msg.(type) {
	case OrgsLoadedMsg:
		m.orgs = msg.Orgs
		m.ownerSelector.SetOrgs(msg.Orgs)
		return m, nil

	case ReposLoadedMsg:
		m.list.SetItems(msg.Items)
		m.list.SetLoading(false)
		return m, nil

	case LoadErrorMsg:
		m.list.SetError(msg.Err)
		return m, nil

	case SwitchModeMsg:
		m.mode = msg.Mode
		m.list.SetLoading(true)

		// Handle owner switching based on mode
		switch msg.Mode {
		case ModePersonal:
			// Reset owner to username when switching to personal mode
			m.owner = m.username
			return m, m.loadRepositories()

		case ModeOrganization:
			// Check if organizations are available
			if len(m.orgs) == 0 {
				m.list.SetError(fmt.Errorf("no organizations found - use 'o' to select an owner"))
				m.list.SetLoading(false)
				return m, nil
			}
			// Set owner to first organization
			m.owner = m.orgs[0]
			return m, m.loadRepositories()

		case ModeLocal:
			// Local mode doesn't use owner
			return m, m.loadRepositories()
		}

		return m, m.loadRepositories()

	case SelectOwnerMsg:
		m.owner = msg.Owner
		m.list.SetLoading(true)
		if msg.IsOrg {
			m.mode = ModeOrganization
			m.tabs.SetActive(ModeOrganization)
		} else {
			m.mode = ModePersonal
			m.tabs.SetActive(ModePersonal)
		}
		return m, m.loadRepositories()

	case SyncCompleteMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		m.syncing = false
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case RepoExistsMsg:
		// Show the dialog when a repository exists
		m.repoExistsDialog.Show(msg.RepoName, msg.RepoPath, msg.RepoIndex, msg.Mode)
		return m, nil

	case RepoExistsResponseMsg:
		// Forward the response to the progress model
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// Update progress if syncing
	if m.syncing {
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update tabs
	var tabCmd tea.Cmd
	m.tabs, tabCmd = m.tabs.Update(msg)
	cmds = append(cmds, tabCmd)

	// Update list
	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	cmds = append(cmds, listCmd)

	return m, tea.Batch(cmds...)
}

// updateSettings handles updates when settings overlay is open.
func (m Model) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if we received a SettingsCloseMsg
	if closeMsg, ok := msg.(SettingsCloseMsg); ok {
		m.showSettings = false
		if closeMsg.Save {
			// Reload config
			persistedCfg, err := m.store.Load()
			if err != nil {
				persistedCfg = &config.PersistedConfig{}
			}
			m.config = m.config.MergeWithPersisted(persistedCfg)
			// Update list compact mode from saved config
			m.list.SetCompactMode(persistedCfg.CompactMode)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.settings, cmd = m.settings.Update(msg)
	return m, cmd
}

// updateOwnerSelector handles updates when owner selector is expanded.
func (m Model) updateOwnerSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.ownerSelector, cmd = m.ownerSelector.Update(msg)
	return m, cmd
}

// updateRepoExistsDialog handles updates when repository exists dialog is visible.
func (m Model) updateRepoExistsDialog(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.repoExistsDialog, cmd = m.repoExistsDialog.Update(msg)
	return m, cmd
}

// startSync initiates the sync process.
func (m Model) startSync() (tea.Model, tea.Cmd) {
	selectedItems := m.list.GetSelectedItems()
	if len(selectedItems) == 0 {
		return m, nil
	}

	targetDir, err := m.config.GetTargetDir()
	if err != nil {
		// Create a sync complete message with the error
		m.syncing = true
		return m, func() tea.Msg {
			return SyncCompleteMsg{
				Results: []SyncResult{{
					Repo:    "config",
					Success: false,
					Error:   fmt.Errorf("failed to get target directory: %w", err),
				}},
			}
		}
	}

	mode := "github"
	if m.mode == ModeLocal {
		mode = "local"
	}

	m.syncing = true
	return m, m.progress.Start(selectedItems, targetDir, mode)
}

// View renders the complete unified view - delegated to view.go.
func (m Model) View() string {
	if m.quitting {
		return RenderSuccess("Thanks for using repo-sync!\n")
	}

	// This will be implemented in view.go
	return m.renderView()
}
