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

Default interactive mode keeps the familiar command/transcript workflow. Open a chat with `\msg` or `\to`, then type plain text to send into the active chat.

If you want the optional Bubble Tea split-pane UI instead, run:

```bash
./termigram --ui tui
```

## Everyday usage

### Interactive UI modes

- `./termigram` uses the legacy command/transcript UI by default
- `./termigram --ui legacy` forces the legacy command/transcript UI
- `./termigram --ui tui` opens the optional Bubble Tea split-pane UI

#### Legacy interactive commands

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

TTY interactive mode defaults to the legacy command/transcript UI, but it now owns prompt input so sent messages do not echo twice and redraws the active chat transcript on resize for more stable bubbles.

Legacy chat flow:

- `\msg <id|@user> <text>` starts a chat and sends immediately
- `\to <id|@user>` switches the active chat
- typing plain text sends to the active chat
- `\unread` jumps to a chat with unread messages
- `\close` exits chat mode

Optional Bubble Tea UI:

- Run `./termigram --ui tui`
- See `/Users/kentaylor/developer/telegram-cli/termigram/docs/tui-guide.md`
- Resize implementation notes remain in `/Users/kentaylor/developer/telegram-cli/termigram/docs/bubbletea-resize-research.md`

## Help and version

```bash
./termigram --help
./termigram --version
./termigram --ui tui
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
