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
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MoshPitCodes/reposync/internal/github"
)

// TemplateTreeModel manages the tree browser for template files.
type TemplateTreeModel struct {
	// Root of the tree
	root *TemplateTreeNode

	// Flattened list of visible nodes for rendering
	flatNodes []*TemplateTreeNode

	// Current cursor position
	cursor int

	// Viewport offset for scrolling
	viewportOffset int

	// Dimensions
	width  int
	height int

	// Template info for display
	templateName string
	templateBranch string
	isLocal bool
}

// NewTemplateTreeModel creates a new tree browser model from a tree response.
func NewTemplateTreeModel(treeResp *github.TreeResponse, templateName, branch string) *TemplateTreeModel {
	root := buildTreeFromResponse(treeResp)

	m := &TemplateTreeModel{
		root:           root,
		cursor:         0,
		viewportOffset: 0,
		width:          60,
		height:         20,
		templateName:   templateName,
		templateBranch: branch,
		isLocal:        false,
	}

	m.flattenTree()
	m.selectAll() // Default: all files selected

	return m
}

// NewTemplateTreeModelFromLocal creates a tree browser model from a local directory.
func NewTemplateTreeModelFromLocal(root *TemplateTreeNode, localPath string) *TemplateTreeModel {
	m := &TemplateTreeModel{
		root:           root,
		cursor:         0,
		viewportOffset: 0,
		width:          60,
		height:         20,
		templateName:   filepath.Base(localPath),
		templateBranch: "",
		isLocal:        true,
	}

	m.flattenTree()
	m.selectAll() // Default: all files selected

	return m
}

// buildTreeFromResponse converts a GitHub tree response to our tree structure.
func buildTreeFromResponse(resp *github.TreeResponse) *TemplateTreeNode {
	root := &TemplateTreeNode{
		Path:     "",
		Name:     "/",
		IsDir:    true,
		Expanded: true,
		Selected: false,
		Children: make([]*TemplateTreeNode, 0),
	}

	// Build a map for easy parent lookup
	nodeMap := make(map[string]*TemplateTreeNode)
	nodeMap[""] = root

	// Sort entries by path for consistent ordering
	entries := resp.Entries
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	for _, entry := range entries {
		node := &TemplateTreeNode{
			Path:     entry.Path,
			Name:     filepath.Base(entry.Path),
			IsDir:    entry.Type == "tree",
			SHA:      entry.SHA,
			Size:     entry.Size,
			Expanded: false,
			Selected: false,
			Children: make([]*TemplateTreeNode, 0),
		}

		// Find parent
		parentPath := filepath.Dir(entry.Path)
		if parentPath == "." {
			parentPath = ""
		}

		parent, ok := nodeMap[parentPath]
		if !ok {
			// Parent doesn't exist yet, create intermediate directories
			parent = ensureParentExists(root, nodeMap, parentPath)
		}

		parent.Children = append(parent.Children, node)
		nodeMap[entry.Path] = node
	}

	// Sort children of each node
	sortChildren(root)

	return root
}

// ensureParentExists creates parent directories as needed.
func ensureParentExists(root *TemplateTreeNode, nodeMap map[string]*TemplateTreeNode, path string) *TemplateTreeNode {
	if path == "" {
		return root
	}

	if node, ok := nodeMap[path]; ok {
		return node
	}

	// Create this node
	parentPath := filepath.Dir(path)
	if parentPath == "." {
		parentPath = ""
	}

	parent := ensureParentExists(root, nodeMap, parentPath)

	node := &TemplateTreeNode{
		Path:     path,
		Name:     filepath.Base(path),
		IsDir:    true,
		Expanded: false,
		Selected: false,
		Children: make([]*TemplateTreeNode, 0),
	}

	parent.Children = append(parent.Children, node)
	nodeMap[path] = node

	return node
}

// sortChildren recursively sorts children (directories first, then alphabetically).
func sortChildren(node *TemplateTreeNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		// Directories come first
		if node.Children[i].IsDir != node.Children[j].IsDir {
			return node.Children[i].IsDir
		}
		return node.Children[i].Name < node.Children[j].Name
	})

	for _, child := range node.Children {
		sortChildren(child)
	}
}

