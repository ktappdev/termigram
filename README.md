# termigram

A lightweight Telegram command-line client written in Go using MTProto (`gotd/td`).

## Get started in 2 minutes

This is the fastest path for most users: install → authenticate → chat.

### 1) Install

If you already have a binary, skip to step 2.

Option A (Go installed):

```bash
go install github.com/ktappdev/termigram@latest
```

Option B (from a local clone):

```bash
make build
```

### 2) Create Telegram API credentials (one-time)

1. Go to [my.telegram.org](https://my.telegram.org)
2. Sign in with your phone number
3. Open **API development tools**
4. Create an app and copy:
   - `app_id`
   - `app_hash`

### 3) Add config

Create `config.json` next to the `termigram` executable:

```json
{
  "telegram_app_id": 123456,
  "telegram_app_hash": "your_telegram_app_hash_here",
  "session_path": ""
}
```

If `session_path` is empty, default is:

```text
~/.termigram/session.json
```

(You can also use env vars: `TELEGRAM_APP_ID`, `TELEGRAM_APP_HASH`, `TELEGRAM_SESSION_PATH`.)

### 4) Authenticate and start chatting

```bash
./termigram
```

On first run, enter your phone number and the Telegram code. After that, your session is reused.

Send your first message:

```text
\msg @username Hello!
```

Or pick a chat and then send plain text:

```text
\chats
```

Select a chat with arrows + Enter, then just type messages directly.

---

## Everyday usage

### Interactive commands

- `\me`
- `\contacts`
- `\find <prefix>`
- `\msg <id|@username> <text>`
- `\to <id|@username>`
- `\chats` (interactive picker with filter + arrows + Enter)
- `\here`
- `\close`
- `\help`
- `\quit`

### One-shot CLI mode (scripting + AI agents)

Use termigram non-interactively for scripts, cron jobs, and AI agents.

```bash
./termigram <command> [--json] [--timeout 30s] [command flags] [arguments]
```

#### Authentication prerequisite (important)

Before one-shot commands work, run interactive once and complete phone login:

```bash
./termigram
```

That creates/reuses a local session (default: `~/.termigram/session.json`).

#### Command reference

- `send <user_id|@username> <message>`
- `get [--limit N] <user_id|@username>`
- `contacts`
- `me`
- `find <prefix>`

#### Flags and options

- `--json`: machine-readable output envelope (`success`, `data`, `error`)
- `--timeout 30s`: command timeout (Go duration syntax, e.g. `10s`, `1m`)
- `--limit N`: only for `get` (default `10`)

---

#### `send` (send a message)

```bash
./termigram send --json @ken "Hello from automation"
```

Example JSON output:

```json
{
  "success": true,
  "data": {
    "target": "@ken",
    "message": "Hello from automation",
    "timestamp": 1719012345,
    "sent_to": "ken",
    "user_id": 123456789
  }
}
```

#### `get` (fetch recent messages)

```bash
./termigram get --json --limit 5 @ken
```

Example JSON output:

```json
{
  "success": true,
  "data": {
    "target": "@ken",
    "count": 2,
    "user": "ken",
    "user_id": 123456789,
    "messages": [
      {
        "id": 101,
        "from_id": 123456789,
        "from_name": "Ken",
        "message": "hey",
        "date": 1719012301
      },
      {
        "id": 102,
        "from_id": 555000111,
        "from_name": "You",
        "message": "hello",
        "date": 1719012345
      }
    ]
  }
}
```

#### `contacts` (list contacts)

```bash
./termigram contacts --json
```

Example JSON output:

```json
{
  "success": true,
  "data": {
    "count": 2,
    "contacts": [
      {
        "user_id": 123456789,
        "first_name": "Ken",
        "last_name": "Taylor",
        "username": "ken"
      },
      {
        "user_id": 987654321,
        "first_name": "Alex",
        "last_name": "Doe"
      }
    ]
  }
}
```

#### `me` (current account)

```bash
./termigram me --json
```

Example JSON output:

```json
{
  "success": true,
  "data": {
    "id": 555000111,
    "first_name": "Your",
    "last_name": "Name",
    "username": "yourname",
    "phone": "+1555000111"
  }
}
```

#### `find` (username prefix lookup from cache)

```bash
./termigram find --json ke
```

Example JSON output:

```json
{
  "success": true,
  "data": {
    "prefix": "ke",
    "count": 3,
    "matches": [
      "@ken",
      "@kendra",
      "@kevin"
    ]
  }
}
```

---

#### Common automation patterns

- Send alert from script:

```bash
./termigram send --json @oncall "Job failed: nightly-import"
```

- Poll recent messages and parse with `jq`:

```bash
./termigram get --json --limit 20 @ken | jq '.data.messages[] | {id, from_name, message}'
```

- Resolve self identity for diagnostics:

```bash
./termigram me --json | jq '.data.id'
```

#### Tips for AI agents

- Prefer `--json` for deterministic parsing.
- Treat non-zero exit codes as failures; parse stderr for details.
- Use explicit `--timeout` in automation to avoid hanging tasks.
- Run one auth bootstrap step (`./termigram`) in environment setup before automated one-shot commands.
- Place flags before positional args (Go flag parsing behavior).

---

## Features overview

### Core CLI features

- Interactive mode for login and chat commands
- One-shot CLI mode for scripting
- Reuses a saved user session
- Commands: `send`, `get`, `contacts`, `me`, `find`

### Terminal UI notes

The repo also contains a Bubble Tea-based UI implementation and related docs/components.

#### Features Overview

- **Split-pane layout** - Chat list on the left (30%), message view on the right (70%)
- **Real-time updates** - Connection status, typing indicators, and message delivery confirmations
- **Message bubbles** - Styled incoming/outgoing messages with timestamps and read receipts
- **Search** - Filter chats by name or username with `Ctrl+F`
- **Unread indicators** - Badge showing unread message count per chat
- **Draft support** - Automatically saves unsent messages
- **Reply threads** - Visual indication when replying to specific messages
- **Responsive design** - Adapts to terminal size, switches to mobile view on small screens
- **Telegram dark theme** - Color scheme matching Telegram Desktop for familiarity

#### Keyboard Shortcuts

##### Navigation

| Shortcut | Action |
|----------|--------|
| `↑` / `↓` | Navigate chats / scroll messages |
| `Enter` | Open selected chat / Send message |
| `Esc` | Go back to chat list / Cancel reply |
| `←` / `→` | Navigate between panels (desktop view) |
| `Home` | Jump to first message |
| `End` | Jump to latest message |

##### Actions

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Quit application (from chat list) |
| `Ctrl+N` | Start new conversation |
| `Ctrl+F` | Focus search bar |
| `Ctrl+Enter` | New line in message |
| `Ctrl+\` | Toggle sidebar collapse |
| `/` | Focus message input |
| `?` | Show help |
| `Tab` | Open attachment menu |

##### Message Actions (in message view)

| Shortcut | Action |
|----------|--------|
| `R` | Reply to message |
| `F` | Forward message |
| `D` | Delete message |
| `Ctrl+C` | Copy message text |

#### Mouse Usage

- **Click on chat** - Select and open a chat from the sidebar
- **Scroll wheel** - Scroll through chat list or messages
- **Click on input area** - Focus the message input field
- **Click on buttons** - Interact with modal dialog buttons

#### Responsive Layout

##### Desktop View (≥60 columns)
- **Chat list**: 30% width on the left
- **Message view**: 70% width on the right
- **Input area**: Full width below messages
- All panels visible simultaneously

##### Mobile View (<60 columns)
- **Single panel view** - Shows either chat list or messages
- **Navigation**: Use `Enter` to open chat, `Esc` to return to chat list
- **Full width** - Active panel uses entire terminal width
- **Optimized for** - SSH sessions, small terminals, split windows

##### Minimum Requirements
- **Minimum width**: 40 columns (mobile view)
- **Recommended**: 80+ columns for optimal desktop experience
- **Minimum height**: 20 rows

#### Color Scheme

- **Background**: Deep blue-gray (#17212b, #0e1621)
- **Message bubbles**: Blue for incoming, dark for outgoing
- **Text**: White for primary, gray for timestamps/metadata
- **Accents**: Green for online status, blue for links, red for errors
- **Status indicators**: Color-coded connection status (blue=connected, yellow=connecting, red=disconnected)

#### Status Indicators

| Indicator | Meaning |
|-----------|---------|
| 🔵 Connected | Successfully connected to Telegram |
| 🟡 Connecting | Establishing connection |
| 🔴 Disconnected | Connection lost or failed |
| ✓ | Message sent (delivered to server) |
| ✓✓ | Message read (green when read by recipient) |
| ● (green) | User is online |
| "typing..." | User is currently typing |

---

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
./termigram get --help
./termigram contacts --help
./termigram me --help
./termigram find --help
```

---

## Build (from source)

This section is intentionally last for users who want to build locally.

Use one command:

```bash
make build
```

`make build` runs `./build.sh`, which automatically:
- uses `git describe --tags --dirty` as version (when tags exist)
- falls back to `dev` when no tags are available
- injects version via ldflags

Optional manual override:

```bash
make build-version VERSION=1.2.3
```

## License

MIT
