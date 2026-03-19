---
title: Features
---

# Features

termigram combines an interactive command-line client with automation-friendly one-shot commands.

## Core CLI features

- Interactive login and chat usage
- One-shot command mode for scripts and bots
- Saved session reuse across runs
- Direct messaging by user id or `@username`
- Contact listing and username prefix search

## Interactive commands

Common interactive commands include:

- `\me`
- `\contacts`
- `\find <prefix>`
- `\msg <id|@username> <text>`
- `\to <id|@username>`
- `\chats`
- `\here`
- `\close`
- `\help`
- `\quit`

## Automation-friendly behavior

One-shot CLI mode supports:

- JSON output for deterministic parsing
- explicit timeouts
- command-oriented usage for shell scripts and AI agents

Examples:

```bash
./termigram me --json
./termigram contacts --json
./termigram send --json @oncall "Nightly job failed"
```

## Terminal UI capabilities

The repository also contains a Bubble Tea-based UI implementation and supporting components.

Highlighted UI capabilities:

- split-pane layout for chats and messages
- chat filtering and search
- unread indicators
- draft preservation
- reply context support
- adaptive layout for smaller terminals
- Telegram-inspired dark color palette

## Responsive layout

### Desktop view

For wider terminals, the UI presents:

- a chat list on the left
- a message view on the right
- an input area below the conversation

### Compact view

For narrower terminals, the UI switches to a single-panel flow that is easier to use over SSH or in split windows.

## Build and release notes

Use the project build target:

```bash
make build
```

The build script injects version metadata using `git describe --tags --dirty` when available and falls back to `dev` when no tag metadata exists.
