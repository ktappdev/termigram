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

The interactive experience is the command/transcript workflow.

Highlighted capabilities:

- recent chat switching
- unread chat picking
- active-chat transcript redraw on resize
- message bubbles for incoming and outgoing text
- adaptive terminal rendering for narrower widths

## Build and release notes

Use the project build target:

```bash
make build
```

The build script injects version metadata using `git describe --tags --dirty` when available and falls back to `dev` when no tag metadata exists.
