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
		meta["archived"] = "📦 Archived"
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
	selected int
	pageSize int

	// Enum (platform-dependent)
	sortMode SortMode

	// Bools (1 byte each)
	searching   bool
	compactMode bool
	loading     bool
}

// NewListModel creates a new list model.
func NewListModel() *ListModel {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 50

	return &ListModel{
		items:       []ListItem{},
		filtered:    []ListItem{},
		selected:    0,
		checked:     make(map[string]bool),
		searchInput: ti,
		searching:   false,
		sortMode:    SortByName,
		compactMode: false,
		pageSize:    12,
		loading:     false,
	}
}

// SetCompactMode sets the compact mode for the list.
func (m *ListModel) SetCompactMode(compact bool) {
	m.compactMode = compact
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
	}

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

// View renders the list.
func (m *ListModel) View(width, height int) string {
	var b strings.Builder

	if m.loading {
		b.WriteString(RenderInfo("Loading..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(RenderError(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	if len(m.items) == 0 {
		b.WriteString(RenderWarning("No items found"))
		return b.String()
	}

	// Search input
	if m.searching {
		b.WriteString(RenderSearchPrompt(m.searchInput.View()))
		b.WriteString("\n\n")
	}

	// Calculate viewport
	displayCount := m.pageSize
	if height > 20 {
		displayCount = height - 15
	}

	start := m.selected - displayCount/2
	if start < 0 {
		start = 0
	}
	end := start + displayCount
	if end > len(m.filtered) {
		end = len(m.filtered)
		start = end - displayCount
		if start < 0 {
			start = 0
		}
	}

	// Count active vs archived repos for section header
	activeCount := 0
	archivedCount := 0
	for _, item := range m.filtered {
		if item.IsArchived() {
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
			b.WriteString("\n")
			b.WriteString(RenderSectionHeader(fmt.Sprintf("Archived (%d)", archivedCount)))
			b.WriteString("\n\n")
			archivedHeaderRendered = true
		}

		// Main line
		title := item.Title()
		if item.Description() != "" {
			maxDescLen := 60
			if width > 120 {
				maxDescLen = 80
			}
			title += " - " + truncate(item.Description(), maxDescLen)
		}

		// Use appropriate renderer based on archived status
		if isArchived {
			b.WriteString(RenderArchivedListItem(title, isSelected, isChecked))
		} else {
			b.WriteString(RenderListItem(title, isSelected, isChecked))
		}
		b.WriteString("\n")

		// Metadata line for selected item
		if isSelected && !m.compactMode {
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
					b.WriteString(RenderMetadata(metadataLine))
					b.WriteString("\n")
				}
			}
		}
	}

	// Show navigation hint if there are more items
	if len(m.filtered) > displayCount {
		b.WriteString("\n")
		navHint := fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.filtered))
		b.WriteString(RenderMetadata(navHint))
	}

	return b.String()
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
