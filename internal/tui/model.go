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

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MoshPitCodes/reposync/internal/config"
	"github.com/MoshPitCodes/reposync/internal/github"
	"github.com/MoshPitCodes/reposync/internal/local"
	"github.com/MoshPitCodes/reposync/internal/template"
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

	// Template mode components
	templateState    *TemplateSyncState
	templateSelector *TemplateSelectorModel
	templateTree     *TemplateTreeModel
	templateTargets  *TemplateTargetsModel
	templateConflict *TemplateConflictModel
	templateEngine   *template.SyncEngine

	// Slices (24 bytes)
	orgs            []string
	localRepoPaths  []string // Cached local repo paths for template targets

	// Strings (16 bytes each)
	owner    string
	username string

	// Ints (8 bytes each)
	width          int
	height         int
	headerHeight   int
	tabsHeight     int
	ownerBarHeight int
	footerHeight   int
	listHeight     int

	// Enum (platform-dependent, typically 4-8 bytes)
	mode ViewMode

	// Bools (1 byte each, grouped together)
	showSettings     bool
	showHelp         bool
	syncing          bool
	quitting         bool
	templateSyncing  bool

	// Channel for template sync progress updates
	templateSyncProgressChan chan tea.Msg
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

	// Create list model
	list := NewListModel()

	// Load recent templates from persisted config
	recentTemplates := make([]string, 0)
	if persistedCfg != nil && len(persistedCfg.RecentTemplates) > 0 {
		recentTemplates = persistedCfg.RecentTemplates
	}

	return Model{
		config:           mergedCfg,
		store:            store,
		mode:             ModePersonal,
		owner:            owner,
		username:         username,
		orgs:             []string{},
		localRepoPaths:   []string{},
		tabs:             NewTabBarModel(),
		list:             list,
		settings:         NewSettingsModel(store),
		progress:         NewInlineProgressModel(),
		ownerSelector:    NewOwnerSelectorModel(username),
		repoExistsDialog: NewRepoExistsDialogModel(),
		templateState:    NewTemplateSyncState(),
		templateSelector: NewTemplateSelectorModel(recentTemplates),
		templateTree:     nil, // Created when tree is loaded
		templateTargets:  NewTemplateTargetsModel(),
		templateConflict: NewTemplateConflictModel(),
		templateEngine:   nil, // Created when sync starts
		showSettings:     false,
		showHelp:         false,
		syncing:          false,
		quitting:         false,
		templateSyncing:  false,
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

		case ModeTemplate:
			// Template mode doesn't load a repo list - it shows a workflow
			// Return empty to indicate workflow mode
			return ReposLoadedMsg{Items: []ListItem{}}
		}
		return nil
	}
}

// loadLocalReposForTemplateTargets loads local repositories as potential template targets.
func (m *Model) loadLocalReposForTemplateTargets() tea.Cmd {
	return func() tea.Msg {
		scanner := local.NewScanner()
		repos, err := scanner.ScanMultipleDirectories(m.config.SourceDirs)
		if err != nil {
			return LoadErrorMsg{Err: err}
		}

		// Extract paths
		paths := make([]string, len(repos))
		for i, repo := range repos {
			paths[i] = repo.Path
		}

		return TemplateTargetsLoadedMsg{Paths: paths}
	}
}

