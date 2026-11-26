# PR body template (draft open at step 7, refreshed at step 11)

The PR body `om-auto-create-pr-loop` opens the draft PR with in step 7 (and refreshes in step 11). It **MUST**
include the `Tracking plan:` line so `om-auto-continue-pr-loop` can resume.

PR title convention: conventional-commit prefix scoped to the primary area.
Examples: `feat(ui): add accessible confirmation dialog wrapper`,
`refactor(pricing): extract shared resolver`,
`security(auth): harden session validation`,
`docs(skills): add om-auto-create-pr-loop and om-auto-continue-pr-loop`.

```markdown
Tracking plan: {RUNS_DIR}/{DATE}-{SLUG}/PLAN.md
Tracking run folder: {RUNS_DIR}/{DATE}-{SLUG}/
Status: in-progress

## 🎯 Goal
- {1–2 sentences: the user-visible outcome this PR delivers, and the problem/root cause when this is a fix}

## External References
- {url — what was adopted, what was rejected}  <!-- only if --skill-url was used -->

## What Changed
- {one bullet per changed area — name the files/modules touched and the behavioral change; write for a reviewer who has not opened the diff, not a one-line restatement of the title}

## 🧪 Tests
- {commands run with result counts, e.g. `npm test — 2927 passed / 172 files`; note any skipped or failing gate}
- {the new/updated test cases and what behavior they lock in}

## 💥 Breaking Changes
- {None | describe affected contracts and migration notes}

## 📋 Progress
See the Tasks table in the plan — that is the authoritative Step-status source (`todo` / `done`).

## Handoff & Notifications
- Live handoff: `{RUNS_DIR}/{DATE}-{SLUG}/HANDOFF.md`
- Notifications log: `{RUNS_DIR}/{DATE}-{SLUG}/NOTIFY.md`
```

Flip `Status:` to `complete` on the PR body once every row in the Tasks table has `Status` = `done`.
