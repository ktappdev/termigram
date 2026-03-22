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

Your auth session is saved automatically to `~/.termigram/session.json` by default. Chats and message history are fetched from Telegram and are not stored locally by default.

Default interactive mode uses the command/transcript workflow. Open a chat with `\msg`, `\to`, `\chats`, or `\unread`, then type plain text to send into the active chat.

## Everyday usage

### Interactive UI mode

- `./termigram` uses the default command/transcript UI
- `./termigram --ui legacy` uses the same command/transcript UI explicitly

#### Interactive commands

- `\me`
- `\contacts`
- `\find <prefix>`
- `\msg <id|@username> <text>`
- `\to <id|@username>`
- `\chats`
- `\unread`
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

TTY interactive mode defaults to the command/transcript UI.

Interactive chat flow:

- `\msg <id|@user> <text>` sends a message and enters that chat
- `\to <id|@user>` switches the active chat
- `\chats` opens the recent chats picker
- `\unread` opens chats with unread messages
- typing plain text sends to the active chat
- `\close` exits chat mode

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
