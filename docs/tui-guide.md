---
title: TUI Guide
---

# TUI Guide

This guide summarizes the optional Bubble Tea-managed terminal UI. The default interactive experience is the legacy command/transcript workflow; start termigram with `--ui tui` when you want this split-pane view.

For deeper implementation notes about resize handling and responsive layout decisions, see `/Users/kentaylor/developer/telegram-cli/termigram/docs/bubbletea-resize-research.md`.

## Navigation basics

- Use arrow keys to move through chats and messages
- Press `Enter` to open the selected chat or send a message
- Press `Esc` to go back or cancel the current action
- Press `Tab` or `Shift+Tab` to cycle focus between the sidebar, transcript, and composer

## Keyboard shortcuts

### Navigation

| Shortcut | Action |
|----------|--------|
| `↑` / `↓` | Navigate chats or scroll messages |
| `Enter` | Open selected chat or send message |
| `Esc` | Return to chat list or cancel reply |
| `Tab` / `Shift+Tab` | Cycle focus between chats, transcript, and composer |

### Actions

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Quit application |
| `Ctrl+F` | Focus search |
| `Ctrl+O` | Insert attach token |
| `R` | Start a reply to the selected chat |
| `D` | Delete the latest message/chat in the active panel |
| `?` | Show help |

## Mouse support

- Click a chat to select it
- Use the scroll wheel to move through messages
- Click the transcript to focus it
- Click the input area to focus message entry

## Layout behavior

### Desktop view

For terminals around 60 columns and wider:

- chat list on the left
- message pane on the right
- input area below the message pane

### Mobile view

For smaller terminals:

- single-panel layout
- `Enter` opens a chat
- `Esc` returns to the chat list
- optimized for narrow terminal sessions

## Recommended terminal sizes

- minimum width: about 24 columns
- recommended width: 80 or more columns
- minimum height: about 20 rows

## Visual cues

The UI exposes:

- connection state indicators
- read and delivery markers
- online presence cues
- draft indicators
- unread badges

## Practical tips

- Use a wider terminal for the best split-pane experience
- Use search to narrow long chat lists quickly
- Use interactive mode first if you have not authenticated yet
- Prefer the CLI reference for scripting-oriented command details