// TemplateTargetsLoadedMsg is sent when local repos are loaded for template targets.
type TemplateTargetsLoadedMsg struct {
	Paths []string
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
		m.calculateLayoutHeights()
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

	// Handle template selector popup (only in template mode)
	// BUT: Always allow global quit commands and template messages to pass through
	if m.mode == ModeTemplate && m.templateSelector != nil && m.templateSelector.IsVisible() {
		// Check for quit commands first - these should always work
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			}
		}
		// Allow template workflow messages to pass through to main handler
		// These are returned by the selector when user submits, and need to be
		// handled by the main Update function, not by updateTemplateSelector
		switch msg.(type) {
		case TemplateRepoSelectedMsg, TemplateTreeLoadedMsg:
			// Don't route to updateTemplateSelector - let them be handled below
		default:
			// Route all other messages (keyboard input, etc) to the selector
			return m.updateTemplateSelector(msg)
		}
	}

	// Handle template conflict dialog
	if m.templateConflict.IsVisible() {
		return m.updateTemplateConflict(msg)
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
			// Only handle enter for sync in non-template modes
			// Template mode handles enter in updateTemplateMode()
			if m.mode != ModeTemplate && !m.syncing {
				return m.startSync()
			}
		}
	}

	// Handle mode switching before routing to mode-specific handlers
	// This ensures tab switching works from any mode, including template mode
	if switchMsg, ok := msg.(SwitchModeMsg); ok {
		m.mode = switchMsg.Mode
		m.list.SetLoading(true)

		// Handle owner switching based on mode
		switch switchMsg.Mode {
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

		case ModeTemplate:
			// Template mode shows the template workflow
			m.templateState.Reset()
			m.templateSelector.Reset()
			// Don't auto-show the template selector - user will press a key to open it
			// Load local repos for potential targets
			return m, tea.Batch(
				m.loadRepositories(),
				m.loadLocalReposForTemplateTargets(),
			)
		}

		return m, m.loadRepositories()
	}

	// Handle template workflow messages BEFORE routing to template mode
	// This ensures these critical messages are always handled regardless of UI state
	switch msg := msg.(type) {
	case TemplateTargetsLoadedMsg:
		m.localRepoPaths = msg.Paths
		m.templateTargets.SetRepos(msg.Paths)
		// Also set local templates for the selector
		m.templateSelector.SetLocalTemplates(msg.Paths)
		return m, nil

	case TemplateRepoSelectedMsg:
		return m.handleTemplateRepoSelected(msg)

	case TemplateTreeLoadedMsg:
		return m.handleTemplateTreeLoaded(msg)

	case TemplateTargetsSelectedMsg:
		return m.handleTemplateTargetsSelected(msg)

	case TemplateConflictResponseMsg:
		return m.handleTemplateConflictResponse(msg)

	case TemplateSyncProgressMsg:
		// Update progress display
		m.templateState.SyncProgress.Current = msg.Current
		m.templateState.SyncProgress.Total = msg.Total
		m.templateState.SyncProgress.CurrentFile = msg.CurrentFile
		m.templateState.SyncProgress.TargetRepo = msg.TargetRepo
		// Continue listening for more progress updates
		return m, m.waitForTemplateSyncProgress()

	case TemplateSyncCompleteMsg:
		m.templateSyncing = false
		m.templateState.SyncedCount = msg.Synced
		m.templateState.SkippedCount = msg.Skipped
		m.templateState.ErrorCount = msg.Errors
		m.templateState.Step = StepComplete
		// Clean up the progress channel
		m.templateSyncProgressChan = nil
		return m, nil
	}

	// Handle template mode after template workflow messages are processed
	if m.mode == ModeTemplate {
		return m.updateTemplateMode(msg)
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

// calculateLayoutHeights calculates the fixed heights of each layout component.
func (m *Model) calculateLayoutHeights() {
	// Recalculate on each call to handle dynamic elements

	// Header removed - no longer displayed
	m.headerHeight = 0

	// Tabs: 1 line content + 1 border bottom + 1 margin bottom = 3 lines
	m.tabsHeight = 3

	// Owner bar (only in GitHub modes): 1 line content + 1 border bottom + 1 margin bottom = 3 lines
	m.ownerBarHeight = 0
	if m.mode != ModeLocal {
		m.ownerBarHeight = 3
	}

	// Footer: 2 lines content + 2 padding (top/bottom) + 1 border top + 1 margin top = 6 lines
	m.footerHeight = 6

	// Progress bar (when visible): variable, estimate 8 lines
	progressHeight := 0
	if m.syncing || m.progress.IsComplete() {
		progressHeight = 8
	}

	// List gets remaining height with safety margin
	fixedHeight := m.headerHeight + m.tabsHeight + m.ownerBarHeight + m.footerHeight + progressHeight
	m.listHeight = m.height - fixedHeight - 2 // Extra 2 lines for safety
	if m.listHeight < 5 {
		m.listHeight = 5 // Minimum height for list
	}
}

// View renders the complete unified view - delegated to view.go.
func (m Model) View() string {
	if m.quitting {
		return RenderSuccess("Thanks for using reposync!\n")
	}

	// This will be implemented in view.go
	return m.renderView()
}

// updateTemplateMode handles updates when in template mode.
func (m Model) updateTemplateMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle global keys first
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "c":
			if !m.templateSyncing {
				m.showSettings = true
				return m, nil
			}

		case "esc":
			// If selector is visible, hide it
			if m.templateSelector.IsVisible() {
				m.templateSelector.Hide()
				return m, nil
			}
			// Otherwise, go back one step or reset
			if m.templateState.Step > StepSelectTemplate {
				m.templateState.PrevStep()
				return m, nil
			}
		}
	}

	// Handle based on current workflow step
	switch m.templateState.Step {
	case StepSelectTemplate:
		// Handle 's' or 'enter' key to show selector popup
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "s", "enter":
				if !m.templateSelector.IsVisible() {
					m.templateSelector.Show()
					return m, nil
				}
			}
		}
		// Selector is now handled as an overlay in the main Update function

	case StepBrowseTree:
		if m.templateTree != nil {
			// Handle enter to proceed to next step
			if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
				if m.templateTree.GetSelectedCount() > 0 {
					m.templateState.SelectedPaths = m.templateTree.GetSelectedPaths()
					m.templateState.Step = StepSelectTargets
					// Set exclude path for local templates
					if m.templateState.IsLocal {
						m.templateTargets.SetExcludePath(m.templateState.LocalTemplatePath)
					}
					return m, nil
				}
			}

			var cmd tea.Cmd
			m.templateTree, cmd = m.templateTree.Update(msg)
			cmds = append(cmds, cmd)
		}

	case StepSelectTargets:
		// Handle enter to start sync
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			if m.templateTargets != nil && m.templateTargets.HasSelections() {
				m.templateState.TargetRepos = m.templateTargets.GetSelectedPaths()
				return m.startTemplateSync()
			}
		}

		if m.templateTargets != nil {
			var cmd tea.Cmd
			m.templateTargets, cmd = m.templateTargets.Update(msg)
			cmds = append(cmds, cmd)
		}

	case StepSyncing:
		// Syncing in progress - no user interaction except viewing
		break

	case StepComplete:
		// Any key returns to template selector
		if _, ok := msg.(tea.KeyMsg); ok {
			m.templateState.Reset()
			m.templateSelector.Reset()
			return m, nil
		}
	}

	// Update tabs
	var tabCmd tea.Cmd
	m.tabs, tabCmd = m.tabs.Update(msg)
	cmds = append(cmds, tabCmd)

	return m, tea.Batch(cmds...)
}

