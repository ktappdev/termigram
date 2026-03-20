# Quickstart on `dev`

Use this quickstart when validating the current documentation and workflow on the `dev` branch.

## Task links

This page tracks the `dev` branch documentation work for:

- `termigram-86w`
- `termigram-898`
- `termigram-ubz`
- `termigram-u5y`
- `termigram-9mr`
- `termigram-vzq`

## 1. Clone and switch to `dev`

```bash
git clone https://github.com/ktappdev/termigram.git
cd termigram
git checkout dev
```

## 2. Build the CLI

```bash
make build
```

## 3. Run once to authenticate

```bash
./termigram
```

On first run:

1. Enter your phone number
2. Enter the Telegram verification code
3. Reuse the saved auth session on later runs

By default, that session is stored in `~/.termigram/session.json`. Chats and message history are fetched from Telegram and are not stored locally by default.

## 4. Validate key CLI commands

```bash
./termigram --help
./termigram me --json
./termigram contacts --json
```

## 5. Verify GitHub Pages docs changes on `dev`

Before merging documentation updates:

1. Confirm your branch was created from `dev`
2. Confirm your pull request targets `dev`
3. Confirm GitHub Pages workflow references and links point to `dev`
4. Confirm merged docs are visible after deployment

## Related docs

- [Docs index](./index.md)
- [Install guide](./install.md)
- [FAQ and troubleshooting](./faq.md)
- [README](../README.md)
- [Contributing](../CONTRIBUTING.md)
- [Branch workflow](../BRANCH_WORKFLOW.md)
