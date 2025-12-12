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

	"github.com/charmbracelet/lipgloss"
)

var (
	// Enhanced Color palette - Modern, high contrast
	primaryColor   = lipgloss.Color("#8B5CF6")   // Vibrant Purple
	secondaryColor = lipgloss.Color("#06B6D4")   // Cyan
	accentColor    = lipgloss.Color("#EC4899")   // Pink
	successColor   = lipgloss.Color("#10B981")   // Green
	errorColor     = lipgloss.Color("#EF4444")   // Red
	warningColor   = lipgloss.Color("#F59E0B")   // Amber
	infoColor      = lipgloss.Color("#3B82F6")   // Blue
	mutedColor     = lipgloss.Color("#6B7280")   // Gray
	dimmedColor    = lipgloss.Color("#4B5563")   // Darker gray
	bgColor        = lipgloss.Color("#1E1E2E")   // Dark background
	fgColor        = lipgloss.Color("#E5E7EB")   // Light foreground
	borderColor    = lipgloss.Color("#374151")   // Border gray

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(0, 2)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			Background(primaryColor).
			Bold(true).
			Padding(0, 2).
			Width(100)

	headerTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	headerVersionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E9D5FF")).
				Italic(true)

	// Title styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1).
			Underline(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Menu styles
	menuItemStyle = lipgloss.NewStyle().
			Padding(1, 3).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor)

	selectedMenuItemStyle = menuItemStyle.Copy().
				Foreground(primaryColor).
				Bold(true).
				Background(bgColor).
				BorderForeground(primaryColor)

	menuIconStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			MarginRight(2)

	// List styles
	listHeaderStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1).
			BorderStyle(lipgloss.ThickBorder()).
			BorderBottom(true).
			BorderForeground(primaryColor)

	listItemStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginLeft(1)

	selectedListItemStyle = listItemStyle.Copy().
				Foreground(fgColor).
				Bold(true).
				Background(bgColor).
				BorderLeft(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(secondaryColor)

	checkedItemStyle = listItemStyle.Copy().
				Foreground(successColor)

	listMetadataStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

	listCountStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Border styles
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)

	focusedBorderStyle = borderStyle.Copy().
				BorderForeground(secondaryColor)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	// Button styles
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 3).
			MarginRight(2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)

	activeButtonStyle = buttonStyle.Copy().
				Background(secondaryColor).
				BorderForeground(secondaryColor).
				Bold(true)

	// Status styles
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(infoColor)

	// Help text styles
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			Padding(0, 1).
			Background(bgColor)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(fgColor)

	helpSeparatorStyle = lipgloss.NewStyle().
				Foreground(dimmedColor)

	// Footer styles
	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(bgColor).
			Padding(1, 2).
			MarginTop(1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(borderColor)

	// Help overlay styles
	helpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(accentColor).
				Padding(2, 3).
				Background(bgColor).
				Foreground(fgColor)

	helpOverlayTitleStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true).
				Underline(true).
				MarginBottom(1)

	helpOverlaySectionStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)

	// Progress styles
	spinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	progressBarStyle = lipgloss.NewStyle().
				Foreground(successColor)

	progressTextStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(mutedColor)

	tableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Input styles
	inputStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			Background(bgColor).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor)

	focusedInputStyle = inputStyle.Copy().
				BorderForeground(secondaryColor).
				BorderStyle(lipgloss.ThickBorder())

	searchPromptStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true).
				MarginRight(1)

	// Confirmation dialog styles
	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(warningColor).
			Padding(1, 2).
			Background(bgColor)

	dialogTitleStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	// Repository exists dialog styles
	repoExistsDialogStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(warningColor).
				Padding(2, 3).
				Background(bgColor).
				Foreground(fgColor)

	repoExistsDialogTitleStyle = lipgloss.NewStyle().
					Foreground(warningColor).
					Bold(true).
					Underline(true)

	repoExistsDialogRepoStyle = lipgloss.NewStyle().
					Foreground(accentColor).
					Bold(true)

	repoExistsDialogPathStyle = lipgloss.NewStyle().
					Foreground(mutedColor).
					Italic(true)

	repoExistsDialogHelpStyle = lipgloss.NewStyle().
					Foreground(dimmedColor).
					Italic(true)
)

// Helper functions for consistent formatting

const AppVersion = "v1.0.0"

// RenderHeader renders the application header with title and version.
func RenderHeader(width int) string {
	title := headerTitleStyle.Render("🔄 Repo Sync")
	version := headerVersionStyle.Render(AppVersion)
	spacer := lipgloss.NewStyle().Width(width - lipgloss.Width(title) - lipgloss.Width(version) - 4).Render("")

	content := lipgloss.JoinHorizontal(lipgloss.Left, title, spacer, version)
	return headerStyle.Width(width).Render(content)
}

