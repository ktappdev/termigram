# FAQ and troubleshooting

This troubleshooting guide was split from `README.md` to keep the main file concise.

## Task links

This documentation split is tied to:

- `termigram-86w`
- `termigram-9mr`
- `termigram-vzq`

## Troubleshooting

### One-shot commands do not work yet

Run interactive mode once first:

```bash
./termigram
```

That creates or reuses the local auth session file before scripted commands run. Chats and message history are still fetched from Telegram when needed.

### Build output shows `dev` as the version

That is expected when no tags are available. `make build` falls back to `dev`.

### I need to override Telegram credentials

Use environment variables:

```bash
export TELEGRAM_APP_ID=your_app_id
export TELEGRAM_APP_HASH=your_app_hash
```

You can also place a `config.json` next to the executable.

### Where can I find contributor workflow docs?

See:

- [`CONTRIBUTING.md`](../CONTRIBUTING.md)
- [`docs/contributing-detail.md`](./contributing-detail.md)
- [`BRANCH_WORKFLOW.md`](../BRANCH_WORKFLOW.md)

### Where can I find GitHub Pages and `dev` branch docs guidance?

See:

- [`docs/index.md`](./index.md)
- [`docs/quickstart.md`](./quickstart.md)
- [`README.md`](../README.md)
