---
title: termigram GitHub Pages on dev
description: Install and deploy the termigram GitHub Pages documentation from the dev branch, with contributor workflow links and task tracking.
keywords:
  - termigram
  - Telegram CLI
  - Go CLI
  - MTProto
  - GitHub Pages
  - dev branch
---

# termigram docs on `dev`

This GitHub Pages content is maintained from the `dev` branch.

## Branch and task tracking

Use the `dev` branch for documentation work and deployment verification tied to:

- `termigram-86w`
- `termigram-898`
- `termigram-ubz`
- `termigram-u5y`
- `termigram-9mr`
- `termigram-vzq`

## Install from `dev`

To review documentation changes locally from the active branch:

```bash
git clone https://github.com/ktappdev/termigram.git
cd termigram
git checkout dev
```

Build the CLI from `dev` if you also want to validate command examples:

```bash
make build
./termigram --help
```

## Deploy GitHub Pages content from `dev`

GitHub Pages documentation updates should be prepared, reviewed, and verified from `dev`.

Recommended workflow:

1. Update docs on a topic branch created from `dev`
2. Open a pull request targeting `dev`
3. Confirm the GitHub Pages workflow and links reference `dev`
4. Merge to `dev`
5. Verify the published site reflects the merged docs change

## What this site covers

termigram is a Telegram CLI written in Go. These docs focus on:

- installing and building the CLI
- running Telegram authentication flows
- scripting with one-shot commands
- contributing documentation and workflow updates

## Related pages

- [Quickstart](./quickstart.md)
- [Install guide](./install.md)
- [FAQ and troubleshooting](./faq.md)
- [Contributor workflow details](./contributing-detail.md)
- [Repository README](../README.md)
- [Contributing guide](../CONTRIBUTING.md)
- [Branch workflow](../BRANCH_WORKFLOW.md)
