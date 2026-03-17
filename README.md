# Modern Telegram CLI

A lightweight Telegram command-line client written in Go using MTProto (`gotd/td`).

## Features

- Interactive mode for login and chat commands
- One-shot CLI mode for scripting
- Reuses a saved user session
- Commands: `send`, `get`, `contacts`, `me`, `find`

## Prerequisites

- Go 1.21+
- Telegram account
- `app_id` and `app_hash` from [my.telegram.org](https://my.telegram.org)

## Configuration

The app loads `config.json` from the same directory as the executable.

Example (`config.json`):

```json
{
  "telegram_app_id": 123456,
  "telegram_app_hash": "your_telegram_app_hash_here",
  "session_path": ""
}
```

Environment variable overrides:

```bash
export TELEGRAM_APP_ID=123456
export TELEGRAM_APP_HASH=your_telegram_app_hash_here
export TELEGRAM_SESSION_PATH=/custom/session/path.json
```

If `session_path` is empty, default is:

```text
~/.modern-telegram-cli/session.json
```

## Build

```bash
go build .
```

## Interactive mode

```bash
./modern-telegram-cli
```

On first run, it prompts for phone number and verification code. Later runs reuse the saved session.

Interactive commands:

- `\me`
- `\contacts`
- `\find <prefix>`
- `\msg <id|@username> <text>`
- `\help`
- `\quit`

## Help and version flags

Root help/version:

```bash
./modern-telegram-cli --help
./modern-telegram-cli -h
./modern-telegram-cli --version
./modern-telegram-cli -v
```

Per-command help:

```bash
./modern-telegram-cli send --help
./modern-telegram-cli get --help
./modern-telegram-cli contacts --help
./modern-telegram-cli me --help
./modern-telegram-cli find --help
```

## One-shot CLI mode

```bash
./modern-telegram-cli <command> [options] [arguments]
```

Commands:

- `send [--json] [--timeout 30s] <user_id|@username> <message>`
- `get [--json] [--timeout 30s] [--limit N] <user_id|@username>`
- `contacts [--json] [--timeout 30s]`
- `me [--json] [--timeout 30s]`
- `find [--json] [--timeout 30s] <prefix>`

Options:

- `--json`
- `--limit N`
- `--timeout duration`

Note: place flags before positional arguments (Go flag parsing behavior).

Examples:

```bash
./modern-telegram-cli send @ken "Hello from script"
./modern-telegram-cli get --limit 20 @ken
./modern-telegram-cli contacts --json
./modern-telegram-cli me
./modern-telegram-cli find ken
```

Note: run interactive mode once to authenticate before using one-shot commands.

## License

MIT
