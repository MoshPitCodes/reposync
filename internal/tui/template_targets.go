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
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TemplateTargetRepo represents a local repository that can be a sync target.
type TemplateTargetRepo struct {
	Path       string
	Name       string
	IsSelected bool
	IsDisabled bool // True if this is the template source (for local templates)
}

// TemplateTargetsModel manages the target repository multi-select.
type TemplateTargetsModel struct {
	// List of target repositories
	repos []TemplateTargetRepo

	// Current cursor position
	cursor int

	// Viewport offset for scrolling
	viewportOffset int

	// Dimensions
	width  int
	height int

	// Filter
	filter string

	// Path to exclude (the template path for local templates)
	excludePath string
}

// NewTemplateTargetsModel creates a new target selector model.
func NewTemplateTargetsModel() *TemplateTargetsModel {
	return &TemplateTargetsModel{
		repos:          make([]TemplateTargetRepo, 0),
		cursor:         0,
		viewportOffset: 0,
		width:          60,
		height:         20,
		filter:         "",
		excludePath:    "",
	}
}

// SetRepos sets the list of local repositories as potential targets.
func (m *TemplateTargetsModel) SetRepos(paths []string) {
	m.repos = make([]TemplateTargetRepo, len(paths))
	for i, path := range paths {
		m.repos[i] = TemplateTargetRepo{
			Path:       path,
			Name:       filepath.Base(path),
			IsSelected: false,
			IsDisabled: m.excludePath != "" && normalizePath(path) == normalizePath(m.excludePath),
		}
	}
}

// SetExcludePath sets the path to exclude from selection (template source).
func (m *TemplateTargetsModel) SetExcludePath(path string) {
	m.excludePath = path
	// Update disabled state for existing repos
	for i := range m.repos {
		m.repos[i].IsDisabled = path != "" && normalizePath(m.repos[i].Path) == normalizePath(path)
		// Deselect if now disabled
		if m.repos[i].IsDisabled {
			m.repos[i].IsSelected = false
		}
	}
}

// SetSize sets the dimensions of the selector.
func (m *TemplateTargetsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Reset clears all selections.
func (m *TemplateTargetsModel) Reset() {
	for i := range m.repos {
		m.repos[i].IsSelected = false
	}
	m.cursor = 0
	m.viewportOffset = 0
	m.filter = ""
}

// getFilteredRepos returns repos matching the current filter.
func (m *TemplateTargetsModel) getFilteredRepos() []int {
	if m.filter == "" {
		indices := make([]int, len(m.repos))
		for i := range m.repos {
			indices[i] = i
		}
		return indices
	}

	filterLower := strings.ToLower(m.filter)
	indices := make([]int, 0)
	for i, repo := range m.repos {
		if strings.Contains(strings.ToLower(repo.Name), filterLower) ||
			strings.Contains(strings.ToLower(repo.Path), filterLower) {
			indices = append(indices, i)
		}
	}
	return indices
}

// Update handles messages for the target selector.
func (m *TemplateTargetsModel) Update(msg tea.Msg) (*TemplateTargetsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			filtered := m.getFilteredRepos()
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible(filtered)
			}
			return m, nil

		case "down", "j":
			filtered := m.getFilteredRepos()
			if m.cursor < len(filtered)-1 {
				m.cursor++
				m.ensureVisible(filtered)
			}
			return m, nil

		case " ":
			// Toggle selection
			filtered := m.getFilteredRepos()
			if m.cursor >= 0 && m.cursor < len(filtered) {
				idx := filtered[m.cursor]
				if !m.repos[idx].IsDisabled {
					m.repos[idx].IsSelected = !m.repos[idx].IsSelected
				}
			}
			return m, nil

		case "a":
			// Select all (non-disabled)
			for i := range m.repos {
				if !m.repos[i].IsDisabled {
					m.repos[i].IsSelected = true
				}
			}
			return m, nil

		case "n":
			// Deselect all
			for i := range m.repos {
				m.repos[i].IsSelected = false
			}
			return m, nil

		case "backspace":
			// Remove last character from filter
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.cursor = 0
				m.viewportOffset = 0
			}
			return m, nil

		case "esc":
			// Clear filter
			if m.filter != "" {
				m.filter = ""
				m.cursor = 0
				m.viewportOffset = 0
				return m, nil
			}
			return m, nil

		default:
			// Add to filter if printable
			if msg.Type == tea.KeyRunes {
				m.filter += string(msg.Runes)
				m.cursor = 0
				m.viewportOffset = 0
			}
			return m, nil
		}
	}

	return m, nil
}

// ensureVisible adjusts viewport to keep cursor visible.
func (m *TemplateTargetsModel) ensureVisible(filtered []int) {
	visibleLines := m.height - 10
	if visibleLines < 1 {
		visibleLines = 5
	}

	if m.cursor < m.viewportOffset {
		m.viewportOffset = m.cursor
	} else if m.cursor >= m.viewportOffset+visibleLines {
		m.viewportOffset = m.cursor - visibleLines + 1
	}
}

