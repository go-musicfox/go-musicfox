# PR body template (draft open at step 6, refreshed at step 10)

The PR body `om-auto-create-pr` opens the draft PR with in step 6 (and refreshes in step 10) — the inline-fallback
copy of the unified template owned by `om-open-pr` (its
`references/pr-body-template.md`); the two must stay in sync (see the cross-skill
contract in this repo's AGENTS.md). It **MUST** include the `Tracking plan:` line
so `om-auto-continue-pr` can resume.

PR title convention: conventional-commit prefix scoped to the primary area.
Examples: `feat(ui): add accessible confirmation dialog wrapper`,
`refactor(pricing): extract shared resolver`,
`security(auth): harden session validation`,
`docs(skills): add om-auto-create-pr and om-auto-continue-pr`.

```markdown
Closes #{issueId}                <!-- only in issue-driven runs; `Refs #{issueId}`
                                      for a spec-only design PR -->
Tracking plan: {RUNS_DIR}/{DATE}-{SLUG}.md
Source doc: {SPECS_DIR}/{spec}.md     <!-- only when a spec/design doc drives the run -->
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
See the Progress section in the tracking plan.
```

Flip `Status:` to `complete` on the PR body once all Progress steps are checked.
