# Contributing workflow details

This guide contains the detailed contributor workflow that was split from `CONTRIBUTING.md` to keep the main file under the project limit.

## Task links

This documentation split is tied to:

- `termigram-86w`
- `termigram-9mr`
- `termigram-vzq`

## Credentials and configuration

termigram supports three credential sources, in this order:

1. `TELEGRAM_APP_ID` / `TELEGRAM_APP_HASH`
2. `config.json` next to the executable
3. baked-in build credentials

To create your own Telegram credentials:

1. Go to [https://my.telegram.org](https://my.telegram.org)
2. Log in with your phone number
3. Click **API development tools**
4. Create a new application
5. Copy your `app_id` and `app_hash`

## Build options

Recommended:

```bash
make build
```

Manual build:

```bash
go build -o termigram .
```

Maintainer build with baked-in credentials:

```bash
make build TELEGRAM_APP_ID_BAKED=123456 TELEGRAM_APP_HASH_BAKED=your_app_hash
```

## Development workflow

### Branch naming

Use descriptive names such as:

- `feature/add-message-search`
- `fix/typing-indicator-bug`
- `docs/update-readme`
- `refactor/simplify-auth-flow`

### Making changes

1. Start from `dev`:
   ```bash
   git checkout dev
   git checkout -b feature/your-feature-name
   ```
2. Make focused commits
3. Test your changes
4. Commit with a clear message

## Code style

- Use standard Go naming conventions
- Pass `context.Context` first
- Return errors instead of panicking
- Wrap errors with context
- Keep UI components focused and composable

## Testing

### Manual testing

1. Test authentication flow with `./termigram`
2. Test CLI commands such as:
   ```bash
   ./termigram send @test "Hello"
   ./termigram get --limit 5 @test
   ./termigram contacts --json
   ./termigram me --json
   ```
3. Test TUI navigation and responsiveness

### Automated testing

```bash
go test ./...
go test -v ./...
go test -race ./...
```

## Pull request quality

A good PR has:

- a clear title
- a concise description
- testing notes
- screenshots for UI changes when needed
- linked bd task references

## Issue reporting

When filing bugs, include:

- what happened
- how to reproduce it
- expected behavior
- OS and Go version
- `./termigram --version` output
- error output or screenshots when useful

## Project structure

Key files to know:

- `main.go`
- `client.go`
- `backend.go`
- `auth.go`
- `messaging.go`
- `commands.go`
- `config.go`
- `helpers.go`
- `interactive_mode.go`
- `legacy_chat_view.go`
- `build.sh`

## Current issues

Documentation changes targeting `dev` should reference:

- `termigram-86w`
- `termigram-9mr`
- `termigram-vzq`

See the GitHub issues page for open work and priorities.
