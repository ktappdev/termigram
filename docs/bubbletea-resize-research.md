---
title: Bubble Tea resize research for chat UI
description: Practical research notes for making the Bubble Tea and Lip Gloss split-pane chat UI resize cleanly in termigram.
---

# Bubble Tea resize research for chat UI

This note summarizes practical guidance for making a Bubble Tea + Lip Gloss chat-style TUI survive terminal resizes cleanly.

The focus is termigram’s split-pane layout: sidebar, message viewport, and input area. The goal is to avoid broken wrapping, clipped panels, background artifacts, and stale viewport content after terminal size changes.

## Why resizing breaks chat TUIs

Bubble Tea applications do not become responsive automatically. A chat UI must treat terminal dimensions as part of model state and rebuild layout whenever the terminal changes size.

The main failure modes are consistent across examples and issue threads:

- storing static widths or heights once and never updating them
- rendering panes without subtracting padding, borders, and margins
- resizing the viewport container but not reflowing its content
- wrapping messages against an old width
- assuming one layout works for both narrow and wide terminals
- painting backgrounds or full-width blocks in ways that leave artifacts on resize

## Core concept: handle `tea.WindowSizeMsg` every time

Bubble Tea apps should handle `tea.WindowSizeMsg` on startup and on every terminal resize. The current width and height should live in the model, and all derived layout dimensions should be recomputed from that state.

Recommended pattern:

1. Store `width` and `height` in the main model.
2. In `Update`, handle `tea.WindowSizeMsg`.
3. Save the new dimensions.
4. Recompute all pane sizes from those dimensions.
5. Rewrap text and rebuild viewport content if line layout depends on width.
6. Update child components such as `viewport` with their new sizes.

For a split-pane chat UI, that means a resize event should trigger a full layout pass, not just a field assignment.

## Lip Gloss sizing should be dynamic

Lip Gloss gives the primitives needed to compute layout correctly, but the application must use them consistently.

Useful APIs and concepts from the package docs:

- `Style.Width()` and `Style.Height()` for explicit box sizing
- `Style.MaxWidth()` when content should shrink responsively instead of forcing a fixed width
- `lipgloss.Width()`, `lipgloss.Height()`, and `lipgloss.Size()` to measure rendered content
- frame size accounting for borders, margins, and padding

In practice, pane sizing should be calculated in two stages:

1. Compute available terminal space.
2. Subtract each pane style’s frame size before rendering content inside that pane.

This is especially important for:

- bordered sidebars
- message panes with horizontal padding
- input composers with top borders or surrounding chrome
- chat bubbles that use padding and rounded borders

If the app computes a pane width of 40 columns but then renders a style with extra frame width, the visual result will overflow or clip.

## Split-pane chat layouts must recompute every region

For a chat-style TUI, resizing is not just “change viewport width.” The entire composition should be recalculated.

A practical layout model is:

- total terminal width and height
- sidebar width
- main content width
- header height
- message viewport height
- input composer height
- optional status/footer height

Recommended approach:

- use a sidebar width formula with minimum and maximum bounds
- enforce a minimum main content width
- switch to a compact or single-pane mode below a breakpoint
- avoid hard-coded dimensions except for minimums and fixed chrome heights

Examples of good layout rules:

- sidebar width is `max(minSidebar, min(maxSidebar, terminalWidth/4))`
- if the remaining content width drops below a safe threshold, collapse or hide the sidebar
- input height may be content-driven, but viewport height must be recomputed after accounting for the input and header

This is the difference between a UI that degrades gracefully and one that becomes unreadable in a narrow terminal.

## Viewport resizing requires both dimension updates and content reflow

The Bubble Tea `viewport` component is central for a scrolling message list, but it will not fix wrapping or stale content by itself.

When the pane width changes, do all of the following:

- call `viewport.SetWidth(newWidth)`
- call `viewport.SetHeight(newHeight)`
- rebuild the rendered message content using the new wrapping width
- call `viewport.SetContent(reflowedContent)`

If only the viewport dimensions are updated, message lines may still reflect the previous width. That causes:

- truncated or ragged bubbles
- incorrect scroll offsets
- empty horizontal gaps
- visual corruption after expanding or shrinking the terminal

### Preserve scroll position thoughtfully

When possible, preserve the user’s reading position during resize.

Typical chat expectations:

- if the user is at the bottom, keep them pinned to the bottom after reflow
- if the user is reading older messages, preserve approximate position instead of snapping to latest
- if new wrapping changes line counts, preserve the nearest meaningful anchor rather than raw line offset alone

The exact strategy depends on implementation, but the key takeaway is that resize handling should consider scroll behavior, not just dimensions.

## Wrapping should use the current pane width

Message wrapping must be based on the actual inner width of the message pane or bubble container at render time.

For a chat UI, that usually means:

- compute main pane width
- subtract pane frame width
- subtract bubble padding and borders
- apply wrapping to message text using that final inner width

This matters for readability and alignment:

- sent and received bubbles should align consistently after resize
- timestamps and metadata should not force awkward overflow
- long words, links, and code-like text should degrade as predictably as possible
- narrow layouts may require simpler bubble styling or less horizontal chrome

The most important rule is to never wrap against a stale width.

## Common pitfalls seen in Bubble Tea and Lip Gloss UIs

### 1. Static dimensions captured once

A frequent bug pattern is reading the initial terminal size and then reusing those values forever. This works until the user resizes the terminal, then panels overlap or leave empty space.

### 2. Forgetting frame size

Borders and padding consume columns and rows. If those are not subtracted from the available space, content spills outside its container or appears clipped.

