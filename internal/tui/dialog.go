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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RepoExistsDialogModel manages the repository exists confirmation dialog.
type RepoExistsDialogModel struct {
	// Strings (16 bytes each)
	repoName string
	repoPath string
	mode     string

	// Int (8 bytes)
	repoIndex int

	// Bool (1 byte)
	visible bool
}

// NewRepoExistsDialogModel creates a new repository exists dialog.
func NewRepoExistsDialogModel() *RepoExistsDialogModel {
	return &RepoExistsDialogModel{
		visible: false,
	}
}

// Show displays the dialog with the given repository information.
func (m *RepoExistsDialogModel) Show(repoName, repoPath string, repoIndex int, mode string) {
	m.repoName = repoName
	m.repoPath = repoPath
	m.repoIndex = repoIndex
	m.mode = mode
	m.visible = true
}

// Hide hides the dialog.
func (m *RepoExistsDialogModel) Hide() {
	m.visible = false
}

// IsVisible returns whether the dialog is currently visible.
func (m *RepoExistsDialogModel) IsVisible() bool {
	return m.visible
}

// Update handles input for the dialog.
func (m *RepoExistsDialogModel) Update(msg tea.Msg) (*RepoExistsDialogModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "s":
			// Skip this repository
			m.visible = false
			return m, func() tea.Msg {
				return RepoExistsResponseMsg{
					Action:    ActionSkip,
					RepoIndex: m.repoIndex,
				}
			}

		case "r":
			// Refresh (git pull) this repository
			m.visible = false
			return m, func() tea.Msg {
				return RepoExistsResponseMsg{
					Action:    ActionRefresh,
					RepoIndex: m.repoIndex,
				}
			}

		case "S":
			// Skip all remaining repositories
			m.visible = false
			return m, func() tea.Msg {
				return RepoExistsResponseMsg{
					Action:    ActionSkipAll,
					RepoIndex: m.repoIndex,
				}
			}

		case "R":
			// Refresh all remaining repositories
			m.visible = false
			return m, func() tea.Msg {
				return RepoExistsResponseMsg{
					Action:    ActionRefreshAll,
					RepoIndex: m.repoIndex,
				}
			}

		case "esc":
			// Default to skip on escape
			m.visible = false
			return m, func() tea.Msg {
				return RepoExistsResponseMsg{
					Action:    ActionSkip,
					RepoIndex: m.repoIndex,
				}
			}
		}
	}

	return m, nil
}

// View renders the dialog.
func (m *RepoExistsDialogModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	title := repoExistsDialogTitleStyle.Render("Repository Already Exists")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Message
	message := "The repository " + repoExistsDialogRepoStyle.Render(m.repoName) + " already exists at:\n"
	message += repoExistsDialogPathStyle.Render(m.repoPath)
	content.WriteString(message)
	content.WriteString("\n\n")

	// Question
	content.WriteString("What would you like to do?")
	content.WriteString("\n\n")

	// Options
	optionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#8B5CF6")).
		Bold(true).
		Padding(0, 1)

	options := []string{
		keyStyle.Render("s") + " " + optionStyle.Render("Skip") + "        " +
			keyStyle.Render("r") + " " + optionStyle.Render("Refresh (git pull)"),
		keyStyle.Render("S") + " " + optionStyle.Render("Skip All") + "    " +
			keyStyle.Render("R") + " " + optionStyle.Render("Refresh All"),
	}

	content.WriteString(strings.Join(options, "\n"))
	content.WriteString("\n\n")

	// Help text
	helpText := repoExistsDialogHelpStyle.Render("Press ESC to skip")
	content.WriteString(helpText)

	return repoExistsDialogStyle.Render(content.String())
}
