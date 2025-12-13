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
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestTabBarRendering tests that all tabs are properly rendered.
func TestTabBarRendering(t *testing.T) {
	tabBar := NewTabBarModel()

	// Test rendering without width
	view := tabBar.View()
	if view == "" {
		t.Error("Tab bar view should not be empty")
	}

	// Check that all tabs are present
	if !strings.Contains(view, "Personal") {
		t.Error("Tab bar should contain 'Personal' tab")
	}
	if !strings.Contains(view, "Orgs") {
		t.Error("Tab bar should contain 'Orgs' tab")
	}
	if !strings.Contains(view, "Local") {
		t.Error("Tab bar should contain 'Local' tab")
	}
}

// TestTabBarRenderingWithWidth tests that tab bar renders correctly with width.
func TestTabBarRenderingWithWidth(t *testing.T) {
	tabBar := NewTabBarModel()
	width := 80

	// Test rendering with width
	view := tabBar.ViewWithWidth(width)
	if view == "" {
		t.Error("Tab bar view with width should not be empty")
	}

	// Check that the view has appropriate length (considering ANSI codes)
	viewWidth := lipgloss.Width(view)
	if viewWidth < width/2 {
		t.Errorf("Tab bar width seems too small: %d (expected around %d)", viewWidth, width)
	}

	// Check that all tabs are present
	if !strings.Contains(view, "Personal") {
		t.Error("Tab bar should contain 'Personal' tab")
	}
	if !strings.Contains(view, "Orgs") {
		t.Error("Tab bar should contain 'Orgs' tab")
	}
	if !strings.Contains(view, "Local") {
		t.Error("Tab bar should contain 'Local' tab")
	}
}

// TestTabBarRenderingWithContainer tests rendering with container and width.
func TestTabBarRenderingWithContainer(t *testing.T) {
	tabBar := NewTabBarModel()
	width := 100

	// Test rendering with container and width
	view := tabBar.ViewWithContainerAndWidth(width)
	if view == "" {
		t.Error("Tab bar view with container and width should not be empty")
	}

	// The view should include the container styling
	// We can't check exact width due to borders, but it should be present
	viewWidth := lipgloss.Width(view)
	if viewWidth == 0 {
		t.Error("Tab bar view with container should have non-zero width")
	}
}

// TestTabBarActiveTab tests that the active tab is properly set and rendered.
func TestTabBarActiveTab(t *testing.T) {
	tabBar := NewTabBarModel()

	// Initial active tab should be Personal
	if tabBar.GetActive() != ModePersonal {
		t.Error("Initial active tab should be Personal")
	}

	// Switch to Organizations tab
	tabBar.SetActive(ModeOrganization)
	if tabBar.GetActive() != ModeOrganization {
		t.Error("Active tab should be Organizations after SetActive")
	}

	view := tabBar.View()
	if !strings.Contains(view, "Orgs") {
		t.Error("View should contain Organizations tab")
	}

	// Switch to Local tab
	tabBar.SetActive(ModeLocal)
	if tabBar.GetActive() != ModeLocal {
		t.Error("Active tab should be Local after SetActive")
	}
}

// TestTabBarNavigation tests tab navigation (Next/Prev).
func TestTabBarNavigation(t *testing.T) {
	tabBar := NewTabBarModel()

	// Start at Personal (0)
	if tabBar.GetActive() != ModePersonal {
		t.Error("Should start at Personal tab")
	}

	// Next should go to Organizations (1)
	nextMode := tabBar.Next()
	if nextMode != ModeOrganization {
		t.Error("Next from Personal should go to Organizations")
	}

	// Next should go to Local (2)
	nextMode = tabBar.Next()
	if nextMode != ModeLocal {
		t.Error("Next from Organizations should go to Local")
	}

	// Next should wrap to Personal (0)
	nextMode = tabBar.Next()
	if nextMode != ModePersonal {
		t.Error("Next from Local should wrap to Personal")
	}

	// Prev should go to Local (2)
	prevMode := tabBar.Prev()
	if prevMode != ModeLocal {
		t.Error("Prev from Personal should wrap to Local")
	}

	// Prev should go to Organizations (1)
	prevMode = tabBar.Prev()
	if prevMode != ModeOrganization {
		t.Error("Prev from Local should go to Organizations")
	}
}

// TestTabBarConsistentRendering tests that tab bar renders consistently
// across multiple calls with the same width.
func TestTabBarConsistentRendering(t *testing.T) {
	tabBar := NewTabBarModel()
	width := 100

	// Render multiple times
	view1 := tabBar.ViewWithContainerAndWidth(width)
	view2 := tabBar.ViewWithContainerAndWidth(width)
	view3 := tabBar.ViewWithContainerAndWidth(width)

	// All renders should be identical
	if view1 != view2 {
		t.Error("First and second render should be identical")
	}
	if view2 != view3 {
		t.Error("Second and third render should be identical")
	}

	// All tabs should be present in every render
	for i, view := range []string{view1, view2, view3} {
		if !strings.Contains(view, "Personal") {
			t.Errorf("Render %d should contain 'Personal' tab", i+1)
		}
		if !strings.Contains(view, "Orgs") {
			t.Errorf("Render %d should contain 'Orgs' tab", i+1)
		}
		if !strings.Contains(view, "Local") {
			t.Errorf("Render %d should contain 'Local' tab", i+1)
		}
	}
}

// TestTabBarDifferentWidths tests rendering at different terminal widths.
func TestTabBarDifferentWidths(t *testing.T) {
	tabBar := NewTabBarModel()

	widths := []int{50, 80, 100, 120, 150}

	for _, width := range widths {
		view := tabBar.ViewWithWidth(width)
		if view == "" {
			t.Errorf("Tab bar should not be empty at width %d", width)
		}

		// All tabs should be present regardless of width
		if !strings.Contains(view, "Personal") {
			t.Errorf("Tab bar at width %d should contain 'Personal' tab", width)
		}
		if !strings.Contains(view, "Orgs") {
			t.Errorf("Tab bar at width %d should contain 'Orgs' tab", width)
		}
		if !strings.Contains(view, "Local") {
			t.Errorf("Tab bar at width %d should contain 'Local' tab", width)
		}
	}
}

// TestTabBarFirstTabAtVariousWidths tests the first tab is visible at all widths.
// This regression test ensures the first tab doesn't disappear at specific widths.
func TestTabBarFirstTabAtVariousWidths(t *testing.T) {
	tabBar := NewTabBarModel()
	tabBar.SetActive(ModePersonal) // Ensure first tab is active

	// Test a wide range of widths, including edge cases
	widths := []int{50, 60, 70, 75, 80, 85, 90, 95, 100, 110, 120, 130, 140, 150, 160}

	for _, width := range widths {
		// Test ViewWithWidth
		view := tabBar.ViewWithWidth(width)
		if !strings.Contains(view, "Personal") {
			t.Errorf("First tab (Personal) missing at width %d in ViewWithWidth", width)
		}

		// Test ViewWithContainerAndWidth
		containerView := tabBar.ViewWithContainerAndWidth(width)
		if !strings.Contains(containerView, "Personal") {
			t.Errorf("First tab (Personal) missing at width %d in ViewWithContainerAndWidth", width)
		}

		// Verify all tabs are still present
		if !strings.Contains(containerView, "Orgs") {
			t.Errorf("Orgs tab missing at width %d", width)
		}
		if !strings.Contains(containerView, "Local") {
			t.Errorf("Local tab missing at width %d", width)
		}
	}
}
