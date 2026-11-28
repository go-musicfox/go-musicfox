---
description: git commit and push
subtask: true
---

commit and push

Follow the Git rules in AGENTS.md.

- Subject: `<type>(<scope>): <description>` — Conventional Commits, English only, imperative mood, lowercase description, no trailing period, no issue number, ≤50 characters.
- Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `revert`.
- Scope: the module you changed, e.g. `ui`, `player`, `lyric`, `playlist`, `track`, `login`, `keybindings`, `storage`, `configs`, `deps`, `build`, `docs`.
- Body: explain WHY from an end user perspective instead of WHAT was done; wrap at 72 characters.
- No generic messages like "improved agent experience"; be specific about the user-facing change.
- Breaking changes: `<type>(<scope>)!:` plus a `BREAKING CHANGE:` footer describing the impact.
- Chinese is not allowed in the subject, body, or footers.

if there are conflicts DO NOT FIX THEM. notify me and I will fix them

## GIT DIFF

!`git diff`

## GIT DIFF --cached

!`git diff --cached`

## GIT STATUS --short

!`git status --short`
