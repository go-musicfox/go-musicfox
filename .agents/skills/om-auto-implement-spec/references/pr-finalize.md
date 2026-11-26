# PR finalize — open or reuse, labels, summary comment, markers

The single procedure for the "commit → push → open (or reuse) the PR → normalize labels → summary comment → chaining reference lines" mechanics. The point is **one** implementation of PR opening + labeling, reused rather than copied, and never a second PR for work that already has one. In `om-auto-implement-spec` these mechanics run **inside the engine skills** (`om-auto-create-pr`, `om-auto-continue-pr`/`-loop`, and through them `om-open-pr`); this file is the contract this skill routes by (step 2) and confirms at the end (step 4).

## Never open a duplicate PR

Before opening anything, check whether a PR already exists for this branch (or, in an issue-driven run, one that references the issue) via **search-prs** / **get-pr**. If one exists, **reuse it** — push new commits to its head branch and update its body/labels — never open a second PR. Only the skill that first opens the PR owns opening it; everyone else updates that same PR.

## Prefer the `om-open-pr` skill when installed

`om-open-pr` already implements exactly this: it commits the worktree, pushes the branch, opens a **ready-for-review** PR against `$BASE_BRANCH` (draft only with `--draft`) with the unified body template, applies the full SDLC label set (pipeline `review`, category, QA meta, one priority, one risk) through the descriptor guards with rationale comments, posts the caller's summary comment, and (in an issue-driven run) hands the issue back and releases the `in-progress` lock — emitting the `PR:` / `Issue:` chaining reference lines. When it is installed, the engine **delegates to it** instead of re-deriving the steps.

## Graceful fallback when `om-open-pr` is NOT installed

`om-open-pr` is an **optional** enhancement — a repo may run this pipeline without it, and it must still work. When `om-open-pr` is absent, the engine performs the mechanics inline:

1. Commit the worktree changes with a conventional-commit subject; push the branch.
2. Open the PR via the tracker operation **create-pr** against `$BASE_BRANCH`, with the engine's unified PR-body template (its mandatory `Tracking plan:` line is what lets `om-auto-continue-pr` resume).
3. Normalize labels per the section below.

Behavior is identical either way — the same PR, the same labels — so installing `om-open-pr` only removes duplication, it never changes the outcome.

## Ready vs draft

The PR ends **ready for review**. Draft only when the run is explicitly handing off incomplete work (a spec-only design PR, or an interrupted run leaving `Status: in-progress`).

## Label normalization

Labels come from the config's taxonomy, always through the `apply_label` guard from the tracker descriptor (missing labels degrade to a logged skip; `labels.enabled: false` skips everything — noted in the summary comment). This is the canonical label contract every PR-opening skill applies and step 4 of this skill confirms:

- The `review` pipeline label. PRs sit in `review` unless the run terminated early with an explicit blocker.
- `skip-qa` **only** for clearly low-risk non-user-facing changes (docs-only, dependency-only, CI-only, test-only, trivial typos, single-file maintenance).
- `needs-qa` when the run touches UI or other user-facing behavior that requires manual exercise.
- Never both `needs-qa` and `skip-qa`.
- Additive category labels when they clearly apply: `bug`, `feature`, `refactor`, `security`, `dependencies`, `documentation`.
- Exactly one priority label and exactly one risk label. The engine maintains **one** marker-idempotent consolidated rationale comment covering the whole applied label set — one label per line with its emoji and a full-sentence reason, updated in place via **update-comment** on later changes — never one comment per label.
- When `qaGate` is `true`, a `needs-qa` PR will not be mergeable until QA signs off with `qa-approved`. `qa-approved` is never added by an authoring skill — it is earned by manual QA or the self-QA exception.

## Summary comment

Every run ends with a single comprehensive summary comment the human reviewer can read top-to-bottom without clicking into the diff, posted via the tracker operation **comment-pr** with a body file so multi-line formatting is preserved. The structure is the engine's summary-comment template; it is never posted before the automated review loop finishes, never claims a completion that was not reached, and never contains secrets.

## Marker emission

End the run's final report with the chaining reference lines, one per line, exact shape — include `Issue:` only when the run has a subject issue:

```
Issue: #<issue number> (link: <full issue URL>)
PR: #<PR number> (link: <full PR URL>)
Spec: <repo-relative spec path>
```

Chained consumers (`om-auto-review-pr`, `om-auto-qa-pr`, orchestration scripts) parse these exact text markers — never rename, translate, or decorate them.

## om-auto-implement-spec specifics

This skill never opens a PR itself; it enforces the contract at two points:

- **Before invoking an engine (step 2)** — the reuse guard decides the route: an open **implementation PR** already referencing the spec (`Source doc:` / `Refs #{SPEC_PR}`) is resumed via `om-auto-continue-pr` / `-loop`; otherwise `om-auto-create-pr` opens the one implementation PR. The **spec PR is never the implementation PR** — it stays design-only, and the implementation PR links it with `Refs #{SPEC_PR}`. One implementation PR per spec, always.
- **Finish-state confirmation (step 4)** — after the engine reports the PR complete, confirm: implementation PR **ready** (the engine flips its draft PR to ready via **mark-pr-ready** once `Status: complete` — except under a `⚠ NEEDS HUMAN CONFIRMATION` assumptions guard, which keeps the PR draft for the user); the full label set above present, with user-facing PRs carrying `needs-qa` and never `qa-approved` / `qa-self-verified`; the engine's summary comment posted, with the UI-verification outcome appended to it or posted as its own evidence comment; when `ISSUE_ID` is known, the PR body carries `Closes #${ISSUE_ID}` and the plan the `Source doc:` line. Anything missing that this skill can add through the descriptor operations (a label via the guard, the UI outcome via **comment-pr**), add; anything engine-owned that is wrong, report as a defect rather than papering over.
- Then report spec path, engine used, branch, PR URL, validation/review outcome, UI verification outcome — and emit the chaining reference lines exactly as above.
