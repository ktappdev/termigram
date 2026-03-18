# Contributing to termigram

Thank you for your interest in contributing to termigram.

## Code of Conduct

We want termigram to be a welcoming and inclusive project. Please be respectful and constructive in all interactions.

## Core contributor rules

- Start documentation and workflow changes from `dev`
- Keep changes focused and clearly scoped
- Build and test before submitting changes
- Reference the related bd tasks in PRs and notes

## Task links

Documentation and workflow updates should reference:

- `termigram-86w`
- `termigram-9mr`
- `termigram-vzq`

## Prerequisites

Before you start contributing, make sure you have:

- **Go 1.21+** installed
- **Git** for version control
- Optional Telegram API credentials if you need to override baked-in credentials

## Quick start

```bash
git clone https://github.com/ktappdev/termigram.git
cd termigram
git checkout dev
git checkout -b feature/your-feature-name
```

Build and test:

```bash
make build
go test ./...
```

## Pull requests

Before you submit:

- Build succeeds
- Tests pass
- Manual validation is complete
- The PR targets `dev`
- Related bd tasks are included in the PR description

## More contributor documentation

Detailed workflow guidance moved to:

- [docs/contributing-detail.md](docs/contributing-detail.md)
- [BRANCH_WORKFLOW.md](BRANCH_WORKFLOW.md)
- [CREDENTIALS.md](CREDENTIALS.md)

## Getting help

- Check [README.md](README.md)
- Read existing issues
- File a new issue if needed
