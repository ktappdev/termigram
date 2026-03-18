# Contributing to termigram

Thank you for your interest in contributing to termigram! This guide will help you get started with development.

## Code of Conduct

We want termigram to be a welcoming and inclusive project. Please be respectful and constructive in all interactions. We're working on adding a formal CODE_OF_CONDUCT.md file soon.

## Prerequisites

Before you start contributing, make sure you have:

- **Go 1.21+** installed ([download](https://go.dev/dl/))
- **Git** for version control
- **Telegram API credentials** (optional for most development; see below)

### Getting Telegram API Credentials

termigram supports three credential sources, in this order:

1. `TELEGRAM_APP_ID` / `TELEGRAM_APP_HASH`
2. `config.json` next to the executable
3. baked-in build credentials

That means personal Telegram API credentials are now optional when you're using a build that already includes baked-in credentials.

You'll still want your own credentials if you are:
- testing override behavior
- building your own distribution
- developing without baked-in maintainer credentials

To create your own credentials:
1. Go to [https://my.telegram.org](https://my.telegram.org)
2. Log in with your phone number
3. Click on **API development tools**
4. Create a new application
5. Copy your `app_id` and `app_hash`

## Getting Started

### Clone the repository

```bash
git clone https://github.com/ktappdev/termigram.git
cd termigram
```

### Set up your configuration

For most contributors, no credential setup is required if your build includes baked-in maintainer credentials.

If you need to override credentials, choose either option below.

Copy the example config file and add your credentials:

```bash
cp config.json.example config.json
```

Or use environment variables:

```bash
export TELEGRAM_APP_ID=your_app_id
export TELEGRAM_APP_HASH=your_app_hash
```

Environment variables override `config.json`, and `config.json` overrides baked-in build credentials.

### Build the project

**Using Make (recommended):**

```bash
make build
```

This runs `./build.sh` which automatically sets the version from git tags.

**Specifying a version:**

```bash
make build-version VERSION=1.2.3
```

**Building with baked-in credentials (maintainers):**

```bash
make build TELEGRAM_APP_ID_BAKED=123456 TELEGRAM_APP_HASH_BAKED=your_app_hash
```

See [CREDENTIALS.md](CREDENTIALS.md) for secure handling guidance.

**Manual build:**

```bash
go build -o termigram .
```

### Run the application

```bash
./termigram
```

On first run, you'll need to authenticate:
1. Enter your phone number
2. Enter the verification code sent by Telegram
3. Start using the CLI

The session is saved to `~/.termigram/session.json` by default.

## Project Structure

```
termigram/
├── main.go              # Entry point, CLI argument parsing
├── client.go            # Core Telegram client (gotd/td wrapper)
├── backend.go           # Backend logic for CLI commands
├── auth.go              # Authentication flow handling
├── messaging.go         # Message sending/receiving logic
├── commands.go          # CLI command implementations
├── config.go            # Configuration loading
├── helpers.go           # Utility functions
├── ui/                  # Bubble Tea TUI implementation
│   ├── app.go           # Main UI model and initialization
│   ├── chatlist.go      # Chat list component
│   ├── messages.go      # Message view component
│   ├── input.go         # Message input component
│   ├── styles.go        # Styling and colors
│   ├── modals.go        # Modal dialogs
│   └── ...
└── config.json.example  # Example configuration
```

### Key Files to Know

| File | Purpose |
|------|---------|
| `main.go` | Application entry point, CLI mode routing |
| `client.go` | Telegram client initialization, update handling |
| `backend.go` | Core backend for both UI and CLI modes |
| `ui/app.go` | Bubble Tea application model |
| `ui/backend.go` | UI backend adapter |
| `build.sh` | Build script with version injection |

## Development Workflow

### Branch naming

Use descriptive branch names:

```
feature/add-message-search
fix/typing-indicator-bug
docs/update-readme
refactor/simplify-auth-flow
```

### Making changes

1. **Create a branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** with clear, focused commits

3. **Test your changes**:
   - Build: `make build`
   - Run interactively: `./termigram`
   - Test CLI commands: `./termigram send --json @test "Hello"`

4. **Commit your changes**:
   ```bash
   git commit -m "Add feature description"
   ```

### Code Style

termigram follows standard Go conventions with these project-specific patterns:

**Type naming:**
- Use PascalCase for exported types and functions
- Use camelCase for private types and functions

**Boolean functions:**
- Prefix with `is` or `has`: `isCLIMode()`, `hasHelpFlag()`
- Example: `func isAuthorized() bool { ... }`

**Context passing:**
- Always pass `context.Context` as the first parameter
- Use `context.Background()` for top-level contexts
- Use `context.WithTimeout()` for operations that should timeout

**Error handling:**
- Return errors, don't panic
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Log errors appropriately for user visibility

**Example:**

```go
func (c *Client) SendMessage(ctx context.Context, target string, text string) error {
    if !c.isAuthorized() {
        return fmt.Errorf("not authenticated")
    }
    
    user, ok := c.usersByName[target]
    if !ok {
        return fmt.Errorf("user not found: %s", target)
    }
    
    _, err := c.sender.To(user).Text(ctx, text)
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }
    
    return nil
}
```

**Imports:**
- Standard library imports first
- External packages second (grouped by domain)
- Local imports last

**UI components:**
- Keep components focused and composable
- Use the `Styles` struct for consistent theming
- Follow Bubble Tea's update/view/init pattern

## Testing

### Manual testing

Since termigram interacts with the Telegram API, most testing is manual:

1. **Test authentication flow** - Run `./termigram` and verify login works
2. **Test CLI commands** - Verify each command works as documented:
   ```bash
   ./termigram send @test "Hello"
   ./termigram get --limit 5 @test
   ./termigram contacts --json
   ./termigram me --json
   ```
3. **Test TUI** - Navigate the UI, send messages, verify responsiveness

### Automated testing

Add unit tests for pure functions and helpers:

```bash
go test ./...
go test -v ./...  # Verbose output
go test -race ./...  # Race detection
```

Run tests before submitting PRs.

## Submitting Changes

### Before you submit

- [ ] Your changes build without errors (`make build`)
- [ ] Tests pass (`go test ./...`)
- [ ] You've tested the changes manually
- [ ] Code follows the style guidelines
- [ ] Commits are squashed/focused (one logical change per commit)

### Creating a Pull Request

1. **Push your branch:**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Open a PR** on GitHub with:
   - **Title**: Clear, concise description of the change
   - **Description**: What you changed and why
   - **Testing**: How you tested the changes
   - **Screenshots**: For UI changes (if applicable)

3. **PR Template** (coming soon): Fill out the template when available

### What makes a good PR?

**Good:**
- "Fix nil pointer panic when chat list is empty"
- "Add message search functionality with Ctrl+F"
- "Improve error messages for authentication failures"

**Needs improvement:**
- "Fix stuff"
- "Various improvements"
- "Bug fixes"

### Code review

Expect feedback on your PR. Common review points:
- Error handling completeness
- Edge cases
- Code clarity
- Adherence to Go idioms

Respond to feedback and push updates. The maintainers will merge once approved.

## Issue Reporting

### Before filing an issue

- [ ] Check existing issues (open and closed)
- [ ] Verify the issue persists on the latest version
- [ ] Gather relevant information (error messages, steps to reproduce)

### Filing a bug report

Include as much detail as possible:

```markdown
**Describe the bug**
Clear description of what's happening vs what you expected.

**To Reproduce**
Steps to reproduce:
1. Run `./termigram`
2. Execute command: `...`
3. See error: `...`

**Expected behavior**
What should happen instead.

**Environment:**
- OS: macOS 14.0 / Ubuntu 22.04 / Windows 11
- Go version: 1.21.0
- termigram version: 1.2.0 (run `./termigram --version`)

**Error output**
Paste any error messages or stack traces.

**Screenshots**
If applicable, add screenshots to help explain.
```

### Filing a feature request

```markdown
**Is your feature request related to a problem?**
"I'm always frustrated when..."

**Describe the solution you'd like**
Clear description of what you want to happen.

**Describe alternatives you've considered**
Other approaches you thought about.

**Additional context**
Any other details, mockups, or examples.
```

## Where to Find Things

### Core logic
- **Authentication**: `auth.go`, `client.go` (auth flow)
- **CLI commands**: `commands.go`, `backend.go`
- **Messaging**: `messaging.go`, `client.go`
- **Configuration**: `config.go`, `config.json.example`

### UI (Bubble Tea)
- **Main app**: `ui/app.go`
- **Chat list**: `ui/chatlist.go`
- **Messages**: `ui/messages.go`
- **Input**: `ui/input.go`
- **Styling**: `ui/styles.go`, `ui/colors.go`
- **Modals**: `ui/modals.go`

### Build & tooling
- **Build script**: `build.sh`
- **Makefile**: `Makefile`
- **Module definition**: `go.mod`

## Getting Help

- **Check the README.md** for usage documentation
- **Read existing issues** for similar questions
- **File a new issue** with your question if nothing helps

## Current Issues

Check the [Issues page](https://github.com/ktappdev/termigram/issues) for open issues. Issues are labeled by type and priority:

- **Priority 0**: Critical (security, broken builds)
- **Priority 1**: High (major features, important bugs)
- **Priority 2**: Medium (default)
- **Priority 3**: Low (polish, optimization)
- **Priority 4**: Backlog

Pick an issue that interests you and leave a comment to let others know you're working on it.

## Thank You!

Contributions make open source projects thrive. Whether you're fixing a typo, reporting a bug, or adding a major feature—your help is appreciated!

If you're unsure about anything, don't hesitate to file an issue or ask questions. The maintainers are happy to help.
