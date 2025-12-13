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
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/MoshPitCodes/reposync/internal/github"
	"github.com/MoshPitCodes/reposync/internal/local"
)

// ListItem is the generic interface for items in the list.
type ListItem interface {
	ID() string
	Title() string
	Description() string
	Metadata() map[string]string
	IsArchived() bool
}

// GitHubRepoItem wraps a GitHub repository as a ListItem.
type GitHubRepoItem struct {
	repo github.Repository
}

// ID returns the unique identifier for the item.
func (i GitHubRepoItem) ID() string {
	return i.repo.FullName
}

// Title returns the display title.
func (i GitHubRepoItem) Title() string {
	return i.repo.Name
}

// Description returns the description.
func (i GitHubRepoItem) Description() string {
	return i.repo.Description
}

// Metadata returns metadata about the repository.
func (i GitHubRepoItem) Metadata() map[string]string {
	meta := make(map[string]string)

	if i.repo.Language != "" {
		meta["language"] = i.repo.Language
	}
	if i.repo.Stars > 0 {
		meta["stars"] = fmt.Sprintf("⭐ %d", i.repo.Stars)
	}
	if i.repo.IsPrivate {
		meta["visibility"] = "🔒 Private"
	} else {
		meta["visibility"] = "🌐 Public"
	}
	if i.repo.IsArchived {
		meta["archived"] = "Archived"
	}
	meta["clone_url"] = i.repo.CloneURL

	return meta
}

// IsArchived returns true if the repository is archived.
func (i GitHubRepoItem) IsArchived() bool {
	return i.repo.IsArchived
}

// LocalRepoItem wraps a local repository as a ListItem.
type LocalRepoItem struct {
	repo local.Repository
}

// ID returns the unique identifier for the item.
func (i LocalRepoItem) ID() string {
	return i.repo.Path
}

// Title returns the display title.
func (i LocalRepoItem) Title() string {
	return i.repo.Name
}

// Description returns the description.
func (i LocalRepoItem) Description() string {
	return "" // Local repos don't have descriptions
}

// Metadata returns metadata about the repository.
func (i LocalRepoItem) Metadata() map[string]string {
	meta := make(map[string]string)

	meta["path"] = i.repo.Path
	meta["size"] = local.FormatSize(i.repo.Size)

	if i.repo.Branch != "" {
		meta["branch"] = i.repo.Branch
	}
	if i.repo.IsGitRepo {
		meta["type"] = "📦 Git Repository"
	}

	return meta
}

// IsArchived returns false for local repositories (they cannot be archived).
func (i LocalRepoItem) IsArchived() bool {
	return false
}

// SortMode represents different ways to sort repositories.
type SortMode int

const (
	SortByName SortMode = iota
	SortByStars
	SortByUpdated
)

// String returns the string representation of the sort mode.
func (s SortMode) String() string {
	switch s {
	case SortByName:
		return "Name"
	case SortByStars:
		return "Stars"
	case SortByUpdated:
		return "Updated"
	default:
		return "Name"
	}
}

// ListModel manages a generic list of items.
type ListModel struct {
	// Complex type
	searchInput textinput.Model

	// Interface/error (16 bytes)
	err error

	// Slices (24 bytes each)
	items    []ListItem
	filtered []ListItem

	// Map (8 bytes pointer)
	checked map[string]bool

	// Ints (8 bytes each)
	selected       int
	pageSize       int
	viewportOffset int

	// Enum (platform-dependent)
	sortMode SortMode

	// Bools (1 byte each)
	searching bool
	loading   bool
}

// NewListModel creates a new list model.
func NewListModel() *ListModel {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 50

	return &ListModel{
		items:          []ListItem{},
		filtered:       []ListItem{},
		selected:       0,
		checked:        make(map[string]bool),
		searchInput:    ti,
		searching:      false,
		sortMode:       SortByName,
		pageSize:       12,
		loading:        false,
		viewportOffset: 0,
	}
}

// SetItems sets the items for the list.
func (m *ListModel) SetItems(items []ListItem) {
	m.items = items
	m.sortItems()
	m.filtered = m.items
	m.selected = 0
}

// SetLoading sets the loading state.
func (m *ListModel) SetLoading(loading bool) {
	m.loading = loading
}

// SetError sets the error state.
func (m *ListModel) SetError(err error) {
	m.err = err
	m.loading = false
}

// GetSelectedItems returns the list of selected item IDs.
func (m *ListModel) GetSelectedItems() []string {
	var selected []string
	for id, checked := range m.checked {
		if checked {
			selected = append(selected, id)
		}
	}
	return selected
}

// GetSelectedCount returns the number of selected items.
func (m *ListModel) GetSelectedCount() int {
	count := 0
	for _, checked := range m.checked {
		if checked {
			count++
		}
	}
	return count
}

