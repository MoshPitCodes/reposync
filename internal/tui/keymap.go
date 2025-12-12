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

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application.
type KeyMap struct {
	// Global
	Quit     key.Binding
	Help     key.Binding
	Settings key.Binding
	Escape   key.Binding

	// Tab navigation
	Tab1      key.Binding
	Tab2      key.Binding
	Tab3      key.Binding
	TabNext   key.Binding
	TabPrev   key.Binding

	// List navigation
	Up         key.Binding
	Down       key.Binding
	PageUp     key.Binding
	PageDown   key.Binding

	// Selection
	Select      key.Binding
	SelectAll   key.Binding
	SelectNone  key.Binding

	// Actions
	Search key.Binding
	Sort   key.Binding
	Enter  key.Binding
	Owner  key.Binding
}

// Keys is the global key map for the application.
var Keys = KeyMap{
	// Global
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Settings: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "settings"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close/cancel"),
	),

	// Tab navigation
	Tab1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "personal"),
	),
	Tab2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "organizations"),
	),
	Tab3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "local"),
	),
	TabNext: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	TabPrev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous tab"),
	),

	// List navigation
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "page down"),
	),

	// Selection
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle selection"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select all"),
	),
	SelectNone: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "deselect all"),
	),

	// Actions
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "cycle sort"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "sync/confirm"),
	),
	Owner: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "change owner"),
	),
}

// ShortHelp returns a list of key bindings for short help.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Enter, k.Help, k.Quit}
}

// FullHelp returns a list of key bindings for full help.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Select, k.SelectAll, k.SelectNone},
		{k.Tab1, k.Tab2, k.Tab3, k.TabNext},
		{k.Search, k.Sort, k.Owner, k.Settings},
		{k.Enter, k.Help, k.Escape, k.Quit},
	}
}
