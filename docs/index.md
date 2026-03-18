---
title: termigram Documentation
---

# termigram Documentation

termigram is a lightweight Telegram command-line client written in Go using MTProto via `gotd/td`.

This docs set is organized for GitHub Pages and quick onboarding.

## Start here

- [Quickstart](./quickstart.md) — install, authenticate, and send your first message
- [Features](./features.md) — interactive CLI, automation support, and terminal UI capabilities
- [CLI reference](./cli-reference.md) — one-shot commands, flags, and examples
- [TUI guide](./tui-guide.md) — keyboard shortcuts, layout, and navigation tips

## What termigram supports

- Interactive login and chat usage
- One-shot CLI commands for scripts and AI agents
- Saved local sessions for repeat use
- Contact lookup and username prefix search
- Terminal UI patterns for chat selection and message browsing

## Recommended path

If you are new to the project:

1. Read [Quickstart](./quickstart.md)
2. Review [CLI reference](./cli-reference.md)
3. Use [TUI guide](./tui-guide.md) for interactive navigation help

## Build from source

```bash
make build
```

The built binary can then be run with:

```bash
./termigram
```

## Configuration summary

termigram supports Telegram credentials from:

1. `TELEGRAM_APP_ID` and `TELEGRAM_APP_HASH`
2. `config.json` next to the executable
3. baked-in build credentials when provided by a release

Session storage defaults to:

```text
~/.termigram/session.json
```

## License

MIT
