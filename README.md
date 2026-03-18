# termigram

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
~/.termigram/session.json
```

## Build

Use one command:

```bash
make build
```

`make build` runs `./build.sh`, which automatically:
- uses the current git tag/describe value as app version (when tags exist)
- falls back to `dev` when no git tags are available
- injects the version via ldflags

Optional manual override:

```bash
make build-version VERSION=1.2.3
```

## Interactive mode

```bash
./termigram
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

## One-shot CLI mode

```bash
./termigram <command> [options] [arguments]
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
./termigram send @ken "Hello from script"
./termigram get --limit 20 @ken
./termigram contacts --json
./termigram me
./termigram find ken
```

Note: run interactive mode once to authenticate before using one-shot commands.

## License

MIT

## Terminal UI (TUI)

The application includes a full-featured terminal user interface built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Launch the UI by running the application without arguments:

```bash
./termigram
```

### Features Overview

- **Split-pane layout** - Chat list on the left (30%), message view on the right (70%)
- **Real-time updates** - Connection status, typing indicators, and message delivery confirmations
- **Message bubbles** - Styled incoming/outgoing messages with timestamps and read receipts
- **Search** - Filter chats by name or username with `Ctrl+F`
- **Unread indicators** - Badge showing unread message count per chat
- **Draft support** - Automatically saves unsent messages
- **Reply threads** - Visual indication when replying to specific messages
- **Responsive design** - Adapts to terminal size, switches to mobile view on small screens
- **Telegram dark theme** - Color scheme matching Telegram Desktop for familiarity

### Keyboard Shortcuts

#### Navigation

| Shortcut | Action |
|----------|--------|
| `↑` / `↓` | Navigate chats / scroll messages |
| `Enter` | Open selected chat / Send message |
| `Esc` | Go back to chat list / Cancel reply |
| `←` / `→` | Navigate between panels (desktop view) |
| `Home` | Jump to first message |
| `End` | Jump to latest message |

#### Actions

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

#### Message Actions (in message view)

| Shortcut | Action |
|----------|--------|
| `R` | Reply to message |
| `F` | Forward message |
| `D` | Delete message |
| `Ctrl+C` | Copy message text |

### Mouse Usage

The UI supports mouse interaction when running in a terminal with mouse support:

- **Click on chat** - Select and open a chat from the sidebar
- **Scroll wheel** - Scroll through chat list or messages
- **Click on input area** - Focus the message input field
- **Click on buttons** - Interact with modal dialog buttons

Note: Mouse support requires a terminal that supports mouse events (e.g., iTerm2, Kitty, Alacritty, or tmux with mouse enabled).

### Responsive Layout

The UI automatically adapts to terminal size:

#### Desktop View (≥60 columns)
- **Chat list**: 30% width on the left
- **Message view**: 70% width on the right
- **Input area**: Full width below messages
- All panels visible simultaneously

#### Mobile View (<60 columns)
- **Single panel view** - Shows either chat list or messages
- **Navigation**: Use `Enter` to open chat, `Esc` to return to chat list
- **Full width** - Active panel uses entire terminal width
- **Optimized for** - SSH sessions, small terminals, split windows

#### Minimum Requirements
- **Minimum width**: 40 columns (mobile view)
- **Recommended**: 80+ columns for optimal desktop experience
- **Minimum height**: 20 rows

Resize your terminal to switch between layouts dynamically.

### Color Scheme

The UI uses Telegram Desktop's dark theme:

- **Background**: Deep blue-gray (#17212b, #0e1621)
- **Message bubbles**: Blue for incoming, dark for outgoing
- **Text**: White for primary, gray for timestamps/metadata
- **Accents**: Green for online status, blue for links, red for errors
- **Status indicators**: Color-coded connection status (blue=connected, yellow=connecting, red=disconnected)

### Status Indicators

| Indicator | Meaning |
|-----------|---------|
| 🔵 Connected | Successfully connected to Telegram |
| 🟡 Connecting | Establishing connection |
| 🔴 Disconnected | Connection lost or failed |
| ✓ | Message sent (delivered to server) |
| ✓✓ | Message read (green when read by recipient) |
| ● (green) | User is online |
| "typing..." | User is currently typing |
# termigram
