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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the current view mode.
type ViewMode int

const (
	ModePersonal ViewMode = iota
	ModeOrganization
	ModeLocal
)

// String returns the string representation of the view mode.
func (v ViewMode) String() string {
	switch v {
	case ModePersonal:
		return "Personal"
	case ModeOrganization:
		return "Organizations"
	case ModeLocal:
		return "Local"
	default:
		return "Unknown"
	}
}

// Tab represents a single tab.
type Tab struct {
	ID       ViewMode
	Label    string
	Shortcut string
	Icon     string
}

// TabBarModel manages the tab bar component.
type TabBarModel struct {
	tabs   []Tab
	active ViewMode
}

// NewTabBarModel creates a new tab bar model.
func NewTabBarModel() *TabBarModel {
	return &TabBarModel{
		tabs: []Tab{
			{
				ID:       ModePersonal,
				Label:    "Personal",
				Shortcut: "1",
				Icon:     "👤",
			},
			{
				ID:       ModeOrganization,
				Label:    "Orgs",
				Shortcut: "2",
				Icon:     "🏢",
			},
			{
				ID:       ModeLocal,
				Label:    "Local",
				Shortcut: "3",
				Icon:     "📁",
			},
		},
		active: ModePersonal,
	}
}

// SetActive sets the active tab.
func (m *TabBarModel) SetActive(mode ViewMode) {
	m.active = mode
}

// GetActive returns the current active tab.
func (m *TabBarModel) GetActive() ViewMode {
	return m.active
}

// Next switches to the next tab.
func (m *TabBarModel) Next() ViewMode {
	m.active = (m.active + 1) % ViewMode(len(m.tabs))
	return m.active
}

// Prev switches to the previous tab.
func (m *TabBarModel) Prev() ViewMode {
	m.active = (m.active - 1 + ViewMode(len(m.tabs))) % ViewMode(len(m.tabs))
	return m.active
}

// Update handles messages for the tab bar.
func (m *TabBarModel) Update(msg tea.Msg) (*TabBarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.active = ModePersonal
			return m, func() tea.Msg {
				return SwitchModeMsg{Mode: ModePersonal}
			}
		case "2":
			m.active = ModeOrganization
			return m, func() tea.Msg {
				return SwitchModeMsg{Mode: ModeOrganization}
			}
		case "3":
			m.active = ModeLocal
			return m, func() tea.Msg {
				return SwitchModeMsg{Mode: ModeLocal}
			}
		case "tab":
			newMode := m.Next()
			return m, func() tea.Msg {
				return SwitchModeMsg{Mode: newMode}
			}
		case "shift+tab":
			newMode := m.Prev()
			return m, func() tea.Msg {
				return SwitchModeMsg{Mode: newMode}
			}
		}

	case SwitchModeMsg:
		m.active = msg.Mode
	}

	return m, nil
}

// View renders the tab bar.
func (m *TabBarModel) View() string {
	var tabs []string

	for _, tab := range m.tabs {
		isActive := tab.ID == m.active
		tabs = append(tabs, m.renderTab(tab, isActive))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// renderTab renders a single tab.
func (m *TabBarModel) renderTab(tab Tab, active bool) string {
	var style lipgloss.Style

	if active {
		style = activeTabStyle
	} else {
		style = inactiveTabStyle
	}

	label := fmt.Sprintf("%s %s", tab.Icon, tab.Label)
	if active {
		label = fmt.Sprintf("[%s: %s]", tab.Shortcut, label)
	} else {
		label = fmt.Sprintf(" %s: %s ", tab.Shortcut, label)
	}

	return style.Render(label)
}

// Styles for tabs
var (
	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Bold(true).
			Padding(0, 2).
			MarginRight(1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Background(bgColor).
				Padding(0, 2).
				MarginRight(1)

	tabBarContainerStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(borderColor).
				MarginBottom(1)
)

// ViewWithContainer renders the tab bar with a container border.
func (m *TabBarModel) ViewWithContainer() string {
	return tabBarContainerStyle.Render(m.View())
}

// GetTabByMode returns the tab for a given mode.
func (m *TabBarModel) GetTabByMode(mode ViewMode) *Tab {
	for i := range m.tabs {
		if m.tabs[i].ID == mode {
			return &m.tabs[i]
		}
	}
	return nil
}

// GetTabLabel returns the label for the active tab.
func (m *TabBarModel) GetTabLabel() string {
	tab := m.GetTabByMode(m.active)
	if tab != nil {
		return tab.Label
	}
	return ""
}

// RenderCompact renders a compact version of the tab bar.
func (m *TabBarModel) RenderCompact(width int) string {
	var parts []string

	for _, tab := range m.tabs {
		if tab.ID == m.active {
			label := fmt.Sprintf("[%s]", tab.Shortcut)
			parts = append(parts, activeTabStyle.Render(label))
		} else {
			label := fmt.Sprintf(" %s ", tab.Shortcut)
			parts = append(parts, inactiveTabStyle.Render(label))
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	// Pad to width
	contentWidth := lipgloss.Width(content)
	if contentWidth < width {
		padding := strings.Repeat(" ", width-contentWidth)
		content = content + padding
	}

	return tabBarContainerStyle.Width(width).Render(content)
}
