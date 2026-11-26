# Run-mode contracts (step 2)

What each mode requires once step 2's classification picks it.

## Simple-run contract

For Simple runs, skip the whole run-folder ceremony. Requirements:

- **No run folder**, no `PLAN.md`, no `HANDOFF.md`, no `NOTIFY.md`, no `step-<X.Y>-checks.md`.
- **No Tasks table** anywhere.
- **One code commit** pushed to the PR branch (may be amended pre-push; once pushed, create a new commit rather than amending).
- Unit tests for behavior changes (still mandatory for code; docs-only exempt).
- Targeted validation for the touched area(s) only — the subset of `validation.commands` relevant to what changed.
- Conventional-commit subject.
- Push the fix directly to the PR branch.
- PR body stays short — summary + test plan + rollback (no `Tracking plan:` line, no `Status:` field, no linked run folder). If the existing body already has these tracking fields from a prior promotion, leave them; otherwise do not add them.
- Still respect: an isolated worktree (skip straight to step 4 for worktree setup); the three-signal `in-progress` lock (already claimed in step 1, released at run end); label discipline (pipeline + category + meta); the single step-8 `om-auto-review-pr` pass in autofix mode (compatibility checks inside it).
- Final summary comment still posts, but compacted to: summary of changes, how to verify, what can go wrong. No "Verification phases" matrix, no "External references honored" section unless actually relevant.

## Spec-implementation-run contract

Keep the full contract documented in the rest of the `SKILL.md` file: run-folder lookup, HANDOFF.md → Tasks table → NOTIFY tail orientation, lean per-Step commits, checkpoint-batched verification, full validation gate before flipping to `complete`, `om-auto-review-pr` autofix pass, comprehensive summary comment with all headings.

## Promotion path (Simple → Spec-implementation)

A Simple run MAY be promoted to a Spec-implementation run mid-flight if the resume discovers the remaining work is larger than it looked:

- Stop the simple flow.
- Draft the plan under `${RUNS_DIR}/<date>-<slug>/PLAN.md` (with Tasks table, including its `Exec` column), `HANDOFF.md`, `NOTIFY.md`.
- Write a seed commit that adds these files.
- Update the PR body to add `Tracking plan:` and `Status: in-progress` lines.
- Continue under the full Spec-implementation contract from step 3 onwards.