// SetSize sets the dimensions of the tree browser.
func (m *TemplateTreeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// flattenTree rebuilds the flat list of visible nodes.
func (m *TemplateTreeModel) flattenTree() {
	m.flatNodes = make([]*TemplateTreeNode, 0)
	m.flattenNode(m.root, 0)
}

// flattenNode recursively adds visible nodes to the flat list.
func (m *TemplateTreeModel) flattenNode(node *TemplateTreeNode, depth int) {
	// Skip the root node itself
	if node.Path != "" {
		m.flatNodes = append(m.flatNodes, node)
	}

	if node.IsDir && (node.Path == "" || node.Expanded) {
		for _, child := range node.Children {
			m.flattenNode(child, depth+1)
		}
	}
}

// getDepth returns the depth of a node in the tree.
func (m *TemplateTreeModel) getDepth(node *TemplateTreeNode) int {
	if node.Path == "" {
		return 0
	}
	return strings.Count(node.Path, "/") + 1
}

// Update handles messages for the tree browser.
func (m *TemplateTreeModel) Update(msg tea.Msg) (*TemplateTreeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.flatNodes)-1 {
				m.cursor++
				m.ensureVisible()
			}
			return m, nil

		case "right", "l":
			// Expand directory
			if m.cursor >= 0 && m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				if node.IsDir && !node.Expanded {
					node.Expanded = true
					m.flattenTree()
				}
			}
			return m, nil

		case "left", "h":
			// Collapse directory or go to parent
			if m.cursor >= 0 && m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				if node.IsDir && node.Expanded {
					node.Expanded = false
					m.flattenTree()
				}
			}
			return m, nil

		case " ":
			// Toggle selection
			if m.cursor >= 0 && m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				m.toggleSelect(node)
			}
			return m, nil

		case "a":
			// Select all
			m.selectAll()
			return m, nil

		case "n":
			// Deselect all
			m.deselectAll()
			return m, nil

		case "e":
			// Expand all
			m.expandAll(m.root)
			m.flattenTree()
			return m, nil

		case "c":
			// Collapse all
			m.collapseAll(m.root)
			m.flattenTree()
			return m, nil
		}
	}

	return m, nil
}

// ensureVisible adjusts viewport to keep cursor visible.
func (m *TemplateTreeModel) ensureVisible() {
	// Calculate actual visible lines accounting for:
	// - Header (1 line)
	// - Blank line (1 line)
	// - Selection count (1 line)
	// - Blank line (1 line)
	// - Scroll indicator (1 line if needed)
	// - Blank line (1 line)
	// - Help text (1 line)
	// Total chrome: ~8 lines
	chromeLines := 8
	visibleLines := m.height - chromeLines
	if visibleLines < 1 {
		visibleLines = 1
	}

	if m.cursor < m.viewportOffset {
		m.viewportOffset = m.cursor
	} else if m.cursor >= m.viewportOffset+visibleLines {
		m.viewportOffset = m.cursor - visibleLines + 1
	}
}

// toggleSelect toggles selection for a node and its children if directory.
func (m *TemplateTreeModel) toggleSelect(node *TemplateTreeNode) {
	newState := !node.Selected
	m.setSelectRecursive(node, newState)
}

// setSelectRecursive sets selection state for a node and all children.
func (m *TemplateTreeModel) setSelectRecursive(node *TemplateTreeNode, selected bool) {
	node.Selected = selected
	for _, child := range node.Children {
		m.setSelectRecursive(child, selected)
	}
}

// selectAll selects all nodes.
func (m *TemplateTreeModel) selectAll() {
	m.setSelectRecursive(m.root, true)
}

// deselectAll deselects all nodes.
func (m *TemplateTreeModel) deselectAll() {
	m.setSelectRecursive(m.root, false)
}

// expandAll expands all directories.
func (m *TemplateTreeModel) expandAll(node *TemplateTreeNode) {
	if node.IsDir {
		node.Expanded = true
		for _, child := range node.Children {
			m.expandAll(child)
		}
	}
}

// collapseAll collapses all directories.
func (m *TemplateTreeModel) collapseAll(node *TemplateTreeNode) {
	if node.IsDir && node.Path != "" {
		node.Expanded = false
		for _, child := range node.Children {
			m.collapseAll(child)
		}
	}
}

