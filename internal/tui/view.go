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

	// Template mode has its own workflow
	if m.mode == ModeTemplate {
		sections = append(sections, m.renderTemplateWorkflow())
	} else {
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

	// Template-specific overlays
	if m.mode == ModeTemplate {
		// Show template selector as overlay (like settings)
		if m.templateSelector != nil && m.templateSelector.IsVisible() {
			view = m.renderWithOverlay(view, m.renderTemplateSelectorOverlay())
		}

		// Show conflict dialog as overlay
		if m.templateConflict != nil && m.templateConflict.IsVisible() {
			view = m.renderWithOverlay(view, m.templateConflict.View())
		}
	}

	return view
}

// renderHeader renders the application header.
func (m Model) renderHeader() string {
	title := headerTitleStyle.Render("🔄 RepoSync")
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

	// Only set width if we have a valid width value
	if m.width > 0 {
		return headerStyle.Width(m.width).Render(content)
	}
	return headerStyle.Render(content)
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

	// Only set width if we have a valid width value
	if m.width > 0 {
		return style.Width(m.width).Render(content)
	}
	return style.Render(content)
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

	// Only set width if we have a valid width value
	if m.width > 0 {
		return style.Width(m.width).Render(progressView)
	}
	return style.Render(progressView)
}

// renderFooter renders the footer with keyboard shortcuts.
func (m Model) renderFooter() string {
	var bindings []string

	if m.mode == ModeTemplate {
		// Template mode bindings based on current step
		if m.templateState == nil || m.templateState.Step == StepSelectTemplate {
			bindings = []string{
				"s/enter", "select template",
				"?", "help",
				"q", "quit",
			}
		} else if m.templateState.Step == StepBrowseTree {
			bindings = []string{
				"↑/↓", "navigate",
				"space", "toggle",
				"a/n", "all/none",
				"←/→", "collapse/expand",
				"e/c", "expand/collapse all",
				"enter", "continue",
				"esc", "back",
				"q", "quit",
			}
		} else if m.templateState.Step == StepSelectTargets {
			bindings = []string{
				"↑/↓", "navigate",
				"space", "toggle",
				"a/n", "all/none",
				"type", "filter",
				"enter", "sync",
				"esc", "back",
				"q", "quit",
			}
		} else if m.templateState.Step == StepComplete {
			bindings = []string{
				"enter/esc", "continue",
				"q", "quit",
			}
		} else {
			bindings = []string{
				"?", "help",
				"q", "quit",
			}
		}
	} else if m.mode == ModeLocal {
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
		"4", "Templates",
		"tab", "Next tab",
		"shift+tab", "Previous tab",
	}

	if m.mode == ModeTemplate {
		// Template-specific help
		sections["Template Selection"] = []string{
			"enter", "Open template selector",
			"ctrl+t", "Toggle GitHub/Local source",
			"↑/↓", "Navigate recent templates",
		}

		sections["Tree Browser"] = []string{
			"↑/↓", "Navigate tree",
			"enter", "Expand/collapse folder",
			"space", "Toggle selection",
			"a", "Select all files",
			"n", "Deselect all",
			"e", "Expand all folders",
			"c", "Collapse all folders",
		}

		sections["Target Selection"] = []string{
			"↑/↓", "Navigate list",
			"space", "Toggle selection",
			"a", "Select all targets",
			"n", "Deselect all",
		}

		sections["Conflict Resolution"] = []string{
			"o", "Overwrite file",
			"s", "Skip file",
			"O", "Overwrite all",
			"S", "Skip all",
		}
	} else {
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
	}

	return RenderHelpOverlay(sections)
}

// renderWithOverlay renders content with an overlay centered on top.
func (m Model) renderWithOverlay(base, overlay string) string {
	// Simply use lipgloss.Place to center the overlay.
	// The overlay will be shown on a backdrop, and when it's dismissed,
	// the base view will be regenerated properly.

	// Use the terminal dimensions for placement
	width := m.width
	height := m.height

	if width == 0 {
		width = 100
	}
	if height == 0 {
		height = 30
	}

	// Place the overlay in the center with a semi-transparent background effect
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars("░"),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#2a2a2a")),
	)
}

// renderTemplateWorkflow renders the template sync workflow based on current step.
func (m Model) renderTemplateWorkflow() string {
	if m.templateState == nil {
		return m.renderTemplateWelcome()
	}

	switch m.templateState.Step {
	case StepSelectTemplate:
		return m.renderTemplateWelcome()
	case StepBrowseTree:
		return m.renderTemplateTree()
	case StepSelectTargets:
		return m.renderTemplateTargets()
	case StepSyncing:
		return m.renderTemplateSyncProgress()
	case StepComplete:
		return m.renderTemplateSyncComplete()
	default:
		return m.renderTemplateWelcome()
	}
}

// renderTemplateWelcome renders the template mode welcome/prompt screen.
func (m Model) renderTemplateWelcome() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Render("📋 Template Sync")

	b.WriteString(title)
	b.WriteString("\n\n")

	desc := lipgloss.NewStyle().
		Foreground(fgColor).
		Render("Sync files from a template repository to your local repositories.")

	b.WriteString(desc)
	b.WriteString("\n\n")

	// Show workflow steps
	steps := []string{
		"1. Select a template repository (GitHub or Local)",
		"2. Browse and select files to sync",
		"3. Choose target repositories",
		"4. Sync files to targets",
	}

	for _, step := range steps {
		stepStyle := lipgloss.NewStyle().Foreground(mutedColor)
		b.WriteString(stepStyle.Render("  " + step))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	hint := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Render("Press 's' or Enter to select a template...")

	b.WriteString(hint)
	b.WriteString("\n")

	style := lipgloss.NewStyle().
		Padding(2, 4).
		MarginTop(1)

	if m.width > 0 {
		style = style.Width(m.width)
	}

	return style.Render(b.String())
}

