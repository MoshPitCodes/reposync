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

	"github.com/charmbracelet/lipgloss"
)

// renderView renders the complete unified view.
func (m Model) renderView() string {
	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Tabs
	sections = append(sections, m.renderTabs())

	// Owner bar (for GitHub modes)
	if m.mode != ModeLocal {
		sections = append(sections, m.renderOwnerBar())
	}

	// List
	sections = append(sections, m.renderList())

	// Progress bar (when syncing)
	if m.syncing || m.progress.IsComplete() {
		sections = append(sections, m.renderProgress())
	}

	// Footer
	sections = append(sections, m.renderFooter())

	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Overlays
	if m.showSettings {
		view = m.renderWithOverlay(view, m.renderSettingsOverlay())
	}

	if m.showHelp {
		view = m.renderWithOverlay(view, m.renderHelpOverlay())
	}

	if m.ownerSelector.IsExpanded() {
		view = m.renderWithOverlay(view, m.ownerSelector.View())
	}

	if m.repoExistsDialog.IsVisible() {
		view = m.renderWithOverlay(view, m.repoExistsDialog.View())
	}

	return view
}

// renderHeader renders the application header.
func (m Model) renderHeader() string {
	title := headerTitleStyle.Render("🔄 Repo Sync")
	version := headerVersionStyle.Render(AppVersion)

	rightSection := ""
	if m.mode != ModeLocal {
		rightSection = " [?] help [c] settings"
	} else {
		rightSection = " [?] help [c] settings"
	}

	spacer := ""
	if m.width > 0 {
		usedWidth := lipgloss.Width(title) + lipgloss.Width(version) + lipgloss.Width(rightSection)
		if m.width > usedWidth {
			spacer = strings.Repeat(" ", m.width-usedWidth-4)
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Left,
		title,
		spacer,
		version,
		headerVersionStyle.Render(rightSection),
	)

	return headerStyle.Width(m.width).Render(content)
}

// renderTabs renders the tab bar.
func (m Model) renderTabs() string {
	return m.tabs.ViewWithContainer()
}

// renderOwnerBar renders the owner selection bar.
func (m Model) renderOwnerBar() string {
	icon := "👤"
	if m.mode == ModeOrganization {
		icon = "🏢"
	}

	selectedCount := m.list.GetSelectedCount()
	totalCount := len(m.list.filtered)

	leftPart := fmt.Sprintf("Owner: %s %s", icon, m.owner)
	rightPart := fmt.Sprintf("%d selected / %d", selectedCount, totalCount)

	spacer := ""
	if m.width > 0 {
		usedWidth := lipgloss.Width(leftPart) + lipgloss.Width(rightPart)
		if m.width > usedWidth+4 {
			spacer = strings.Repeat(" ", m.width-usedWidth-4)
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(leftPart),
		spacer,
		lipgloss.NewStyle().Foreground(accentColor).Render(rightPart),
	)

	style := lipgloss.NewStyle().
		Padding(0, 2).
		MarginBottom(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(borderColor)

	return style.Width(m.width).Render(content)
}

// renderList renders the repository list.
func (m Model) renderList() string {
	return m.list.View(m.width, m.height)
}

// renderProgress renders the inline progress bar.
func (m Model) renderProgress() string {
	if !m.syncing && !m.progress.IsComplete() {
		return ""
	}

	progressView := m.progress.View()
	if progressView == "" {
		return ""
	}

	style := lipgloss.NewStyle().
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(borderColor)

	return style.Width(m.width).Render(progressView)
}

// renderFooter renders the footer with keyboard shortcuts.
func (m Model) renderFooter() string {
	var bindings []string

	if m.mode == ModeLocal {
		bindings = []string{
			"↑/↓", "navigate",
			"space", "toggle",
			"a/n", "all/none",
			"/", "search",
			"s", "sort",
			"enter", "sync",
			"?", "help",
			"q", "quit",
		}
	} else {
		bindings = []string{
			"↑/↓", "navigate",
			"space", "toggle",
			"a/n", "all/none",
			"/", "search",
			"s", "sort",
			"o", "owner",
			"enter", "sync",
			"?", "help",
			"q", "quit",
		}
	}

	return RenderFooter(bindings...)
}

// renderSettingsOverlay renders the settings modal overlay.
func (m Model) renderSettingsOverlay() string {
	return m.settings.View()
}

// renderHelpOverlay renders the help overlay.
func (m Model) renderHelpOverlay() string {
	sections := make(map[string][]string)

	// Global shortcuts
	sections["Global"] = []string{
		"?", "Toggle this help",
		"q", "Quit application",
		"c", "Open settings",
		"esc", "Close overlay",
	}

	// Tab navigation
	sections["Tabs"] = []string{
		"1", "Personal repositories",
		"2", "Organization repositories",
		"3", "Local repositories",
		"tab", "Next tab",
		"shift+tab", "Previous tab",
	}

	// List navigation
	sections["Navigation"] = []string{
		"↑/k", "Move up",
		"↓/j", "Move down",
		"pgup", "Page up",
		"pgdown", "Page down",
	}

	// Selection
	sections["Selection"] = []string{
		"space", "Toggle selection",
		"a", "Select all",
		"n", "Deselect all",
	}

	// Actions
	sections["Actions"] = []string{
		"/", "Search/filter",
		"s", "Cycle sort mode",
		"enter", "Start sync",
	}

	if m.mode != ModeLocal {
		sections["GitHub"] = []string{
			"o", "Change owner",
		}
	}

	return RenderHelpOverlay(sections)
}

// renderWithOverlay renders content with an overlay centered on top.
func (m Model) renderWithOverlay(base, overlay string) string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
	)
}
