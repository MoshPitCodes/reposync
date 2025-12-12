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

// OwnerSelectorModel manages the owner selector dropdown.
type OwnerSelectorModel struct {
	selectedOwner string
	isOrg         bool
	username      string
	orgs          []string
	expanded      bool
	filterInput   textinput.Model
	cursor        int
	width         int
}

// NewOwnerSelectorModel creates a new owner selector model.
func NewOwnerSelectorModel(username string) *OwnerSelectorModel {
	ti := textinput.New()
	ti.Placeholder = "Filter organizations..."
	ti.CharLimit = 50

	return &OwnerSelectorModel{
		selectedOwner: username,
		isOrg:         false,
		username:      username,
		orgs:          []string{},
		expanded:      false,
		filterInput:   ti,
		cursor:        0,
		width:         60,
	}
}

// SetOrgs sets the list of organizations.
func (m *OwnerSelectorModel) SetOrgs(orgs []string) {
	m.orgs = orgs
}

// SetSelectedOwner sets the selected owner.
func (m *OwnerSelectorModel) SetSelectedOwner(owner string, isOrg bool) {
	m.selectedOwner = owner
	m.isOrg = isOrg
}

// GetSelectedOwner returns the current selected owner and whether it's an org.
func (m *OwnerSelectorModel) GetSelectedOwner() (string, bool) {
	return m.selectedOwner, m.isOrg
}

// Toggle toggles the dropdown expansion.
func (m *OwnerSelectorModel) Toggle() {
	m.expanded = !m.expanded
	if m.expanded {
		m.filterInput.Focus()
		m.cursor = 0
	} else {
		m.filterInput.Blur()
		m.filterInput.SetValue("")
	}
}

// Close closes the dropdown.
func (m *OwnerSelectorModel) Close() {
	m.expanded = false
	m.filterInput.Blur()
	m.filterInput.SetValue("")
	m.cursor = 0
}

// IsExpanded returns whether the dropdown is expanded.
func (m *OwnerSelectorModel) IsExpanded() bool {
	return m.expanded
}

// Update handles messages for the owner selector.
func (m *OwnerSelectorModel) Update(msg tea.Msg) (*OwnerSelectorModel, tea.Cmd) {
	if !m.expanded {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Close()
			return m, nil

		case "enter":
			// Select the current item
			items := m.getFilteredItems()
			if m.cursor >= 0 && m.cursor < len(items) {
				if m.cursor == 0 {
					// Personal
					m.selectedOwner = m.username
					m.isOrg = false
				} else {
					// Organization
					m.selectedOwner = items[m.cursor]
					m.isOrg = true
				}
				m.Close()
				return m, func() tea.Msg {
					return SelectOwnerMsg{
						Owner: m.selectedOwner,
						IsOrg: m.isOrg,
					}
				}
			}
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			items := m.getFilteredItems()
			if m.cursor < len(items)-1 {
				m.cursor++
			}
			return m, nil

		default:
			// Update filter input
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			m.cursor = 0 // Reset cursor when filter changes
			return m, cmd
		}

	case OrgsLoadedMsg:
		m.SetOrgs(msg.Orgs)
	}

	return m, nil
}

// getFilteredItems returns the filtered list of items (personal + orgs).
func (m *OwnerSelectorModel) getFilteredItems() []string {
	filter := strings.ToLower(m.filterInput.Value())
	items := []string{m.username} // Personal is always first

	if filter == "" {
		// No filter, return all
		return append(items, m.orgs...)
	}

	// Filter organizations
	for _, org := range m.orgs {
		if strings.Contains(strings.ToLower(org), filter) {
			items = append(items, org)
		}
	}

	return items
}

// View renders the owner selector.
func (m *OwnerSelectorModel) View() string {
	ownerType := "Personal"
	if m.isOrg {
		ownerType = "Org"
	}

	indicator := "▼"
	if m.expanded {
		indicator = "▲"
	}

	label := fmt.Sprintf("Owner: %s (%s) %s", m.selectedOwner, ownerType, indicator)

	if !m.expanded {
		return ownerBarStyle.Render(label)
	}

	// Render dropdown
	var dropdown strings.Builder

	// Filter input
	dropdown.WriteString(ownerDropdownHeaderStyle.Render("Select Owner (type to filter)"))
	dropdown.WriteString("\n")
	dropdown.WriteString(focusedInputStyle.Render(m.filterInput.View()))
	dropdown.WriteString("\n\n")

	// Items
	items := m.getFilteredItems()
	maxDisplay := 10
	if len(items) > maxDisplay {
		items = items[:maxDisplay]
	}

	for i, item := range items {
		isPersonal := i == 0
		isCursor := i == m.cursor

		var prefix string
		if isCursor {
			prefix = "▸ "
		} else {
			prefix = "  "
		}

		var itemText string
		if isPersonal {
			itemText = fmt.Sprintf("%s👤 %s (Personal)", prefix, item)
		} else {
			itemText = fmt.Sprintf("%s🏢 %s", prefix, item)
		}

		var style lipgloss.Style
		if isCursor {
			style = selectedListItemStyle
		} else {
			style = listItemStyle
		}

		dropdown.WriteString(style.Render(itemText))
		dropdown.WriteString("\n")
	}

	dropdown.WriteString("\n")
	dropdown.WriteString(helpStyle.Render("↑/↓ navigate • enter select • esc close"))

	return ownerDropdownStyle.Render(dropdown.String())
}

// ViewInline renders a compact inline version of the owner selector.
func (m *OwnerSelectorModel) ViewInline() string {
	icon := "👤"
	if m.isOrg {
		icon = "🏢"
	}

	label := fmt.Sprintf("%s %s", icon, m.selectedOwner)
	return ownerInlineStyle.Render(label)
}

// Styles for owner selector
var (
	ownerBarStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			Background(bgColor).
			Padding(0, 2).
			MarginBottom(1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(borderColor)

	ownerInlineStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Padding(0, 1)

	ownerDropdownStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(secondaryColor).
				Padding(1, 2).
				Background(bgColor).
				Foreground(fgColor)

	ownerDropdownHeaderStyle = lipgloss.NewStyle().
					Foreground(primaryColor).
					Bold(true).
					Underline(true)
)