// GetSelectedPaths returns the paths of all selected repositories.
func (m *TemplateTargetsModel) GetSelectedPaths() []string {
	paths := make([]string, 0)
	for _, repo := range m.repos {
		if repo.IsSelected && !repo.IsDisabled {
			paths = append(paths, repo.Path)
		}
	}
	return paths
}

// GetSelectedCount returns the count of selected repositories.
func (m *TemplateTargetsModel) GetSelectedCount() int {
	return len(m.GetSelectedPaths())
}

// HasSelections returns true if at least one repo is selected.
func (m *TemplateTargetsModel) HasSelections() bool {
	return m.GetSelectedCount() > 0
}

// View renders the target selector.
func (m *TemplateTargetsModel) View() string {
	var b strings.Builder

	// Header
	header := templateTargetsHeaderStyle.Render("📁 Select Target Repositories")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Selection count
	selectedCount := m.GetSelectedCount()
	totalRepos := len(m.repos)
	countStr := fmt.Sprintf("Selected: %d/%d repositories", selectedCount, totalRepos)
	b.WriteString(templateTargetsCountStyle.Render(countStr))
	b.WriteString("\n")

	// Filter display
	if m.filter != "" {
		filterStr := fmt.Sprintf("Filter: %s", m.filter)
		b.WriteString(templateTargetsFilterStyle.Render(filterStr))
	} else {
		b.WriteString(templateTargetsHintStyle.Render("Type to filter..."))
	}
	b.WriteString("\n\n")

	// Repository list
	filtered := m.getFilteredRepos()
	visibleLines := m.height - 12
	if visibleLines < 1 {
		visibleLines = 5
	}

	if len(filtered) == 0 {
		if m.filter != "" {
			b.WriteString(templateTargetsHintStyle.Render("No repositories match the filter"))
		} else {
			b.WriteString(templateTargetsHintStyle.Render("No local repositories available"))
		}
		b.WriteString("\n")
	} else {
		startIdx := m.viewportOffset
		endIdx := startIdx + visibleLines
		if endIdx > len(filtered) {
			endIdx = len(filtered)
		}

		for i := startIdx; i < endIdx; i++ {
			repoIdx := filtered[i]
			repo := m.repos[repoIdx]

			// Checkbox
			checkbox := "[ ]"
			if repo.IsSelected {
				checkbox = "[✓]"
			}
			if repo.IsDisabled {
				checkbox = "[×]"
			}

			// Build line
			line := fmt.Sprintf("%s 📁 %s", checkbox, repo.Name)

			// Show path hint on cursor
			if i == m.cursor {
				line = fmt.Sprintf("%s\n      %s", line, repo.Path)
			}

			// Apply style
			var style lipgloss.Style
			if repo.IsDisabled {
				style = templateTargetsDisabledStyle
			} else if i == m.cursor {
				style = templateTargetsSelectedStyle
			} else if repo.IsSelected {
				style = templateTargetsCheckedStyle
			} else {
				style = templateTargetsItemStyle
			}

			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}

		// Scroll indicator
		if len(filtered) > visibleLines {
			scrollInfo := fmt.Sprintf("(%d-%d of %d)", startIdx+1, endIdx, len(filtered))
			b.WriteString(templateTargetsHintStyle.Render(scrollInfo))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Warning for disabled items
	if m.excludePath != "" {
		warning := "Note: Template source repository cannot be selected as target"
		b.WriteString(templateTargetsWarningStyle.Render(warning))
		b.WriteString("\n")
	}

	// Help text
	helpText := "↑/↓ navigate • space toggle • a all • n none • type to filter"
	b.WriteString(templateTargetsHelpStyle.Render(helpText))

	return templateTargetsStyle.Width(m.width).Render(b.String())
}

// Styles for target selector
var (
	templateTargetsStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor)

	templateTargetsHeaderStyle = lipgloss.NewStyle().
					Foreground(primaryColor).
					Bold(true)

	templateTargetsCountStyle = lipgloss.NewStyle().
					Foreground(secondaryColor).
					Bold(true)

	templateTargetsFilterStyle = lipgloss.NewStyle().
					Foreground(accentColor).
					Bold(true)

	templateTargetsItemStyle = lipgloss.NewStyle().
				Foreground(fgColor)

	templateTargetsSelectedStyle = lipgloss.NewStyle().
					Foreground(secondaryColor).
					Bold(true)

	templateTargetsCheckedStyle = lipgloss.NewStyle().
					Foreground(successColor)

	templateTargetsDisabledStyle = lipgloss.NewStyle().
					Foreground(mutedColor).
					Italic(true)

	templateTargetsHintStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

	templateTargetsWarningStyle = lipgloss.NewStyle().
					Foreground(warningColor).
					Italic(true)

	templateTargetsHelpStyle = lipgloss.NewStyle().
				Foreground(mutedColor)
)
