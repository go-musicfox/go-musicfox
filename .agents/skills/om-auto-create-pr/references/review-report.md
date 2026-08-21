# Review and report — authoritative automated review pass

Detailed procedure for step 10, the single code-review/autofix loop of `om-auto-create-pr`.

## Automated review with `om-auto-review-pr` (step 10)

Before you post the final summary comment, push the last commits, or report back, subject the PR to its authoritative review with the `om-auto-review-pr` skill. Do not run a separate direct `om-code-review` pass first: `om-auto-review-pr` invokes that engine verbatim and adds the autofix loop.

`om-auto-create-pr` may not hold an `in-progress` lock on the PR at this point (only `om-auto-continue-pr` does). `om-auto-review-pr`'s claim check runs **first, before any review work**, in every invocation mode: it either claims the PR fresh (`in-progress` label + claim comment — the PR must never be under review while observably unclaimed) or, when the lock is already held by `$CURRENT_USER`, re-enters with a take-over comment and leaves release to the lock's owner. It releases only a claim its own run opened, per its workflow (see `references/claim-pr.md`, chained hand-off). Do not second-guess its claim/release protocol.

Invoke the `om-auto-review-pr` skill against `{prNumber}` in autofix mode:

1. Follow the entire `om-auto-review-pr` workflow verbatim — do not cherry-pick steps.
2. When it flags actionable issues, apply fixes directly in the same worktree used for this run. Never rewrite earlier commits; always add new commits.
3. After each batch of fixes:
   - Re-run the targeted validation for the changed areas.
   - Re-run the full validation gate from step 8 whenever a fix touches code outside a single module/test file.
   - Update the plan's **Progress** section if the fix corresponds to a plan Step (flip `- [ ]` to `- [x]` with the commit SHA); otherwise add a short note under the relevant Phase heading in the plan (e.g. `- [x] Post-review fix: {one-line summary} — {sha}`).
   - Commit using a clear conventional-commit subject (e.g. `fix(ui): address review feedback on confirmation dialog focus trap`). Push immediately.
4. Loop until `om-auto-review-pr` returns a clean verdict (no actionable blockers) or the remaining findings are non-actionable (out-of-scope, false positive) and explicitly documented in the summary comment you post in step 11.

## Verdict handling

- **Clean verdict** → proceed to the summary comment (step 11); the summary names the verdict and the SHA range of any follow-up commits.
- **Only non-actionable findings remain** → proceed, but list each remaining finding and why it is out of scope or a false positive in the summary comment.
- **Cannot run** (e.g., required checks not yet green, missing context) → escalate: leave `Status: in-progress` in the PR body, stop here, and report the blocker to the user so they can decide whether to resume via `om-auto-continue-pr`.