// updateTemplateSelector handles updates when template selector popup is visible.
func (m Model) updateTemplateSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle ESC to close the selector
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.templateSelector.Hide()
		return m, nil
	}

	var cmd tea.Cmd
	m.templateSelector, cmd = m.templateSelector.Update(msg)
	return m, cmd
}

// updateTemplateConflict handles updates when template conflict dialog is visible.
func (m Model) updateTemplateConflict(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.templateConflict, cmd = m.templateConflict.Update(msg)
	return m, cmd
}

// handleTemplateRepoSelected handles when a template repository is selected.
func (m Model) handleTemplateRepoSelected(msg TemplateRepoSelectedMsg) (tea.Model, tea.Cmd) {
	m.templateSelector.SetLoading(true)
	// Don't hide selector yet - we'll hide it when the tree loads successfully

	if msg.IsLocal {
		// Local template - validate and load tree
		m.templateState.SetLocalTemplate(msg.LocalPath)
		return m, m.loadLocalTemplateTree(msg.LocalPath)
	}

	// GitHub template - fetch default branch and tree
	m.templateState.TemplateOwner = msg.Owner
	m.templateState.TemplateRepo = msg.Repo
	return m, m.loadGitHubTemplateTree(msg.Owner, msg.Repo)
}

// loadGitHubTemplateTree loads the tree for a GitHub template repository.
func (m *Model) loadGitHubTemplateTree(owner, repo string) tea.Cmd {
	return func() tea.Msg {
		// Get default branch
		branch, err := m.githubClient.GetDefaultBranch(owner, repo)
		if err != nil {
			return TemplateTreeLoadedMsg{Err: fmt.Errorf("failed to get default branch: %w", err)}
		}

		// Get tree
		treeResp, err := m.githubClient.GetRepoTree(owner, repo, branch)
		if err != nil {
			return TemplateTreeLoadedMsg{Err: fmt.Errorf("failed to get repository tree: %w", err)}
		}

		// Build tree model
		return TemplateTreeLoadedMsg{
			Root: buildTemplateTreeFromGitHub(treeResp, branch),
			Err:  nil,
		}
	}
}

// loadLocalTemplateTree loads the tree for a local template directory.
func (m *Model) loadLocalTemplateTree(localPath string) tea.Cmd {
	return func() tea.Msg {
		root, err := buildLocalTemplateTree(localPath)
		if err != nil {
			return TemplateTreeLoadedMsg{Err: err}
		}
		return TemplateTreeLoadedMsg{Root: root, Err: nil}
	}
}

