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
)

// TestRenderFooter tests that the footer renders correctly.
func TestRenderFooter(t *testing.T) {
	bindings := []string{
		"↑/↓", "navigate",
		"space", "toggle",
		"a/n", "all/none",
		"/", "search",
	}

	footer := RenderFooter(bindings...)
	if footer == "" {
		t.Error("Footer should not be empty")
	}

	// Check that all bindings are present
	for i := 0; i < len(bindings); i += 2 {
		if !strings.Contains(footer, bindings[i]) {
			t.Errorf("Footer should contain key binding '%s'", bindings[i])
		}
		if !strings.Contains(footer, bindings[i+1]) {
			t.Errorf("Footer should contain description '%s'", bindings[i+1])
		}
	}
}

// TestRenderFooterEmpty tests footer with no bindings.
func TestRenderFooterEmpty(t *testing.T) {
	footer := RenderFooter()
	if footer == "" {
		t.Error("Footer with no bindings should still render (empty styled container)")
	}
}

// TestRenderFooterDoubleRow tests that footer splits bindings into two rows.
func TestRenderFooterDoubleRow(t *testing.T) {
	// Many bindings to ensure two-row layout
	bindings := []string{
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

	footer := RenderFooter(bindings...)
	if footer == "" {
		t.Error("Footer should not be empty")
	}

	// Count newlines to verify two-row layout
	// The footer should contain at least one newline for the double-row layout
	lineCount := strings.Count(footer, "\n")
	if lineCount < 1 {
		t.Errorf("Footer should have at least 2 rows (found %d newlines)", lineCount)
	}

	// Verify all bindings are still present
	for i := 0; i < len(bindings); i += 2 {
		if !strings.Contains(footer, bindings[i]) {
			t.Errorf("Footer should contain key binding '%s'", bindings[i])
		}
		if !strings.Contains(footer, bindings[i+1]) {
			t.Errorf("Footer should contain description '%s'", bindings[i+1])
		}
	}
}

// TestRenderFooterOddBindings tests footer with odd number of bindings.
func TestRenderFooterOddBindings(t *testing.T) {
	// Odd number of elements (incomplete last pair)
	bindings := []string{
		"↑/↓", "navigate",
		"space", "toggle",
		"a/n", // Missing description
	}

	footer := RenderFooter(bindings...)
	if footer == "" {
		t.Error("Footer should not be empty even with incomplete bindings")
	}

	// Complete pairs should be present
	if !strings.Contains(footer, "navigate") {
		t.Error("Footer should contain complete pairs")
	}
	if !strings.Contains(footer, "toggle") {
		t.Error("Footer should contain complete pairs")
	}
}

// TestRenderFooterConsistency tests that footer renders consistently.
func TestRenderFooterConsistency(t *testing.T) {
	bindings := []string{
		"↑/↓", "navigate",
		"space", "toggle",
		"enter", "sync",
		"q", "quit",
	}

	footer1 := RenderFooter(bindings...)
	footer2 := RenderFooter(bindings...)
	footer3 := RenderFooter(bindings...)

	if footer1 != footer2 {
		t.Error("Footer rendering should be consistent across multiple calls")
	}
	if footer2 != footer3 {
		t.Error("Footer rendering should be consistent across multiple calls")
	}
}

// TestRenderSuccess tests success message rendering.
func TestRenderSuccess(t *testing.T) {
	msg := RenderSuccess("Operation completed")
	if !strings.Contains(msg, "Operation completed") {
		t.Error("Success message should contain the text")
	}
	if !strings.Contains(msg, "✓") {
		t.Error("Success message should contain checkmark")
	}
}

// TestRenderError tests error message rendering.
func TestRenderError(t *testing.T) {
	msg := RenderError("Something went wrong")
	if !strings.Contains(msg, "Something went wrong") {
		t.Error("Error message should contain the text")
	}
	if !strings.Contains(msg, "✗") {
		t.Error("Error message should contain cross mark")
	}
}

// TestRenderWarning tests warning message rendering.
func TestRenderWarning(t *testing.T) {
	msg := RenderWarning("Be careful")
	if !strings.Contains(msg, "Be careful") {
		t.Error("Warning message should contain the text")
	}
	if !strings.Contains(msg, "⚠") {
		t.Error("Warning message should contain warning symbol")
	}
}

// TestRenderInfo tests info message rendering.
func TestRenderInfo(t *testing.T) {
	msg := RenderInfo("For your information")
	if !strings.Contains(msg, "For your information") {
		t.Error("Info message should contain the text")
	}
	if !strings.Contains(msg, "ℹ") {
		t.Error("Info message should contain info symbol")
	}
}

// TestRenderCount tests count rendering.
func TestRenderCount(t *testing.T) {
	count := RenderCount(5, 10)
	if !strings.Contains(count, "5") {
		t.Error("Count should contain selected count")
	}
	if !strings.Contains(count, "10") {
		t.Error("Count should contain total count")
	}
}
