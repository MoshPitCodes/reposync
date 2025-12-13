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

// VisualTestTabBar prints tab bar renderings at various widths for visual inspection.
// This is a helper function for manual testing and debugging.
func VisualTestTabBar() {
	tabBar := NewTabBarModel()
	widths := []int{50, 60, 70, 75, 80, 85, 90, 95, 100, 110, 120}

	fmt.Println("=== Tab Bar Visual Test ===")
	fmt.Println()

	for _, width := range widths {
		fmt.Printf("Width: %d\n", width)
		fmt.Println(strings.Repeat("-", width))

		// Test ViewWithWidth
		view := tabBar.ViewWithWidth(width)
		fmt.Println(view)
		fmt.Printf("Actual width: %d\n", lipgloss.Width(view))

		// Test ViewWithContainerAndWidth
		containerView := tabBar.ViewWithContainerAndWidth(width)
		fmt.Println(containerView)
		fmt.Printf("Actual width: %d\n", lipgloss.Width(containerView))

		// Check for presence of all tabs
		hasPersonal := strings.Contains(containerView, "Personal")
		hasOrgs := strings.Contains(containerView, "Orgs")
		hasLocal := strings.Contains(containerView, "Local")
		fmt.Printf("Tabs present: Personal=%v, Orgs=%v, Local=%v\n", hasPersonal, hasOrgs, hasLocal)

		if !hasPersonal || !hasOrgs || !hasLocal {
			fmt.Println("⚠️  WARNING: Some tabs are missing!")
		}

		fmt.Println()
	}
}

// VisualTestFooter prints footer renderings with different binding counts.
// This is a helper function for manual testing and debugging.
func VisualTestFooter() {
	fmt.Println("=== Footer Visual Test ===")
	fmt.Println()

	// Test with few bindings (should fit in one row)
	fmt.Println("Few bindings (should be 1-2 rows):")
	fewBindings := []string{
		"↑/↓", "navigate",
		"space", "toggle",
	}
	fmt.Println(RenderFooter(fewBindings...))
	fmt.Println()

	// Test with moderate bindings
	fmt.Println("Moderate bindings (should be 2 rows):")
	moderateBindings := []string{
		"↑/↓", "navigate",
		"space", "toggle",
		"a/n", "all/none",
		"/", "search",
		"enter", "sync",
		"q", "quit",
	}
	fmt.Println(RenderFooter(moderateBindings...))
	fmt.Println()

	// Test with many bindings (should be 2 rows, demonstrating split)
	fmt.Println("Many bindings (should be 2 rows):")
	manyBindings := []string{
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
	footer := RenderFooter(manyBindings...)
	fmt.Println(footer)
	fmt.Printf("Row count: %d\n", strings.Count(footer, "\n")+1)
	fmt.Println()
}
