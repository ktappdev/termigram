# Maintainer credentials guide

termigram supports Telegram API credentials from three sources, in this order:

1. `TELEGRAM_APP_ID` / `TELEGRAM_APP_HASH`
2. `config.json` next to the executable
3. baked-in build credentials

This allows release builds to work out of the box while still letting developers override credentials locally.

## Build-time injection

Inject baked-in credentials with ldflags:

```bash
go build -ldflags "-X main.appVersion=v1.2.3 -X main.telegramAppIDBaked=123456 -X main.telegramAppHashBaked=your_app_hash" -o termigram .
```

Or use the project build helpers:

```bash
make build TELEGRAM_APP_ID_BAKED=123456 TELEGRAM_APP_HASH_BAKED=your_app_hash
make build-version VERSION=1.2.3 TELEGRAM_APP_ID_BAKED=123456 TELEGRAM_APP_HASH_BAKED=your_app_hash
```

`build.sh` only injects baked credentials when those environment variables are set.

## Secure handling

- Do not commit real credentials to source control.
- Prefer passing credentials through CI/CD secret storage or a local shell environment.
- Avoid storing maintainers' shared credentials in `config.json`.
- Use personal credentials locally when testing override behavior.

Example with environment variables in CI:

```bash
export TELEGRAM_APP_ID_BAKED="${TELEGRAM_APP_ID_BAKED}"
export TELEGRAM_APP_HASH_BAKED="${TELEGRAM_APP_HASH_BAKED}"
make build-version VERSION="${GIT_TAG}"
```

## Expected runtime behavior

At startup, termigram resolves credentials like this:

- If `TELEGRAM_APP_ID` / `TELEGRAM_APP_HASH` are set, they win.
- Otherwise, if `config.json` provides both values, those are used.
- Otherwise, baked-in credentials are used if the binary was built with them.
- If none of the three sources provide a complete pair, termigram shows a helpful setup error.

## Backward compatibility

Existing `config.json` files continue to work unchanged. A user-provided `config.json` still overrides baked-in credentials.