### 3. Updating viewport size without rebuilding content

This is one of the biggest causes of broken scrolling and wrapping in message panes. The viewport dimensions change, but content still reflects the old width.

### 4. Hard-coded pane widths

A fixed-width sidebar or message area may look fine on one terminal size but fail badly on another. Use minimums, maximums, and breakpoints instead.

### 5. Background and overflow artifacts

Rendered backgrounds and full-width styles can leave visual artifacts when the terminal shrinks or when content no longer fills prior dimensions. This shows up as stale blocks of color or partially repainted rows.

### 6. No compact-mode breakpoint

Split-pane layouts eventually become too narrow to be useful. A responsive chat UI should switch to a compact mode instead of squeezing every pane indefinitely.

### 7. Input height not included in total layout math

If the input grows to multiple lines, the viewport height must shrink accordingly. Otherwise the input and message list will overlap or the bottom rows will render incorrectly.

## Practical recommendations for termigram

For termigram’s chat-style TUI, the safest implementation direction is:

1. Treat terminal width and height as first-class model state.
2. Create a single layout recomputation function that derives all pane sizes.
3. Keep style frame accounting explicit for every major pane.
4. Rebuild wrapped message output whenever message pane width changes.
5. Resize and repopulate the message viewport after every layout reflow.
6. Add a compact-mode breakpoint for narrow terminals.
7. Prefer minimum widths and adaptive proportions over constants.
8. Preserve bottom pinning for the message viewport when the user is already at latest.
9. Test resizing in both directions repeatedly, especially around breakpoint thresholds.
10. Verify multi-line input, long messages, and bordered styles together.

## Recommended implementation pattern

A practical update flow for a split-pane Bubble Tea model looks like this:

### On `tea.WindowSizeMsg`

- save `msg.Width` and `msg.Height`
- compute whether the UI should be in split-pane or compact mode
- derive sidebar width and main pane width
- derive header, viewport, and input heights
- update child component sizes
- re-render message content using the current message width
- refresh the viewport content
- restore the desired scroll position

### During `View()`

- render from current derived dimensions only
- avoid hidden assumptions about default widths
- use Lip Gloss measurement helpers where rendered size is uncertain

### For message rendering

- use a bubble or message renderer that accepts width as input
- compute inner wrap width after subtracting all frame size
- keep metadata alignment tied to the final rendered width, not the raw terminal width

## Short implementation checklist

Use this as a future implementation checklist for the termigram chat UI:

- [ ] Store terminal `width` and `height` in the root Bubble Tea model.
- [ ] Handle `tea.WindowSizeMsg` on startup and every resize.
- [ ] Centralize pane math in one layout recomputation function.
- [ ] Subtract Lip Gloss frame sizes for sidebar, message pane, bubbles, and input.
- [ ] Replace hard-coded widths with min/max bounds and a compact breakpoint.
- [ ] Recompute sidebar, content, viewport, and input dimensions on every resize.
- [ ] Rewrap all visible message content using the current message pane width.
- [ ] Call `viewport.SetWidth`, `viewport.SetHeight`, and `viewport.SetContent` after reflow.
- [ ] Preserve bottom pinning or reader position when resizing.
- [ ] Test narrow, medium, and wide terminals for clipping, artifacts, and stale wrapping.

## Source summary

### 1. Bubble Tea WindowSize handling overview

The Bubble Tea architecture guidance emphasizes handling `tea.WindowSizeMsg` and structuring model updates cleanly. For this task, the key takeaway is that resize handling belongs in the event loop and should drive layout recomputation.

Source:
- https://leg100.github.io/en/posts/building-bubbletea-programs/

### 2. Lip Gloss package docs

The Lip Gloss docs provide the sizing and measurement APIs needed for dynamic layout, including width, height, max width, and content measurement helpers. They also reinforce the importance of style frame accounting.

Source:
- https://pkg.go.dev/github.com/charmbracelet/lipgloss/v2

### 3. Bubble Tea viewport docs

The viewport docs cover the API needed to resize scrollable regions and update their content. For chat UIs, width and height changes should usually be paired with `SetContent` after content reflow.

Source:
- https://pkg.go.dev/github.com/charmbracelet/bubbles/viewport

### 4. Bubble Tea layout and multi-panel example article

The multi-panel layout article is useful for responsive pane composition and breakpoint thinking. The practical lesson is that multi-region TUIs need explicit width allocation logic and should adapt layout as the terminal narrows.

Source:
- https://substack.com/home/post/p-152418733

### 5. Bubble Tea discussion on rendering and background pitfalls

The discussion highlights rendering behavior that can produce visual artifacts, especially when backgrounds and full-width rendering assumptions interact badly with terminal updates.

Source:
- https://github.com/charmbracelet/bubbletea/discussions/544

### 6. Bubble Tea resize issue example

The issue illustrates the class of bugs where resize events are not fully propagated through layout and component state, producing broken rendering after terminal changes.

Source:
- https://github.com/charmbracelet/bubbletea/issues/658

### 7. Optional related artifact issue

This related issue is a useful extra reference for redraw and artifact problems that can appear in table-like or pane-based components after resizing.

Source:
- https://github.com/Everras/bubble-table/issues/121

## Bottom line

To make a Bubble Tea + Lip Gloss chat UI resize cleanly, termigram should treat resize as a full layout and reflow event:

- update model dimensions
- recompute pane sizes
- subtract style frame space correctly
- rebuild wrapped message content
- resize and repopulate the viewport
- preserve user scroll position intentionally

That combination is what prevents the most common resize bugs in split-pane TUIs.
