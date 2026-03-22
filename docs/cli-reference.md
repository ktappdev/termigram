---
title: CLI Reference
---

# CLI Reference

termigram supports both interactive use and one-shot command execution.

## General syntax

```bash
./termigram <command> [--json] [--timeout 30s] [command flags] [arguments]
```

## Global usage notes

- Run `./termigram` once interactively before relying on one-shot commands
- Place flags before positional arguments
- Prefer `--json` for scripts and AI integrations

## Commands

### `send`

Send a message to a user id or `@username`.

```bash
./termigram send --json @ken "Hello from automation"
```

### `send-image`

Send a JPG/PNG/WEBP image from a local path, file URL, or HTTP/HTTPS URL.

```bash
./termigram send-image --json @ken ./meme.png "hello"
```

### `get`

Fetch recent messages for a target.

```bash
./termigram get --json --limit 5 @ken
```

Options:

- `--limit N` — number of messages to fetch, default `10`

### `contacts`

List known contacts.

```bash
./termigram contacts --json
```

### `me`

Print information about the current authenticated account.

```bash
./termigram me --json
```

### `find`

Find usernames by cached prefix.

```bash
./termigram find --json ke
```

## Common flags

- `--json` — machine-readable success/data/error envelope
- `--timeout 30s` — command timeout using Go duration syntax

## Examples

### Send an alert from a script

```bash
./termigram send --json @oncall "Job failed: nightly-import"
./termigram send-image --json @ken ./meme.png "hello"
```

### Fetch and parse recent messages

```bash
./termigram get --json --limit 20 @ken | jq '.data.messages[] | {id, from_name, message}'
```

### Resolve current user id

```bash
./termigram me --json | jq '.data.id'
```

## Interactive slash commands

When termigram is running interactively, these commands are available:

- `\me`
- `\contacts`
- `\find <prefix>`
- `\msg <id|@username> <text>`
- `\to <id|@username>`
- `\image <source> [caption]`
- `\openimage [last|message-id|query]`
- `\chats`
- `\here`
- `\close`
- `\help`
- `\quit`

Interactive image tips:

- `\openimage` opens the recent-image picker for the active chat
- `\openimage last` opens the newest image immediately
- `\openimage meme` opens the picker pre-filtered by `meme`
- supported terminals can render visible chat images inline in the transcript flow

Inline preview environment variables:

- `TERMIGRAM_INLINE_IMAGES=auto|on|off`
- `TERMIGRAM_INLINE_IMAGE_PROTOCOL=kitty|iterm2`
- `TERMIGRAM_INLINE_IMAGE_COLS=28`
- `TERMIGRAM_INLINE_IMAGE_ROWS=10`

## Help and version

```bash
./termigram --help
./termigram -h
./termigram --version
./termigram -v
```

Per-command help:

```bash
./termigram send --help
./termigram send-image --help
./termigram get --help
./termigram contacts --help
./termigram me --help
./termigram find --help
```
