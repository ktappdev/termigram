---
title: TUI Guide
---

# TUI Guide

This guide summarizes the terminal UI concepts and keyboard interactions described in the project documentation.

## Navigation basics

- Use arrow keys to move through chats and messages
- Press `Enter` to open the selected chat or send a message
- Press `Esc` to go back or cancel the current action

## Keyboard shortcuts

### Navigation

| Shortcut | Action |
|----------|--------|
| `↑` / `↓` | Navigate chats or scroll messages |
| `Enter` | Open selected chat or send message |
| `Esc` | Return to chat list or cancel reply |
| `←` / `→` | Move between panels in desktop view |
| `Home` | Jump to first message |
| `End` | Jump to latest message |

### Actions

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Quit application from the chat list |
| `Ctrl+N` | Start a new conversation |
| `Ctrl+F` | Focus search |
| `Ctrl+Enter` | Insert a newline |
| `Ctrl+\` | Toggle sidebar collapse |
| `/` | Focus message input |
| `?` | Show help |
| `Tab` | Open attachment menu |

### Message actions

| Shortcut | Action |
|----------|--------|
| `R` | Reply to a message |
| `F` | Forward a message |
| `D` | Delete a message |
| `Ctrl+C` | Copy message text |

## Mouse support

- Click a chat to select it
- Use the scroll wheel to move through lists and messages
- Click the input area to focus message entry
- Click modal buttons when available

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

- minimum width: about 40 columns
- recommended width: 80 or more columns
- minimum height: about 20 rows

## Visual cues

The UI documentation references:

- connection state indicators
- read and delivery markers
- online presence cues
- typing indicators
- unread badges

## Practical tips

- Use a wider terminal for the best split-pane experience
- Use search to narrow long chat lists quickly
- Use interactive mode first if you have not authenticated yet
- Prefer the CLI reference for scripting-oriented command details