// Update handles messages for the list.
func (m *ListModel) Update(msg tea.Msg) (*ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			return m.handleSearchInput(msg)
		}
		return m.handleNavigation(msg)
	}

	return m, nil
}

// handleSearchInput processes input when searching.
func (m *ListModel) handleSearchInput(msg tea.KeyMsg) (*ListModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.searching = false
		m.searchInput.SetValue("")
		m.filtered = m.items
		m.selected = 0
		return m, nil

	case "enter":
		m.searching = false
		return m, nil

	default:
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.filterItems()
		return m, cmd
	}
}

// handleNavigation processes navigation and selection keys.
func (m *ListModel) handleNavigation(msg tea.KeyMsg) (*ListModel, tea.Cmd) {
	oldSelected := m.selected

	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}

	case "down", "j":
		if m.selected < len(m.filtered)-1 {
			m.selected++
		}

	case "pgup":
		m.selected -= m.pageSize
		if m.selected < 0 {
			m.selected = 0
		}

	case "pgdown":
		m.selected += m.pageSize
		if m.selected >= len(m.filtered) {
			m.selected = len(m.filtered) - 1
		}

	case " ":
		// Toggle selection
		if m.selected >= 0 && m.selected < len(m.filtered) {
			id := m.filtered[m.selected].ID()
			m.checked[id] = !m.checked[id]
		}

	case "/":
		m.searching = true
		m.searchInput.Focus()
		return m, textinput.Blink

	case "a":
		// Select all
		for _, item := range m.filtered {
			m.checked[item.ID()] = true
		}

	case "n":
		// Deselect all
		m.checked = make(map[string]bool)

	case "s":
		// Cycle sort mode
		m.sortMode = (m.sortMode + 1) % 3
		m.sortItems()
		m.filterItems()
		m.selected = 0
		m.viewportOffset = 0
	}

	// Viewport will be updated in View() based on selection
	// No need to update it here since we don't know display count yet
	_ = oldSelected

	return m, nil
}

// sortItems sorts items based on current sort mode, with archived repos at the end.
func (m *ListModel) sortItems() {
	// Separate active and archived items
	var active, archived []ListItem
	for _, item := range m.items {
		if item.IsArchived() {
			archived = append(archived, item)
		} else {
			active = append(active, item)
		}
	}

	// Sort function based on current mode
	sortFn := func(items []ListItem) {
		switch m.sortMode {
		case SortByName:
			sort.Slice(items, func(i, j int) bool {
				return strings.ToLower(items[i].Title()) < strings.ToLower(items[j].Title())
			})

		case SortByStars:
			sort.Slice(items, func(i, j int) bool {
				starsI := items[i].Metadata()["stars"]
				starsJ := items[j].Metadata()["stars"]
				return starsJ < starsI
			})

		case SortByUpdated:
			sort.Slice(items, func(i, j int) bool {
				return strings.ToLower(items[i].Title()) < strings.ToLower(items[j].Title())
			})
		}
	}

	// Sort each group independently
	sortFn(active)
	sortFn(archived)

	// Recombine: active first, then archived
	m.items = append(active, archived...)
}

// filterItems filters items based on search input.
func (m *ListModel) filterItems() {
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		m.filtered = m.items
		return
	}

	m.filtered = []ListItem{}
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.Title()), query) ||
			strings.Contains(strings.ToLower(item.Description()), query) {
			m.filtered = append(m.filtered, item)
		}
	}

	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

