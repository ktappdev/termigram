# termigram

A lightweight Telegram command-line client written in Go using MTProto (`gotd/td`).

## Get started in 2 minutes

This is the fastest path for most users: install → run → authenticate → chat.

### Install and first run

See the dedicated install guide for full setup steps, build options, and configuration:

- [docs/install.md](docs/install.md)

Quick path:

```bash
go install github.com/ktappdev/termigram@latest
./termigram
```

On first run:
1. Enter your phone number
2. Enter the verification code sent by Telegram
3. Start chatting

Your session is saved automatically to `~/.termigram/session.json` by default.

Send your first message:

```text
\msg @username Hello!
```

Or open the interactive chat picker:

```text
\chats
```

## Everyday usage

### Interactive commands

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

### One-shot CLI mode

Use termigram non-interactively for scripts, cron jobs, and AI agents.

```bash
./termigram <command> [--json] [--timeout 30s] [command flags] [arguments]
```

Before one-shot commands work, run interactive once and complete phone login:

```bash
./termigram
```

#### Command reference

- `send <user_id|@username> <message>`
- `get [--limit N] <user_id|@username>`
- `contacts`
- `me`
- `find <prefix>`

#### Flags and options

- `--json`: machine-readable output envelope (`success`, `data`, `error`)
- `--timeout 30s`: command timeout
- `--limit N`: only for `get` (default `10`)

#### Common automation patterns

```bash
./termigram send --json @oncall "Job failed: nightly-import"
./termigram get --json --limit 20 @ken | jq '.data.messages[] | {id, from_name, message}'
./termigram me --json | jq '.data.id'
```

#### Tips for AI agents

- Prefer `--json` for deterministic parsing.
- Treat non-zero exit codes as failures.
- Use explicit `--timeout` in automation.
- Run one auth bootstrap step before automated one-shot commands.
- Place flags before positional args.

## Features overview

### Core CLI features

- Interactive mode for login and chat commands
- One-shot CLI mode for scripting
- Reuses a saved user session
- Commands: `send`, `get`, `contacts`, `me`, `find`

### Terminal UI notes

The repo also contains a Bubble Tea-based UI implementation and related docs/components.

#### Features

- Split-pane layout
- Real-time updates
- Message bubbles
- Search
- Unread indicators
- Draft support
- Reply threads
- Responsive design
- Telegram dark theme

#### Keyboard shortcuts

| Shortcut | Action |
|----------|--------|
| `↑` / `↓` | Navigate chats / scroll messages |
| `Enter` | Open selected chat / Send message |
| `Esc` | Go back to chat list / Cancel reply |
| `Ctrl+C` | Quit application |
| `Ctrl+N` | Start new conversation |
| `Ctrl+F` | Focus search bar |
| `Ctrl+Enter` | New line in message |
| `/` | Focus message input |
| `?` | Show help |

#### Responsive layout

- Desktop view: split chat list and message view
- Mobile view: single active panel
- Recommended width: 80+ columns
- Minimum height: 20 rows

## Help and version

```bash
./termigram --help
./termigram --version
./termigram send --help
./termigram get --help
./termigram contacts --help
./termigram me --help
./termigram find --help
```

## Documentation guides

For content moved out of the main README, see:

- [docs/install.md](docs/install.md)
- [docs/faq.md](docs/faq.md)
- [docs/index.md](docs/index.md)
- [docs/quickstart.md](docs/quickstart.md)
- [BRANCH_WORKFLOW.md](BRANCH_WORKFLOW.md)
- [CONTRIBUTING.md](CONTRIBUTING.md)

## GitHub Pages docs workflow

GitHub Pages documentation updates should explicitly target the `dev` branch.

Task links for this docs workflow update:

- `termigram-86w`
- `termigram-9mr`
- `termigram-vzq`

Use this workflow when updating published docs:

1. Create your docs branch from `dev`
2. Open your pull request against `dev`
3. Verify documentation workflow references and deployment guidance point to `dev`
4. Merge into `dev`
5. Confirm the GitHub Pages site reflects the merged change

## License

MIT
