# Installation and build guide

This install guide collects the detailed setup steps that were split out of `README.md` to keep the main README under the project file-size limit.

## Task links

This documentation split is tied to:

- `termigram-86w`
- `termigram-9mr`
- `termigram-vzq`

## Installation options

### Option A: Install with Go

```bash
go install github.com/ktappdev/termigram@latest
```

### Option B: Build from a local clone

```bash
make build
```

### Run termigram

```bash
./termigram
```

On first run:

1. Enter your phone number
2. Enter the verification code sent by Telegram
3. Start chatting

## JSON command examples

### Send a message

```bash
./termigram send --json @ken "Hello from automation"
```

### Fetch recent messages

```bash
./termigram get --json --limit 5 @ken
```

### List contacts

```bash
./termigram contacts --json
```

### Show current account

```bash
./termigram me --json
```

### Find cached usernames

```bash
./termigram find --json ke
```

## Build from source

Use one command:

```bash
make build
```

`make build` runs `./build.sh`, which automatically:

- uses `git describe --tags --dirty` as version when tags exist
- falls back to `dev` when no tags are available
- injects version via ldflags

Optional manual override:

```bash
make build-version VERSION=1.2.3
```

## Advanced configuration

Credential lookup order:

1. `TELEGRAM_APP_ID` / `TELEGRAM_APP_HASH`
2. `config.json` next to the executable
3. baked-in build credentials

### Session storage

By default, sessions are stored at `~/.termigram/session.json`.

Override with:

```bash
export TELEGRAM_SESSION_PATH=/custom/path/session.json
```

### Override Telegram API credentials

```bash
export TELEGRAM_APP_ID=your_app_id
export TELEGRAM_APP_HASH=your_app_hash
```

Or create `config.json` next to the executable by copying `config.json.example` and filling in your own values.

Need your own Telegram app credentials? Create them at [https://my.telegram.org](https://my.telegram.org).

### Environment variables

| Variable | Description |
|----------|-------------|
| `TELEGRAM_APP_ID` | Override Telegram app id |
| `TELEGRAM_APP_HASH` | Override Telegram app hash |
| `TELEGRAM_SESSION_PATH` | Custom session file location |