// GetSelectedPaths returns the paths of all selected files.
func (m *TemplateTreeModel) GetSelectedPaths() []string {
	paths := make([]string, 0)
	m.collectSelectedPaths(m.root, &paths)
	return paths
}

// collectSelectedPaths recursively collects selected file paths.
func (m *TemplateTreeModel) collectSelectedPaths(node *TemplateTreeNode, paths *[]string) {
	// Only include files, not directories
	if !node.IsDir && node.Selected {
		*paths = append(*paths, node.Path)
	}

	for _, child := range node.Children {
		m.collectSelectedPaths(child, paths)
	}
}

// GetSelectedCount returns the count of selected files.
func (m *TemplateTreeModel) GetSelectedCount() int {
	return len(m.GetSelectedPaths())
}

// View renders the tree browser.
func (m *TemplateTreeModel) View() string {
	var b strings.Builder

	// Header
	branchInfo := ""
	if m.templateBranch != "" {
		branchInfo = fmt.Sprintf(" (%s)", m.templateBranch)
	}

	sourceIcon := "🌐"
	if m.isLocal {
		sourceIcon = "📁"
	}

	header := templateTreeHeaderStyle.Render(
		fmt.Sprintf("%s Template: %s%s", sourceIcon, m.templateName, branchInfo),
	)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Selection count
	selectedCount := m.GetSelectedCount()
	totalFiles := m.countFiles(m.root)
	countStr := fmt.Sprintf("Selected: %d/%d files", selectedCount, totalFiles)
	b.WriteString(templateTreeCountStyle.Render(countStr))
	b.WriteString("\n\n")

	// Tree content - must match ensureVisible() calculation
	// Chrome: header(1) + blank(1) + count(1) + blank(1) + scroll(1) + blank(1) + help(1) + padding(1) = 8
	chromeLines := 8
	visibleLines := m.height - chromeLines
	if visibleLines < 1 {
		visibleLines = 5
	}

	startIdx := m.viewportOffset
	endIdx := startIdx + visibleLines
	if endIdx > len(m.flatNodes) {
		endIdx = len(m.flatNodes)
	}

	for i := startIdx; i < endIdx; i++ {
		node := m.flatNodes[i]
		depth := m.getDepth(node)
		indent := strings.Repeat("  ", depth-1)

		// Selection checkbox
		checkbox := "[ ]"
		if node.Selected {
			checkbox = "[✓]"
		}

		// Icon
		icon := "📄"
		if node.IsDir {
			if node.Expanded {
				icon = "📂"
			} else {
				icon = "📁"
			}
		}

		// Build line
		line := fmt.Sprintf("%s%s %s %s", indent, checkbox, icon, node.Name)

		// Apply style
		var style lipgloss.Style
		if i == m.cursor {
			style = templateTreeSelectedStyle
		} else if node.Selected {
			style = templateTreeCheckedStyle
		} else {
			style = templateTreeItemStyle
		}

		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Show scroll indicator if needed
	if len(m.flatNodes) > visibleLines {
		scrollInfo := fmt.Sprintf("(%d-%d of %d)", startIdx+1, endIdx, len(m.flatNodes))
		b.WriteString(templateTreeHintStyle.Render(scrollInfo))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Help text
	helpText := "↑/↓ navigate • space toggle • a all • n none • ←/→ collapse/expand • e/c expand/collapse all • enter continue"
	b.WriteString(templateTreeHelpStyle.Render(helpText))

	return templateTreeStyle.Width(m.width).Render(b.String())
}

// countFiles counts the total number of files in the tree.
func (m *TemplateTreeModel) countFiles(node *TemplateTreeNode) int {
	count := 0
	if !node.IsDir {
		count = 1
	}
	for _, child := range node.Children {
		count += m.countFiles(child)
	}
	return count
}

// Styles for tree browser
var (
	templateTreeStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor)

	templateTreeHeaderStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	templateTreeCountStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	templateTreeItemStyle = lipgloss.NewStyle().
				Foreground(fgColor)

	templateTreeSelectedStyle = lipgloss.NewStyle().
					Foreground(secondaryColor).
					Bold(true).
					Background(bgColor)

	templateTreeCheckedStyle = lipgloss.NewStyle().
				Foreground(successColor)

	templateTreeHintStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

	templateTreeHelpStyle = lipgloss.NewStyle().
				Foreground(mutedColor)
)