// RenderFooter renders a footer with keyboard shortcuts.
func RenderFooter(bindings ...string) string {
	var parts []string
	for i := 0; i < len(bindings); i += 2 {
		if i+1 < len(bindings) {
			key := helpKeyStyle.Render(bindings[i])
			desc := helpDescStyle.Render(bindings[i+1])
			sep := helpSeparatorStyle.Render(" • ")
			parts = append(parts, key+" "+desc)
			if i+2 < len(bindings) {
				parts = append(parts, sep)
			}
		}
	}
	return footerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, parts...))
}

// RenderTitle renders a styled title with optional subtitle.
func RenderTitle(title, subtitle string) string {
	result := titleStyle.Render(title)
	if subtitle != "" {
		result += "\n" + subtitleStyle.Render(subtitle)
	}
	return result
}

// RenderMenuItem renders a menu item with icon and selection state.
func RenderMenuItem(icon, text string, selected bool) string {
	iconPart := menuIconStyle.Render(icon)
	content := iconPart + " " + text

	if selected {
		return selectedMenuItemStyle.Render("▸ " + content)
	}
	return menuItemStyle.Render("  " + content)
}

// RenderListItem renders a list item with selection and checked states.
func RenderListItem(text string, selected, checked bool) string {
	var prefix string
	if checked {
		prefix = successStyle.Render("✓")
	} else {
		prefix = lipgloss.NewStyle().Foreground(mutedColor).Render("○")
	}

	content := prefix + " " + text

	if selected {
		return selectedListItemStyle.Render("▸ " + content)
	}

	style := listItemStyle
	if checked {
		style = checkedItemStyle
	}

	return style.Render("  " + content)
}

// RenderListHeader renders a section header for lists.
func RenderListHeader(text string) string {
	return listHeaderStyle.Render(text)
}

// RenderButton renders a styled button.
func RenderButton(text string, active bool) string {
	if active {
		return activeButtonStyle.Render(text)
	}
	return buttonStyle.Render(text)
}

// RenderSuccess renders a success message.
func RenderSuccess(text string) string {
	return successStyle.Render("✓ " + text)
}

// RenderError renders an error message.
func RenderError(text string) string {
	return errorStyle.Render("✗ " + text)
}

// RenderWarning renders a warning message.
func RenderWarning(text string) string {
	return warningStyle.Render("⚠ " + text)
}

// RenderInfo renders an info message.
func RenderInfo(text string) string {
	return infoStyle.Render("ℹ " + text)
}

// RenderHelp renders help text with key bindings.
func RenderHelp(bindings ...string) string {
	var parts []string
	for i := 0; i < len(bindings); i += 2 {
		if i+1 < len(bindings) {
			key := helpKeyStyle.Render(bindings[i])
			desc := helpDescStyle.Render(bindings[i+1])
			sep := helpSeparatorStyle.Render(" • ")
			parts = append(parts, key+" "+desc)
			if i+2 < len(bindings) {
				parts = append(parts, sep)
			}
		}
	}
	return helpStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, parts...))
}

// RenderHelpOverlay renders a full help overlay with all keyboard shortcuts.
func RenderHelpOverlay(sections map[string][]string) string {
	var content []string

	content = append(content, helpOverlayTitleStyle.Render("Keyboard Shortcuts"))

	for sectionName, bindings := range sections {
		content = append(content, helpOverlaySectionStyle.Render(sectionName))

		for i := 0; i < len(bindings); i += 2 {
			if i+1 < len(bindings) {
				key := helpKeyStyle.Render(bindings[i])
				desc := helpDescStyle.Render(bindings[i+1])
				line := "  " + key + " - " + desc
				content = append(content, line)
			}
		}
	}

	content = append(content, "")
	content = append(content, helpDescStyle.Render("Press ? to close this help"))

	return helpOverlayStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

// RenderBorder renders content within a styled border.
func RenderBorder(content string, focused bool) string {
	if focused {
		return focusedBorderStyle.Render(content)
	}
	return borderStyle.Render(content)
}

// RenderBox renders content in a double-bordered box.
func RenderBox(content string) string {
	return boxStyle.Render(content)
}

// RenderMetadata renders metadata text in a muted style.
func RenderMetadata(text string) string {
	return listMetadataStyle.Render(text)
}

// RenderCount renders a count in an accented style.
func RenderCount(count int, total int) string {
	return listCountStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left,
		"Selected: ",
		fmt.Sprintf("%d", count),
		lipgloss.NewStyle().Foreground(mutedColor).Render("/"),
		fmt.Sprintf("%d", total),
	))
}

// RenderSearchPrompt renders a search prompt.
func RenderSearchPrompt(query string) string {
	prompt := searchPromptStyle.Render("🔍")
	return focusedInputStyle.Render(prompt + " " + query)
}

// RenderDialog renders a confirmation dialog.
func RenderDialog(title, message string, options ...string) string {
	content := dialogTitleStyle.Render(title) + "\n\n"
	content += message + "\n\n"

	for i, option := range options {
		content += RenderButton(option, i == 0) + " "
	}

	return dialogStyle.Render(content)
}
