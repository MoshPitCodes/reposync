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
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TemplateSourceType represents the type of template source.
type TemplateSourceType int

const (
	// TemplateSourceGitHub represents a GitHub repository template.
	TemplateSourceGitHub TemplateSourceType = iota
	// TemplateSourceLocal represents a local directory template.
	TemplateSourceLocal
)

// TemplateSelectorModel manages the template repository selector.
type TemplateSelectorModel struct {
	// Text input for owner/repo or local path
	input textinput.Model

	// Recent templates list (owner/repo for GitHub, paths for local)
	recentTemplates []string

	// Local template directories
	localTemplates []string

	// Current source type
	sourceType TemplateSourceType

	// Cursor position in recent list
	cursor int

	// Dimensions
	width  int
	height int

	// State
	visible bool
	loading bool
	err     error
}

// NewTemplateSelectorModel creates a new template selector model.
func NewTemplateSelectorModel(recentTemplates []string) *TemplateSelectorModel {
	ti := textinput.New()
	ti.Placeholder = "owner/repo or /path/to/local/template"
	ti.CharLimit = 200
	ti.Width = 50
	ti.Focus()

	return &TemplateSelectorModel{
		input:           ti,
		recentTemplates: recentTemplates,
		localTemplates:  []string{},
		sourceType:      TemplateSourceGitHub,
		cursor:          -1, // -1 means input is focused, not list
		width:           60,
		height:          20,
		loading:         false,
		err:             nil,
	}
}

// SetLocalTemplates sets the list of local template directories.
func (m *TemplateSelectorModel) SetLocalTemplates(templates []string) {
	m.localTemplates = templates
}

// GetSourceType returns the current source type.
func (m *TemplateSelectorModel) GetSourceType() TemplateSourceType {
	return m.sourceType
}

// ToggleSourceType toggles between GitHub and local sources.
func (m *TemplateSelectorModel) ToggleSourceType() {
	if m.sourceType == TemplateSourceGitHub {
		m.sourceType = TemplateSourceLocal
		m.input.Placeholder = "/path/to/local/template"
	} else {
		m.sourceType = TemplateSourceGitHub
		m.input.Placeholder = "owner/repo (e.g., MoshPitCodes/template-go)"
	}
	m.cursor = -1
	m.input.SetValue("")
	m.input.Focus()
}

// SetRecentTemplates updates the recent templates list.
func (m *TemplateSelectorModel) SetRecentTemplates(templates []string) {
	m.recentTemplates = templates
}

// Show displays the template selector.
func (m *TemplateSelectorModel) Show() {
	m.visible = true
	m.input.Focus()
}

// Hide hides the template selector.
func (m *TemplateSelectorModel) Hide() {
	m.visible = false
}

// IsVisible returns true if the selector is visible.
func (m *TemplateSelectorModel) IsVisible() bool {
	return m.visible
}

// SetSize sets the dimensions of the selector.
func (m *TemplateSelectorModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.Width = width - 10
}

// SetLoading sets the loading state.
func (m *TemplateSelectorModel) SetLoading(loading bool) {
	m.loading = loading
}

// SetError sets an error to display.
func (m *TemplateSelectorModel) SetError(err error) {
	m.err = err
}

// Reset resets the selector to initial state.
func (m *TemplateSelectorModel) Reset() {
	m.input.SetValue("")
	m.cursor = -1
	m.loading = false
	m.err = nil
	m.input.Focus()
}

