---
title: Quickstart
---

# Quickstart

This is the fastest path to getting termigram running.

## Install

### Option A: install with Go

```bash
go install github.com/ktappdev/termigram@latest
```

### Option B: build from a local clone

```bash
make build
```

## Run termigram

```bash
./termigram
```

On first launch:

1. Enter your phone number
2. Enter the verification code from Telegram
3. Start chatting

termigram saves your session automatically and reuses it on later runs.

## Send your first message

### Interactive flow

Open a chat picker:

```text
\chats
```

Or send directly:

```text
\msg @username Hello!
```

### One-shot CLI flow

```bash
./termigram send --json @username "Hello from termigram"
```

## Authentication prerequisite for automation

Before one-shot commands work reliably in scripts, complete one interactive login first:

```bash
./termigram
```

This creates or reuses the local session file.

## Common commands

- `./termigram me --json`
- `./termigram contacts --json`
- `./termigram find --json ke`
- `./termigram get --json --limit 5 @username`
- `./termigram send --json @username "Hello"`

## Session storage

By default, the session is stored at:

```text
~/.termigram/session.json
```

To override it:

```bash
export TELEGRAM_SESSION_PATH=/custom/path/session.json
```

## Credential overrides

Use environment variables when you want to override configured credentials:

```bash
export TELEGRAM_APP_ID=your_app_id
export TELEGRAM_APP_HASH=your_app_hash
```

You can also copy `config.json.example` to `config.json` next to the executable and fill in your own values.

## Troubleshooting checklist

- Make sure you completed at least one interactive login
- Put flags before positional arguments in one-shot mode
- Use `--json` for scripts and automation
- Use `--timeout` in automated contexts to avoid hanging commands
