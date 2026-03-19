# Branch workflow

This repository uses `dev` as the active integration branch for documentation workflow updates.

## Task links

This workflow update is tied to:

- `termigram-86w`
- `termigram-898`
- `termigram-ubz`
- `termigram-u5y`
- `termigram-9mr`
- `termigram-vzq`

## Phase 1: Start from `dev`

1. Sync your local repository.
2. Check out `dev`.
3. Create a focused branch from `dev`.

```bash
git checkout dev
git pull --rebase origin dev
git checkout -b docs/your-change
```

## Phase 2: Make documentation updates

1. Keep changes scoped to the requested docs.
2. Update existing pages before adding new ones when possible.
3. Add branch-specific guidance when deployment or publishing depends on `dev`.
4. Link related docs so the workflow is easy to follow.

## Phase 3: Review locally

1. Read changed Markdown files for accuracy.
2. Confirm commands reference `dev` where required.
3. Keep files concise and under project size limits.

## Phase 4: Open a pull request

1. Push your branch.
2. Open a pull request targeting `dev`.
3. Reference the linked bd tasks in the pull request description.

## Phase 5: Merge and deploy

1. Merge approved documentation changes into `dev`.
2. Confirm any GitHub Pages workflow references are `dev`-based.
3. Re-check published documentation links after merge.

## Phase 6: Verify `dev` branch state

Use this verification checklist for planner phase 6 style validation:

1. Confirm the merged commit exists on `dev`.
2. Confirm documentation links reference `dev` where branch context matters.
3. Confirm GitHub Pages guidance points contributors to `dev`.
4. Confirm README, contributing docs, and docs pages are consistent.
5. Record the related bd tasks:
   - `termigram-86w`
   - `termigram-898`
   - `termigram-ubz`
   - `termigram-u5y`
   - `termigram-9mr`
   - `termigram-vzq`
