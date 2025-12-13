# Tab Disappearing Bug - Root Cause Analysis and Fix

## Issue Description
The first tab (Personal) in the TUI application was intermittently disappearing when switching between tabs or during initial render.

## Root Cause Analysis

### Primary Issue: Width Calculation Overflow
The bug was caused by a **width calculation overflow** in the tab bar rendering logic in `/internal/tui/tabs.go`.

#### The Problem Flow:

1. **Width Padding Calculation** (`ViewWithWidth` function, lines 163-186):
   - The function calculates `contentWidth` using `lipgloss.Width(content)`
   - It then adds padding: `padding := strings.Repeat(" ", width-contentWidth)`
   - **BUG**: This calculation didn't account for the container's internal rendering behavior

2. **Container Width Application** (`ViewWithContainerAndWidth` function, lines 236-244):
   - The container style applies `.Width(width)` which includes borders and padding
   - When the content width + padding equals exactly the terminal width, lipgloss can truncate content from the left side
   - This truncation caused the first tab to disappear

3. **Tab Style Margins** (lines 205-216):
   - Both `activeTabStyle` and `inactiveTabStyle` have `.MarginRight(1)`
   - These margins are applied OUTSIDE the tab content
   - The `lipgloss.Width()` function includes these margins, but the padding calculation didn't account for them properly

4. **Initial Render Race Condition**:
   - On startup, `m.width` and `m.height` are initialized to `0`
   - The first `WindowSizeMsg` sets these values
   - If rendering occurs before `WindowSizeMsg`, the fallback logic could produce inconsistent results

### Secondary Issue: Inconsistent Fallback Logic
The `renderTabs()` function in `/internal/tui/view.go` had conditional logic:
```go
if m.width > 0 {
    return m.tabs.ViewWithContainerAndWidth(m.width)
}
return m.tabs.ViewWithContainer()
```
This created two different rendering paths that could produce different results.

## The Fix

### 1. Added Safety Margin to Width Calculation
**File**: `/internal/tui/tabs.go`, lines 163-186

**Changes**:
- Added a 2-character safety margin when calculating padding
- This prevents the content from exactly matching the container width
- Prevents lipgloss from truncating content

```go
// Only add padding if we have a valid width and content is smaller
// Account for a safety margin to prevent overflow
contentWidth := lipgloss.Width(content)
if width > 0 && contentWidth < width {
    // Add safety margin to prevent overflow issues
    safeWidth := width - 2
    if contentWidth < safeWidth {
        padding := strings.Repeat(" ", safeWidth-contentWidth)
        content = content + padding
    }
}
```

### 2. Added Minimum Width Protection
**File**: `/internal/tui/tabs.go`, lines 236-244

**Changes**:
- Added validation to prevent rendering with invalid widths
- If width is ≤0 or <50, fall back to simple rendering
- This protects against edge cases during initial render

```go
// Don't render with width if it's too small or zero
if width <= 0 || width < 50 {
    return tabBarContainerStyle.Render(m.View())
}
```

### 3. Unified Rendering Path
**File**: `/internal/tui/view.go`, lines 107-111

**Changes**:
- Simplified `renderTabs()` to always use `ViewWithContainerAndWidth`
- The function now handles all edge cases internally
- Eliminates inconsistent behavior from having two rendering paths

```go
// Always use ViewWithContainerAndWidth for consistency
// It will fall back to ViewWithContainer if width is invalid
return m.tabs.ViewWithContainerAndWidth(m.width)
```

## Testing

All existing tests pass:
```
=== RUN   TestTabBarRendering
--- PASS: TestTabBarRendering (0.00s)
=== RUN   TestTabBarRenderingWithWidth
--- PASS: TestTabBarRenderingWithWidth (0.00s)
=== RUN   TestTabBarRenderingWithContainer
--- PASS: TestTabBarRenderingWithContainer (0.00s)
=== RUN   TestTabBarActiveTab
--- PASS: TestTabBarActiveTab (0.00s)
=== RUN   TestTabBarNavigation
--- PASS: TestTabBarNavigation (0.00s)
=== RUN   TestTabBarConsistentRendering
--- PASS: TestTabBarConsistentRendering (0.00s)
=== RUN   TestTabBarDifferentWidths
--- PASS: TestTabBarDifferentWidths (0.00s)
```

## Expected Behavior After Fix

1. ✅ Tabs render consistently across all terminal widths
2. ✅ First tab no longer disappears when switching tabs
3. ✅ Initial render (before WindowSizeMsg) displays correctly
4. ✅ No overflow or truncation issues
5. ✅ Tab bar adapts gracefully to narrow terminals (width < 50)

## Files Modified

1. `/internal/tui/tabs.go` - Core tab rendering logic
2. `/internal/tui/view.go` - View composition logic

## Related Issues

- Previously attempted fix: Added width-aware rendering methods (incomplete)
- Archived repositories feature: No conflict with this fix

## Prevention

To prevent similar issues in the future:

1. Always account for container borders/padding when calculating widths
2. Add safety margins to width calculations involving lipgloss
3. Validate width values before applying them
4. Use consistent rendering paths (avoid conditional logic that creates divergent code paths)
5. Test with various terminal widths, especially edge cases (very narrow, zero width)
