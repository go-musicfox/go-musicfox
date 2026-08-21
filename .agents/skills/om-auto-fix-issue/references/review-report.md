# Review and report — the automated review loop

Detailed procedure for step 9 of `om-auto-fix-issue` (bug route): subject the fresh PR to the same scrutiny an incoming PR would get. (This skill performs no separate review step of its own; `om-auto-review-pr` is the bug route.s single authoritative review, and the feature route.s delegated engine owns its single review loop.)

## Automated pass with `om-auto-review-pr` (step 9)

Before the final report, subject the PR to an automated pass with the `om-auto-review-pr` skill. This is the equivalent of a peer reviewer catching issues the implementation missed.

The PR already carries the chain's `in-progress` lock: `om-open-pr` transferred it there (`--handoff om-auto-review-pr`) *before* releasing the issue lock, so at no point between chain steps is the PR observably unclaimed. `om-auto-review-pr`'s claim check therefore finds the lock held by `$CURRENT_USER`, treats it as re-entry, and posts its take-over comment **before any review work** (see `references/claim-pr.md`, chained hand-off). When it finishes it keeps the inherited lock (`Lock retained — chain continues.`) — `om-auto-fix-issue` releases the PR lock exactly once, at the end of the run (step 12). Do not second-guess its claim protocol, and do not let review work begin before its take-over comment is posted.

Invoke the `om-auto-review-pr` skill against `PR_NUMBER` in autofix mode:

1. Follow the entire `om-auto-review-pr` workflow verbatim — do not cherry-pick steps.
2. When it flags actionable issues, apply fixes directly in the same worktree used for this run, as new commits. Never rewrite history.
3. After each batch of fixes, re-run the targeted validation for the changed areas, and the full configured validation gate whenever a fix reaches beyond a single module/test file. Push after each batch.
4. Loop until `om-auto-review-pr` returns a clean verdict (no actionable blockers) or the remaining findings are non-actionable (out of scope, false positive) — document those explicitly in a PR comment.

## Verdict handling

- **Clean verdict** → proceed to cleanup and the final report; name the verdict and any follow-up commits there.
- **Only non-actionable findings remain** → proceed, but list each remaining finding and why it is out of scope or a false positive in the PR comment and the final report.
- **Cannot run** (e.g., required checks not yet reported, missing context) → skip the loop, release the chain's PR lock with a comment explaining why (an idle locked PR would block the later sweep), note it in the final report, and leave the PR in the `review` pipeline state for a human or a later `om-review-prs` sweep.