// Update handles messages for the template selector.
func (m *TemplateSelectorModel) Update(msg tea.Msg) (*TemplateSelectorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// While loading, only allow ESC to cancel - block all other keys
		if m.loading {
			switch msg.String() {
			case "esc":
				// Allow ESC to cancel/close even during loading
				return m, nil
			default:
				// Block all other keys while loading
				return m, nil
			}
		}

		switch msg.String() {
		case "enter":
			return m.handleSubmit()

		case "up", "k":
			if m.cursor > -1 {
				m.cursor--
				if m.cursor == -1 {
					m.input.Focus()
				}
			}
			return m, nil

		case "down", "j":
			currentList := m.getCurrentList()
			if len(currentList) > 0 && m.cursor < len(currentList)-1 {
				m.cursor++
				m.input.Blur()
			}
			return m, nil

		case "ctrl+t":
			// Toggle between GitHub and local source
			m.ToggleSourceType()
			return m, nil

		case "tab":
			// Toggle between input and list
			currentList := m.getCurrentList()
			if m.cursor == -1 && len(currentList) > 0 {
				m.cursor = 0
				m.input.Blur()
			} else {
				m.cursor = -1
				m.input.Focus()
			}
			return m, nil

		default:
			// If typing, focus the input
			if m.cursor != -1 && msg.Type == tea.KeyRunes {
				m.cursor = -1
				m.input.Focus()
			}

			// Update text input
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}

	// Update text input for other messages
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// getCurrentList returns the current list based on source type.
func (m *TemplateSelectorModel) getCurrentList() []string {
	if m.sourceType == TemplateSourceLocal {
		return m.localTemplates
	}
	return m.recentTemplates
}

// handleSubmit handles the enter key submission.
func (m *TemplateSelectorModel) handleSubmit() (*TemplateSelectorModel, tea.Cmd) {
	currentList := m.getCurrentList()

	if m.sourceType == TemplateSourceLocal {
		// Local template selection
		var localPath string

		if m.cursor >= 0 && m.cursor < len(currentList) {
			localPath = currentList[m.cursor]
		} else {
			localPath = strings.TrimSpace(m.input.Value())
		}

		if localPath != "" {
			m.loading = true
			m.err = nil
			return m, func() tea.Msg {
				return TemplateRepoSelectedMsg{
					LocalPath: localPath,
					IsLocal:   true,
				}
			}
		}

		m.err = fmt.Errorf("please enter a valid local path")
		return m, nil
	}

	// GitHub template selection
	var owner, repo string

	if m.cursor >= 0 && m.cursor < len(currentList) {
		// Selected from recent list
		parts := strings.SplitN(currentList[m.cursor], "/", 2)
		if len(parts) == 2 {
			owner = parts[0]
			repo = parts[1]
		}
	} else {
		// Parse from input
		value := strings.TrimSpace(m.input.Value())
		parts := strings.SplitN(value, "/", 2)
		if len(parts) == 2 {
			owner = strings.TrimSpace(parts[0])
			repo = strings.TrimSpace(parts[1])
		}
	}

	if owner != "" && repo != "" {
		m.loading = true
		m.err = nil
		return m, func() tea.Msg {
			return TemplateRepoSelectedMsg{
				Owner:   owner,
				Repo:    repo,
				IsLocal: false,
			}
		}
	}

	m.err = fmt.Errorf("please enter a valid owner/repo format")
	return m, nil
}

// View renders the template selector.
func (m *TemplateSelectorModel) View() string {
	var b strings.Builder

	// Title with source type indicator
	sourceIcon := "🌐"
	sourceLabel := "GitHub"
	if m.sourceType == TemplateSourceLocal {
		sourceIcon = "📁"
		sourceLabel = "Local"
	}

	title := templateSelectorTitleStyle.Render(fmt.Sprintf("📋 Select Template (%s %s)", sourceIcon, sourceLabel))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Loading state
	if m.loading {
		b.WriteString(templateSelectorLoadingStyle.Render("Loading template..."))
		return templateSelectorStyle.Width(m.width).Render(b.String())
	}

	// Error display
	if m.err != nil {
		b.WriteString(errorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	// Input field
	var inputLabel string
	if m.sourceType == TemplateSourceLocal {
		inputLabel = "Enter local path:"
	} else {
		inputLabel = "Enter repository (owner/repo):"
	}
	b.WriteString(templateSelectorLabelStyle.Render(inputLabel))
	b.WriteString("\n")

	inputStyle := templateSelectorInputStyle
	if m.cursor == -1 {
		inputStyle = templateSelectorInputFocusedStyle
	}
	b.WriteString(inputStyle.Render(m.input.View()))
	b.WriteString("\n\n")

	// Templates list (recent GitHub or local repos)
	currentList := m.getCurrentList()
	if len(currentList) > 0 {
		var listLabel string
		if m.sourceType == TemplateSourceLocal {
			listLabel = "Local Repositories:"
		} else {
			listLabel = "Recent Templates:"
		}
		b.WriteString(templateSelectorLabelStyle.Render(listLabel))
		b.WriteString("\n")

		maxDisplay := 8
		displayList := currentList
		if len(displayList) > maxDisplay {
			displayList = displayList[:maxDisplay]
		}

		for i, tmpl := range displayList {
			var prefix string
			var style lipgloss.Style

			if i == m.cursor {
				prefix = "▸ "
				style = templateSelectorItemSelectedStyle
			} else {
				prefix = "  "
				style = templateSelectorItemStyle
			}

			icon := "📋"
			if m.sourceType == TemplateSourceLocal {
				icon = "📁"
			}

			item := fmt.Sprintf("%s%s %s", prefix, icon, tmpl)
			b.WriteString(style.Render(item))
			b.WriteString("\n")
		}

		if len(currentList) > maxDisplay {
			b.WriteString(templateSelectorHintStyle.Render(
				fmt.Sprintf("  ... and %d more", len(currentList)-maxDisplay),
			))
			b.WriteString("\n")
		}
	} else {
		if m.sourceType == TemplateSourceLocal {
			b.WriteString(templateSelectorHintStyle.Render("No local repositories available"))
		} else {
			b.WriteString(templateSelectorHintStyle.Render("No recent templates"))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Help text with source toggle hint
	helpText := "↑/↓ navigate • enter select • ctrl+t toggle GitHub/Local"
	b.WriteString(templateSelectorHelpStyle.Render(helpText))

	return templateSelectorStyle.Width(m.width).Render(b.String())
}

// Styles for template selector
var (
	templateSelectorStyle = lipgloss.NewStyle().
				Padding(2, 3).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor)

	templateSelectorTitleStyle = lipgloss.NewStyle().
					Foreground(primaryColor).
					Bold(true).
					MarginBottom(1)

	templateSelectorLabelStyle = lipgloss.NewStyle().
					Foreground(fgColor).
					Bold(true)

	templateSelectorInputStyle = lipgloss.NewStyle().
					Padding(0, 1).
					Border(lipgloss.NormalBorder()).
					BorderForeground(borderColor)

	templateSelectorInputFocusedStyle = lipgloss.NewStyle().
						Padding(0, 1).
						Border(lipgloss.NormalBorder()).
						BorderForeground(secondaryColor)

	templateSelectorItemStyle = lipgloss.NewStyle().
					Foreground(fgColor).
					Padding(0, 1)

	templateSelectorItemSelectedStyle = lipgloss.NewStyle().
						Foreground(secondaryColor).
						Bold(true).
						Padding(0, 1)

	templateSelectorHintStyle = lipgloss.NewStyle().
					Foreground(mutedColor).
					Italic(true)

	templateSelectorHelpStyle = lipgloss.NewStyle().
					Foreground(mutedColor)

	templateSelectorLoadingStyle = lipgloss.NewStyle().
					Foreground(secondaryColor).
					Italic(true)
)