// buildTemplateTreeFromGitHub converts a GitHub tree response to TemplateTreeNode.
func buildTemplateTreeFromGitHub(resp *github.TreeResponse, branch string) *TemplateTreeNode {
	// This is handled by NewTemplateTreeModel, just pass through data
	// We'll set the branch on the state after this
	return &TemplateTreeNode{
		Path:     "",
		Name:     "/",
		IsDir:    true,
		Expanded: true,
		Selected: false,
		Children: nil, // Will be built by NewTemplateTreeModel
	}
}

// buildLocalTemplateTree builds a tree from a local directory.
func buildLocalTemplateTree(rootPath string) (*TemplateTreeNode, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", rootPath)
	}

	root := &TemplateTreeNode{
		Path:     "",
		Name:     filepath.Base(rootPath),
		IsDir:    true,
		Expanded: true,
		Selected: false,
		Children: make([]*TemplateTreeNode, 0),
	}

	err = buildLocalTreeRecursive(rootPath, "", root)
	if err != nil {
		return nil, err
	}

	return root, nil
}

// buildLocalTreeRecursive recursively builds the tree from local filesystem.
func buildLocalTreeRecursive(basePath, relativePath string, parent *TemplateTreeNode) error {
	fullPath := filepath.Join(basePath, relativePath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Skip .git directory
		if entry.Name() == ".git" {
			continue
		}

		childPath := filepath.Join(relativePath, entry.Name())
		if relativePath == "" {
			childPath = entry.Name()
		}

		child := &TemplateTreeNode{
			Path:     childPath,
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			Expanded: false,
			Selected: false,
			Children: make([]*TemplateTreeNode, 0),
		}

		if entry.IsDir() {
			if err := buildLocalTreeRecursive(basePath, childPath, child); err != nil {
				return err
			}
		} else {
			info, _ := entry.Info()
			if info != nil {
				child.Size = info.Size()
			}
		}

		parent.Children = append(parent.Children, child)
	}

	return nil
}

// handleTemplateTreeLoaded handles when the template tree is loaded.
func (m Model) handleTemplateTreeLoaded(msg TemplateTreeLoadedMsg) (tea.Model, tea.Cmd) {
	m.templateSelector.SetLoading(false)

	if msg.Err != nil {
		m.templateSelector.SetError(msg.Err)
		// Keep selector visible to show error
		return m, nil
	}

	if msg.Root == nil {
		m.templateSelector.SetError(fmt.Errorf("failed to load template tree: root is nil"))
		// Keep selector visible to show error
		return m, nil
	}

	// Hide the selector popup now that we successfully loaded the tree
	m.templateSelector.Hide()

	// Create tree model based on source type
	if m.templateState.IsLocal {
		m.templateTree = NewTemplateTreeModelFromLocal(msg.Root, m.templateState.LocalTemplatePath)
	} else {
		// For GitHub, we need to load the tree properly
		branch, err := m.githubClient.GetDefaultBranch(m.templateState.TemplateOwner, m.templateState.TemplateRepo)
		if err != nil {
			m.templateSelector.SetError(fmt.Errorf("failed to get default branch: %w", err))
			return m, nil
		}
		m.templateState.TemplateBranch = branch

		treeResp, err := m.githubClient.GetRepoTree(m.templateState.TemplateOwner, m.templateState.TemplateRepo, branch)
		if err != nil {
			m.templateSelector.SetError(err)
			return m, nil
		}

		templateName := m.templateState.TemplateOwner + "/" + m.templateState.TemplateRepo
		m.templateTree = NewTemplateTreeModel(treeResp, templateName, branch)
	}

	// Safely set tree size
	treeWidth := m.width
	if treeWidth < 40 {
		treeWidth = 80 // Default width
	}
	// Match the calculation in renderTemplateTree()
	// Main UI chrome: tabs (3) + footer (6) = 9 lines
	mainChrome := 9
	treeHeight := m.height - mainChrome
	if treeHeight < 10 {
		treeHeight = 20 // Default height
	}
	m.templateTree.SetSize(treeWidth, treeHeight)
	m.templateState.Step = StepBrowseTree

	// Save to recent templates
	m.saveRecentTemplate()

	return m, nil
}

// handleTemplateTargetsSelected handles when target repositories are selected.
func (m Model) handleTemplateTargetsSelected(msg TemplateTargetsSelectedMsg) (tea.Model, tea.Cmd) {
	m.templateState.TargetRepos = msg.TargetPaths
	return m.startTemplateSync()
}