// renderTemplateSelectorOverlay renders the template selector as a popup overlay.
func (m Model) renderTemplateSelectorOverlay() string {
	if m.templateSelector == nil {
		return ""
	}

	// Set reasonable size for the popup
	selectorWidth := 70
	if m.width > 0 && m.width < 80 {
		selectorWidth = m.width - 10
	}

	selectorHeight := 25
	if m.height > 0 && m.height < 35 {
		selectorHeight = m.height - 10
	}

	m.templateSelector.SetSize(selectorWidth, selectorHeight)

	return m.templateSelector.View()
}

// renderTemplateTree renders the template tree browser.
func (m Model) renderTemplateTree() string {
	if m.templateTree == nil {
		return lipgloss.NewStyle().
			Foreground(warningColor).
			Padding(2, 4).
			Render("Loading template tree...")
	}

	// Update tree size based on available space
	// Main UI chrome in template mode:
	// - Tabs: 3 lines (content + border + margin)
	// - Footer: 6 lines (content + padding + border + margin)
	// Total: 9 lines
	mainChrome := 9
	treeHeight := m.height - mainChrome
	if treeHeight < 10 {
		treeHeight = 10
	}

	// Safely set size (avoid negative values)
	treeWidth := m.width - 8
	if treeWidth < 40 {
		treeWidth = 40
	}
	m.templateTree.SetSize(treeWidth, treeHeight)

	return m.templateTree.View()
}

// renderTemplateTargets renders the target repository selector.
func (m Model) renderTemplateTargets() string {
	if m.templateTargets == nil {
		return lipgloss.NewStyle().
			Foreground(warningColor).
			Padding(2, 4).
			Render("Loading target repositories...")
	}

	// Update targets size based on available space
	targetsHeight := m.height - 12
	if targetsHeight < 10 {
		targetsHeight = 10
	}

	// Safely set size (avoid negative values)
	targetsWidth := m.width - 8
	if targetsWidth < 40 {
		targetsWidth = 40
	}
	m.templateTargets.SetSize(targetsWidth, targetsHeight)

	return m.templateTargets.View()
}

// renderTemplateSyncProgress renders the template sync progress.
func (m Model) renderTemplateSyncProgress() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Render("📋 Syncing Template Files")

	b.WriteString(title)
	b.WriteString("\n\n")

	if m.templateState != nil {
		// Progress bar
		progress := float64(m.templateState.SyncProgress.Current) / float64(m.templateState.SyncProgress.Total)
		if m.templateState.SyncProgress.Total == 0 {
			progress = 0
		}

		barWidth := 40
		filled := int(progress * float64(barWidth))
		empty := barWidth - filled

		bar := lipgloss.NewStyle().Foreground(successColor).Render(strings.Repeat("█", filled))
		bar += lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("░", empty))

		percentage := fmt.Sprintf(" %.0f%%", progress*100)
		b.WriteString(bar)
		b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render(percentage))
		b.WriteString("\n\n")

		// Current file info
		if m.templateState.SyncProgress.CurrentFile != "" {
			fileInfo := fmt.Sprintf("Syncing: %s", m.templateState.SyncProgress.CurrentFile)
			b.WriteString(lipgloss.NewStyle().Foreground(fgColor).Render(fileInfo))
			b.WriteString("\n")
		}

		if m.templateState.SyncProgress.TargetRepo != "" {
			targetInfo := fmt.Sprintf("Target: %s", m.templateState.SyncProgress.TargetRepo)
			b.WriteString(lipgloss.NewStyle().Foreground(secondaryColor).Render(targetInfo))
			b.WriteString("\n")
		}

		// Progress stats
		stats := fmt.Sprintf("\n%d/%d files processed",
			m.templateState.SyncProgress.Current,
			m.templateState.SyncProgress.Total)
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(stats))
	}

	style := lipgloss.NewStyle().
		Padding(2, 4).
		MarginTop(1)

	if m.width > 0 {
		style = style.Width(m.width)
	}

	return style.Render(b.String())
}

// renderTemplateSyncComplete renders the sync completion summary.
func (m Model) renderTemplateSyncComplete() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true).
		Render("✓ Template Sync Complete")

	b.WriteString(title)
	b.WriteString("\n\n")

	if m.templateState != nil {
		// Use the deprecated fields which are actually populated
		synced := m.templateState.SyncedCount
		skipped := m.templateState.SkippedCount
		errors := m.templateState.ErrorCount

		// Summary stats
		if synced > 0 {
			syncedStr := fmt.Sprintf("✓ %d files synced", synced)
			b.WriteString(lipgloss.NewStyle().Foreground(successColor).Render(syncedStr))
			b.WriteString("\n")
		}

		if skipped > 0 {
			skippedStr := fmt.Sprintf("○ %d files skipped", skipped)
			b.WriteString(lipgloss.NewStyle().Foreground(warningColor).Render(skippedStr))
			b.WriteString("\n")
		}

		if errors > 0 {
			errorsStr := fmt.Sprintf("✗ %d errors", errors)
			b.WriteString(lipgloss.NewStyle().Foreground(errorColor).Render(errorsStr))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	hint := lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Render("Press Enter or Esc to continue...")

	b.WriteString(hint)

	style := lipgloss.NewStyle().
		Padding(2, 4).
		MarginTop(1)

	if m.width > 0 {
		style = style.Width(m.width)
	}

	return style.Render(b.String())
}
