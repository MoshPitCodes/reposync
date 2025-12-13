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

// TemplateConflictModel manages the conflict resolution dialog.
type TemplateConflictModel struct {
	// Current conflict information
	filePath       string
	targetRepoPath string
	targetRepoName string

	// Is the dialog visible
	visible bool

	// Cursor position (0=Overwrite, 1=Skip, 2=OverwriteAll, 3=SkipAll)
	cursor int

	// Dimensions
	width int
}

// NewTemplateConflictModel creates a new conflict dialog model.
func NewTemplateConflictModel() *TemplateConflictModel {
	return &TemplateConflictModel{
		visible: false,
		cursor:  0,
		width:   50,
	}
}

// Show displays the conflict dialog for a specific file conflict.
func (m *TemplateConflictModel) Show(filePath, targetRepoPath string) {
	m.filePath = filePath
	m.targetRepoPath = targetRepoPath
	m.targetRepoName = filepath.Base(targetRepoPath)
	m.visible = true
	m.cursor = 0
}

// Hide hides the dialog.
func (m *TemplateConflictModel) Hide() {
	m.visible = false
}

// IsVisible returns whether the dialog is visible.
func (m *TemplateConflictModel) IsVisible() bool {
	return m.visible
}

// SetWidth sets the dialog width.
func (m *TemplateConflictModel) SetWidth(width int) {
	m.width = width
}

// Update handles messages for the conflict dialog.
func (m *TemplateConflictModel) Update(msg tea.Msg) (*TemplateConflictModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "right", "l":
			if m.cursor < 3 {
				m.cursor++
			}
			return m, nil

		case "up", "k":
			// Move between rows (0-1 <-> 2-3)
			if m.cursor >= 2 {
				m.cursor -= 2
			}
			return m, nil

		case "down", "j":
			// Move between rows (0-1 -> 2-3)
			if m.cursor < 2 {
				m.cursor += 2
			}
			return m, nil

		case "o":
			// Overwrite
			m.visible = false
			return m, func() tea.Msg {
				return TemplateConflictResponseMsg{
					Action:   ConflictOverwrite,
					FilePath: m.filePath,
				}
			}

		case "s":
			// Skip
			m.visible = false
			return m, func() tea.Msg {
				return TemplateConflictResponseMsg{
					Action:   ConflictSkip,
					FilePath: m.filePath,
				}
			}

		case "O":
			// Overwrite All
			m.visible = false
			return m, func() tea.Msg {
				return TemplateConflictResponseMsg{
					Action:   ConflictOverwriteAll,
					FilePath: m.filePath,
				}
			}

		case "S":
			// Skip All
			m.visible = false
			return m, func() tea.Msg {
				return TemplateConflictResponseMsg{
					Action:   ConflictSkipAll,
					FilePath: m.filePath,
				}
			}

		case "enter", " ":
			// Select current option
			action := TemplateConflictAction(m.cursor)
			m.visible = false
			return m, func() tea.Msg {
				return TemplateConflictResponseMsg{
					Action:   action,
					FilePath: m.filePath,
				}
			}

		case "esc":
			// Escape = Skip
			m.visible = false
			return m, func() tea.Msg {
				return TemplateConflictResponseMsg{
					Action:   ConflictSkip,
					FilePath: m.filePath,
				}
			}
		}
	}

	return m, nil
}

// View renders the conflict dialog.
func (m *TemplateConflictModel) View() string {
	if !m.visible {
		return ""
	}

	var b strings.Builder

	// Title
	title := templateConflictTitleStyle.Render("⚠ File Conflict")
	b.WriteString(title)
	b.WriteString("\n\n")

	// File information
	b.WriteString(templateConflictLabelStyle.Render("File:"))
	b.WriteString(" ")
	b.WriteString(templateConflictFileStyle.Render(m.filePath))
	b.WriteString("\n")

	b.WriteString(templateConflictLabelStyle.Render("Target:"))
	b.WriteString(" ")
	b.WriteString(templateConflictTargetStyle.Render(m.targetRepoName))
	b.WriteString("\n\n")

	// Message
	message := "This file already exists in the target repository.\nWhat would you like to do?"
	b.WriteString(templateConflictMessageStyle.Render(message))
	b.WriteString("\n\n")

	// Options - two rows
	options := []struct {
		key   string
		label string
		hint  string
	}{
		{"o", "Overwrite", "Replace this file"},
		{"s", "Skip", "Keep existing file"},
		{"O", "Overwrite All", "Replace all conflicts"},
		{"S", "Skip All", "Keep all existing"},
	}

	// Row 1
	row1 := make([]string, 2)
	for i := 0; i < 2; i++ {
		opt := options[i]
		style := templateConflictOptionStyle
		if m.cursor == i {
			style = templateConflictOptionSelectedStyle
		}
		row1[i] = style.Render(fmt.Sprintf("[%s] %s", opt.key, opt.label))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row1...))
	b.WriteString("\n")

	// Row 2
	row2 := make([]string, 2)
	for i := 2; i < 4; i++ {
		opt := options[i]
		style := templateConflictOptionStyle
		if m.cursor == i {
			style = templateConflictOptionSelectedStyle
		}
		row2[i-2] = style.Render(fmt.Sprintf("[%s] %s", opt.key, opt.label))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row2...))
	b.WriteString("\n\n")

	// Show hint for selected option
	if m.cursor >= 0 && m.cursor < len(options) {
		hint := options[m.cursor].hint
		b.WriteString(templateConflictHintStyle.Render(hint))
	}

	return templateConflictStyle.Width(m.width).Render(b.String())
}

// Styles for conflict dialog
var (
	templateConflictStyle = lipgloss.NewStyle().
				Padding(2, 3).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(warningColor).
				Background(bgColor)

	templateConflictTitleStyle = lipgloss.NewStyle().
					Foreground(warningColor).
					Bold(true)

	templateConflictLabelStyle = lipgloss.NewStyle().
					Foreground(mutedColor).
					Bold(true)

	templateConflictFileStyle = lipgloss.NewStyle().
					Foreground(accentColor).
					Bold(true)

	templateConflictTargetStyle = lipgloss.NewStyle().
					Foreground(secondaryColor)

	templateConflictMessageStyle = lipgloss.NewStyle().
					Foreground(fgColor)

	templateConflictOptionStyle = lipgloss.NewStyle().
					Foreground(fgColor).
					Padding(0, 2).
					MarginRight(2)

	templateConflictOptionSelectedStyle = lipgloss.NewStyle().
						Foreground(warningColor).
						Bold(true).
						Padding(0, 2).
						MarginRight(2).
						Underline(true)

	templateConflictHintStyle = lipgloss.NewStyle().
					Foreground(mutedColor).
					Italic(true)
)
