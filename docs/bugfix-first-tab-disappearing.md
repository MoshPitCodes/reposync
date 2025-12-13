# Bug Fix: First Tab Disappearing Intermittently

## Issue Description

The first tab (Personal) in the TUI application was disappearing intermittently during rendering. This was a critical UI bug that made the application appear broken and confused users about which tab was active.

## Root Cause Analysis

The issue was caused by **inconsistent width handling** in the tab bar rendering logic. Specifically:

1. **Missing Width Awareness**: The `ViewWithContainer()` method in `/internal/tui/tabs.go` did not set a width on the `tabBarContainerStyle` when rendering the tab bar.

2. **Inconsistent Rendering**: Unlike other UI components (header, owner bar, footer) which properly set their width based on `m.width`, the tab bar relied on the default lipgloss behavior without explicit width constraints.

3. **Terminal Resize Issues**: When the terminal was resized or the view was re-rendered, the tab bar could be cut off or improperly rendered, causing the first tab to disappear.

### Problematic Code

```go
// Original ViewWithContainer method - NO WIDTH SET
func (m *TabBarModel) ViewWithContainer() string {
    return tabBarContainerStyle.Render(m.View())
}

// tabBarContainerStyle definition - NO WIDTH
tabBarContainerStyle = lipgloss.NewStyle().
    BorderStyle(lipgloss.NormalBorder()).
    BorderBottom(true).
    BorderForeground(borderColor).
    MarginBottom(1)
```

### Comparison with Working Code

The `RenderCompact()` method correctly set the width:

```go
func (m *TabBarModel) RenderCompact(width int) string {
    // ... code ...
    return tabBarContainerStyle.Width(width).Render(content)  // ✓ Width set!
}
```

## Solution

The fix involved three changes to ensure consistent, width-aware tab rendering:

### 1. Added `ViewWithWidth()` Method

Created a new method that renders tabs with padding to ensure consistent width:

```go
// ViewWithWidth renders the tab bar with a specified width.
func (m *TabBarModel) ViewWithWidth(width int) string {
    var tabs []string

    for _, tab := range m.tabs {
        isActive := tab.ID == m.active
        tabs = append(tabs, m.renderTab(tab, isActive))
    }

    content := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

    // Pad to width if needed to ensure consistent rendering
    contentWidth := lipgloss.Width(content)
    if width > 0 && contentWidth < width {
        padding := strings.Repeat(" ", width-contentWidth)
        content = content + padding
    }

    return content
}
```

### 2. Added `ViewWithContainerAndWidth()` Method

Created a new method that combines width-aware rendering with the container border:

```go
// ViewWithContainerAndWidth renders the tab bar with a container border and specified width.
func (m *TabBarModel) ViewWithContainerAndWidth(width int) string {
    content := m.ViewWithWidth(width)
    if width > 0 {
        return tabBarContainerStyle.Width(width).Render(content)
    }
    return tabBarContainerStyle.Render(content)
}
```

### 3. Updated `renderTabs()` in view.go

Modified the view rendering to use the width-aware method:

```go
// renderTabs renders the tab bar.
func (m Model) renderTabs() string {
    if m.width > 0 {
        return m.tabs.ViewWithContainerAndWidth(m.width)
    }
    return m.tabs.ViewWithContainer()
}
```

## Benefits of the Fix

1. **Consistent Rendering**: Tab bar now renders consistently across terminal resizes and view updates
2. **Proper Width Handling**: Tab bar respects the terminal width like other UI components
3. **No Visual Glitches**: First tab no longer disappears intermittently
4. **Backward Compatible**: Original methods (`View()`, `ViewWithContainer()`) are preserved for compatibility
5. **Testable**: Added comprehensive unit tests to verify the fix

## Test Coverage

Added comprehensive tests in `/internal/tui/tabs_test.go`:

- `TestTabBarRendering`: Verifies all tabs are present in basic rendering
- `TestTabBarRenderingWithWidth`: Tests width-aware rendering
- `TestTabBarRenderingWithContainer`: Tests container rendering with width
- `TestTabBarActiveTab`: Verifies active tab selection
- `TestTabBarNavigation`: Tests tab navigation (Next/Prev)
- `TestTabBarConsistentRendering`: Ensures consistent rendering across multiple calls
- `TestTabBarDifferentWidths`: Tests rendering at different terminal widths

All tests pass successfully.

## Files Modified

1. `/internal/tui/tabs.go`:
   - Added `ViewWithWidth()` method
   - Added `ViewWithContainerAndWidth()` method

2. `/internal/tui/view.go`:
   - Updated `renderTabs()` to use width-aware rendering

3. `/internal/tui/tabs_test.go`:
   - Created comprehensive test suite

## Impact

- **User-Facing**: The first tab will no longer disappear, providing a consistent and reliable UI experience
- **Developer-Facing**: The tab rendering logic is now more robust and testable
- **Performance**: Minimal performance impact (only adds string padding when needed)
- **Compatibility**: Fully backward compatible with existing code

## Related Issues

This fix also addresses potential issues with:
- Terminal resize handling
- Different terminal emulators with varying width calculation
- ANSI color code handling in width calculations

## Verification

To verify the fix:

1. Build the application: `go build .`
2. Run the application: `./reposync`
3. Switch between tabs using `1`, `2`, `3` or `Tab`/`Shift+Tab`
4. Resize the terminal window
5. Verify all three tabs (Personal, Organizations, Local) remain visible

The first tab should now consistently appear regardless of terminal operations.

## Prevention

To prevent similar issues in the future:

1. Always set explicit widths on lipgloss components that are part of the main view
2. Ensure consistency between different rendering methods (e.g., `View()`, `RenderCompact()`)
3. Test rendering at different terminal widths
4. Add unit tests for UI components to catch rendering regressions
