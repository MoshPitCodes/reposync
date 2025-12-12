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

	"github.com/MoshPitCodes/reposync/internal/config"
)

// SettingsField represents a field in the settings form.
type SettingsField struct {
	Label       string
	Key         string
	Value       string
	Placeholder string
	Help        string
}

// SettingsModel manages the settings overlay modal.
type SettingsModel struct {
	// Slices (24 bytes each)
	fields []SettingsField
	inputs []textinput.Model

	// Pointer (8 bytes)
	store *config.ConfigStore

	// Ints (8 bytes each)
	selected int
	width    int
	height   int

	// Bool (1 byte)
	compactMode bool
}

// NewSettingsModel creates a new settings model.
func NewSettingsModel(store *config.ConfigStore) *SettingsModel {
	// Load persisted config
	persistedCfg, err := store.Load()
	if err != nil {
		persistedCfg = &config.PersistedConfig{}
	}

	fields := []SettingsField{
		{
			Label:       "Target Directory",
			Key:         "target_dir",
			Value:       persistedCfg.TargetDir,
			Placeholder: "~/repos",
			Help:        "Default directory where repositories will be cloned",
		},
		{
			Label:       "Source Directories",
			Key:         "source_dirs",
			Value:       strings.Join(persistedCfg.SourceDirs, ":"),
			Placeholder: "/path/to/repos1:/path/to/repos2",
			Help:        "Colon-separated list of directories to scan for local repos",
		},
		{
			Label:       "Default Owner",
			Key:         "default_owner",
			Value:       persistedCfg.DefaultOwner,
			Placeholder: "your-github-username",
			Help:        "Default GitHub user or organization",
		},
	}

	// Create text inputs for each field
	inputs := make([]textinput.Model, len(fields))
	for i, field := range fields {
		ti := textinput.New()
		ti.Placeholder = field.Placeholder
		ti.SetValue(field.Value)
		ti.CharLimit = 200

		if i == 0 {
			ti.Focus()
		}

		inputs[i] = ti
	}

	return &SettingsModel{
		store:       store,
		fields:      fields,
		inputs:      inputs,
		selected:    0,
		compactMode: persistedCfg.CompactMode,
		width:       80,
		height:      30,
	}
}

// Update handles messages for the settings modal.
func (m *SettingsModel) Update(msg tea.Msg) (*SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Close without saving
			return m, func() tea.Msg {
				return SettingsCloseMsg{Save: false}
			}

		case "ctrl+s", "enter":
			// Save and close
			if err := m.Save(); err != nil {
				// TODO: Show error message
				return m, nil
			}
			return m, func() tea.Msg {
				return SettingsCloseMsg{Save: true}
			}

		case "up", "k", "shift+tab":
			if m.selected > 0 {
				m.inputs[m.selected].Blur()
				m.selected--
				m.inputs[m.selected].Focus()
			}
			return m, nil

		case "down", "j", "tab":
			if m.selected < len(m.inputs)-1 {
				m.inputs[m.selected].Blur()
				m.selected++
				m.inputs[m.selected].Focus()
			}
			return m, nil

		case "ctrl+c":
			// Toggle compact mode
			m.compactMode = !m.compactMode
			return m, nil
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	m.inputs[m.selected], cmd = m.inputs[m.selected].Update(msg)

	return m, cmd
}

// Save saves the current settings to the config store.
func (m *SettingsModel) Save() error {
	persistedCfg := &config.PersistedConfig{
		CompactMode: m.compactMode,
	}

	for i, field := range m.fields {
		value := m.inputs[i].Value()
		switch field.Key {
		case "target_dir":
			persistedCfg.TargetDir = value
		case "source_dirs":
			if value != "" {
				persistedCfg.SourceDirs = strings.Split(value, ":")
			}
		case "default_owner":
			persistedCfg.DefaultOwner = value
		}
	}

	return m.store.Save(persistedCfg)
}

// View renders the settings modal.
func (m *SettingsModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(settingsOverlayTitleStyle.Render("Settings"))
	b.WriteString("\n\n")

	// Instructions
	b.WriteString(helpDescStyle.Render("Configure default settings for repo-sync"))
	b.WriteString("\n\n")

	// Fields
	for i, field := range m.fields {
		isFocused := i == m.selected

		// Label
		labelStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
		if isFocused {
			labelStyle = labelStyle.Foreground(secondaryColor)
		}
		b.WriteString(labelStyle.Render(field.Label))
		b.WriteString("\n")

		// Input
		inputView := m.inputs[i].View()
		if isFocused {
			inputView = focusedInputStyle.Render(inputView)
		} else {
			inputView = inputStyle.Render(inputView)
		}
		b.WriteString(inputView)
		b.WriteString("\n")

		// Help text
		helpText := helpDescStyle.Render(field.Help)
		b.WriteString("  " + helpText)
		b.WriteString("\n\n")
	}

	// Compact mode toggle
	compactLabel := "Compact Mode: "
	compactValue := "Off"
	if m.compactMode {
		compactValue = "On"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(compactLabel))
	b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render(compactValue))
	b.WriteString("\n")
	b.WriteString(helpDescStyle.Render("  Press Ctrl+C to toggle compact display mode"))
	b.WriteString("\n\n")

	// Config file location
	b.WriteString(RenderMetadata(fmt.Sprintf("Config file: %s", m.store.Path())))
	b.WriteString("\n\n")

	// Footer with help
	footer := RenderFooter(
		"↑/↓", "navigate",
		"enter", "save",
		"esc", "cancel",
		"ctrl+c", "toggle compact",
	)
	b.WriteString(footer)

	return settingsOverlayStyle.Render(b.String())
}

// SetSize sets the size for the settings modal.
func (m *SettingsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Styles for settings overlay
var (
	settingsOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(primaryColor).
				Padding(2, 3).
				Background(bgColor).
				Foreground(fgColor)

	settingsOverlayTitleStyle = lipgloss.NewStyle().
					Foreground(primaryColor).
					Bold(true).
					Underline(true).
					MarginBottom(1)
)