// View renders the list with a fixed height and viewport scrolling.
func (m *ListModel) View(width, height int) string {
	var lines []string

	if m.loading {
		lines = append(lines, RenderInfo("Loading..."))
		return m.renderWithFixedHeight(lines, height)
	}

	if m.err != nil {
		lines = append(lines, RenderError(fmt.Sprintf("Error: %v", m.err)))
		return m.renderWithFixedHeight(lines, height)
	}

	if len(m.items) == 0 {
		lines = append(lines, RenderWarning("No items found"))
		return m.renderWithFixedHeight(lines, height)
	}

	// Search input
	searchHeight := 0
	if m.searching {
		lines = append(lines, RenderSearchPrompt(m.searchInput.View()))
		lines = append(lines, "")
		searchHeight = 2
	}

	// Calculate available height for items (reserve 2 lines for nav hint)
	availableHeight := height - searchHeight - 2
	if availableHeight < 3 {
		availableHeight = 3
	}

	// Calculate lines per item (1 for item, 1 for description if exists, 1 for metadata for selected item)
	// Account for potential description + metadata lines for selected item
	linesPerItem := 3

	// Calculate how many items can fit
	displayCount := availableHeight / linesPerItem
	if displayCount < 1 {
		displayCount = 1
	}

	// Update viewport offset to keep selected item visible
	m.updateViewportForDisplay(displayCount)

	// Calculate viewport window
	start := m.viewportOffset
	end := start + displayCount
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	// Count active vs archived repos for section header
	activeCount := 0
	archivedCount := 0
	firstArchivedIndex := -1
	for i, item := range m.filtered {
		if item.IsArchived() {
			if firstArchivedIndex == -1 {
				firstArchivedIndex = i
			}
			archivedCount++
		} else {
			activeCount++
		}
	}

	// Render items
	archivedHeaderRendered := false
	for i := start; i < end; i++ {
		item := m.filtered[i]
		isSelected := i == m.selected
		isChecked := m.checked[item.ID()]
		isArchived := item.IsArchived()

		// Render archived section header when we reach first archived item in viewport
		if isArchived && !archivedHeaderRendered && archivedCount > 0 && activeCount > 0 {
			// Only show header if this is the first archived item OR if we're at viewport start
			if i == firstArchivedIndex || (i == start && i > 0) {
				lines = append(lines, "")
				lines = append(lines, RenderSectionHeader(fmt.Sprintf("Archived (%d)", archivedCount)))
				lines = append(lines, "")
				archivedHeaderRendered = true
			}
		}

		// Main line - just the title, no description
		title := item.Title()

		// Use appropriate renderer based on archived status
		if isArchived {
			lines = append(lines, RenderArchivedListItem(title, isSelected, isChecked))
		} else {
			lines = append(lines, RenderListItem(title, isSelected, isChecked))
		}

		// Description line for selected item (only if it exists)
		if isSelected && item.Description() != "" {
			maxDescLen := 80
			if width < 120 {
				maxDescLen = 60
			}
			descLine := "    " + truncate(item.Description(), maxDescLen)
			lines = append(lines, RenderMetadata(descLine))
		}

		// Metadata line for selected item
		if isSelected {
			meta := item.Metadata()
			if len(meta) > 0 {
				var metaParts []string
				// Order: language, stars, visibility, archived, branch, size, type, path
				if v, ok := meta["language"]; ok {
					metaParts = append(metaParts, v)
				}
				if v, ok := meta["stars"]; ok {
					metaParts = append(metaParts, v)
				}
				if v, ok := meta["visibility"]; ok {
					metaParts = append(metaParts, v)
				}
				if v, ok := meta["archived"]; ok {
					metaParts = append(metaParts, v)
				}
				if v, ok := meta["branch"]; ok {
					metaParts = append(metaParts, fmt.Sprintf("Branch: %s", v))
				}
				if v, ok := meta["size"]; ok {
					metaParts = append(metaParts, fmt.Sprintf("Size: %s", v))
				}
				if v, ok := meta["type"]; ok {
					metaParts = append(metaParts, v)
				}
				if v, ok := meta["path"]; ok {
					metaParts = append(metaParts, truncate(v, 60))
				}

				if len(metaParts) > 0 {
					metadataLine := "    " + strings.Join(metaParts, " • ")
					lines = append(lines, RenderMetadata(metadataLine))
				}
			}
		}
	}

	// Show navigation hint if there are more items
	if len(m.filtered) > 0 {
		lines = append(lines, "")
		navHint := fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.filtered))
		if m.selected >= 0 && m.selected < len(m.filtered) {
			navHint += fmt.Sprintf(" (selected: %d)", m.selected+1)
		}
		lines = append(lines, RenderMetadata(navHint))
	}

	return m.renderWithFixedHeight(lines, height)
}

// updateViewportForDisplay updates the viewport offset to keep selected item visible.
func (m *ListModel) updateViewportForDisplay(displayCount int) {
	if m.selected < m.viewportOffset {
		// Selected item is above viewport, scroll up
		m.viewportOffset = m.selected
	} else if m.selected >= m.viewportOffset+displayCount {
		// Selected item is below viewport, scroll down
		m.viewportOffset = m.selected - displayCount + 1
	}

	// Ensure viewport doesn't go past the end
	maxOffset := len(m.filtered) - displayCount
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.viewportOffset > maxOffset {
		m.viewportOffset = maxOffset
	}

	// Ensure viewport doesn't go negative
	if m.viewportOffset < 0 {
		m.viewportOffset = 0
	}
}

// renderWithFixedHeight renders content, truncating if needed but NOT padding.
// Padding was causing total view height to exceed terminal bounds.
func (m *ListModel) renderWithFixedHeight(lines []string, height int) string {
	// If content is taller than height, truncate
	if len(lines) > height && height > 0 {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// FromGitHubRepos converts GitHub repositories to ListItems.
func FromGitHubRepos(repos []github.Repository) []ListItem {
	items := make([]ListItem, len(repos))
	for i, repo := range repos {
		items[i] = GitHubRepoItem{repo: repo}
	}
	return items
}

// FromLocalRepos converts local repositories to ListItems.
func FromLocalRepos(repos []local.Repository) []ListItem {
	items := make([]ListItem, len(repos))
	for i, repo := range repos {
		items[i] = LocalRepoItem{repo: repo}
	}
	return items
}