// handleTemplateConflictResponse handles user response to a conflict prompt.
func (m Model) handleTemplateConflictResponse(msg TemplateConflictResponseMsg) (tea.Model, tea.Cmd) {
	if m.templateEngine == nil {
		return m, nil
	}

	switch msg.Action {
	case ConflictOverwriteAll:
		m.templateEngine.SetOverwriteAll(true)
	case ConflictSkipAll:
		m.templateEngine.SetSkipAll(true)
	}

	// Continue syncing
	return m, nil
}

// startTemplateSync starts the template synchronization process.
func (m Model) startTemplateSync() (tea.Model, tea.Cmd) {
	if len(m.templateState.SelectedPaths) == 0 || len(m.templateState.TargetRepos) == 0 {
		return m, nil
	}

	m.templateSyncing = true
	m.templateState.Step = StepSyncing

	// Create sync engine
	if m.templateState.IsLocal {
		m.templateEngine = template.NewLocalSyncEngine(m.templateState.LocalTemplatePath)
	} else {
		m.templateEngine = template.NewSyncEngine(
			m.githubClient,
			m.templateState.TemplateOwner,
			m.templateState.TemplateRepo,
			m.templateState.TemplateBranch,
		)
	}

	// Start sync
	return m, m.runTemplateSync()
}

// runTemplateSync executes the template sync operation.
// This uses a subscription-like pattern where progress updates are sent
// through a channel and converted into Bubbletea messages.
func (m *Model) runTemplateSync() tea.Cmd {
	return tea.Batch(
		m.executeTemplateSync(),
		m.waitForTemplateSyncProgress(),
	)
}

// executeTemplateSync runs the sync in a goroutine and sends progress to a shared channel.
func (m *Model) executeTemplateSync() tea.Cmd {
	return func() tea.Msg {
		go func() {
			results := m.templateEngine.SyncFiles(
				m.templateState.SelectedPaths,
				m.templateState.TargetRepos,
				func(progress template.SyncProgress) {
					// Send progress update through the program
					if m.templateSyncProgressChan != nil {
						m.templateSyncProgressChan <- TemplateSyncProgressMsg{
							Current:     progress.Current,
							Total:       progress.Total,
							CurrentFile: progress.CurrentFile,
							TargetRepo:  progress.TargetRepo,
						}
					}
				},
				func(conflict template.ConflictInfo) template.ConflictAction {
					// For now, use batch flags or skip
					if m.templateEngine.ShouldOverwriteAll() {
						return template.ActionOverwrite
					}
					if m.templateEngine.ShouldSkipAll() {
						return template.ActionSkip
					}
					// Default: skip (in a real impl, would show dialog)
					return template.ActionSkip
				},
			)

			synced, skipped, errors := template.GetSyncSummary(results)

			// Send completion message
			if m.templateSyncProgressChan != nil {
				m.templateSyncProgressChan <- TemplateSyncCompleteMsg{
					Synced:  synced,
					Skipped: skipped,
					Errors:  errors,
				}
				close(m.templateSyncProgressChan)
			}
		}()
		return nil // Return immediately, goroutine will send messages
	}
}

// waitForTemplateSyncProgress waits for messages from the sync goroutine.
func (m *Model) waitForTemplateSyncProgress() tea.Cmd {
	// Initialize the channel if needed
	if m.templateSyncProgressChan == nil {
		m.templateSyncProgressChan = make(chan tea.Msg, 100)
	}

	return func() tea.Msg {
		msg, ok := <-m.templateSyncProgressChan
		if !ok {
			// Channel closed, sync is done
			return nil
		}
		return msg
	}
}

// saveRecentTemplate saves the current template to recent templates.
func (m *Model) saveRecentTemplate() {
	var templateName string
	if m.templateState.IsLocal {
		templateName = m.templateState.LocalTemplatePath
	} else {
		templateName = m.templateState.TemplateOwner + "/" + m.templateState.TemplateRepo
	}

	if templateName == "" {
		return
	}

	// Load current config
	persistedCfg, err := m.store.Load()
	if err != nil {
		persistedCfg = &config.PersistedConfig{}
	}

	// Add to recent (dedupe and limit)
	recent := []string{templateName}
	for _, t := range persistedCfg.RecentTemplates {
		if t != templateName && len(recent) < 10 {
			recent = append(recent, t)
		}
	}
	persistedCfg.RecentTemplates = recent

	// Save
	_ = m.store.Save(persistedCfg)

	// Update selector
	m.templateSelector.SetRecentTemplates(recent)
}
