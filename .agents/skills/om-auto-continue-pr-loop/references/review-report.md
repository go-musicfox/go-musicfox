# Review and report — authoritative automated review pass

Detailed procedure for the single automated review loop of `om-auto-continue-pr-loop`.

## Automated review with `om-auto-review-pr`

Before you post the final summary comment, push the final changes, or flip the PR body to `complete`, subject the resumed PR to its authoritative review with the `om-auto-review-pr` skill. Do not run a separate direct `om-code-review` pass first: `om-auto-review-pr` invokes that engine verbatim and applies compatibility, security, contract, breaking-change, and scope checks.

Invoke the `om-auto-review-pr` skill against `{prNumber}` in autofix mode:

1. Follow the entire `om-auto-review-pr` workflow verbatim — do not cherry-pick steps.
2. Apply fixes directly in the same worktree used for this resume. Never rewrite earlier commits; always add new commits.
3. Loop until `om-auto-review-pr` returns a clean verdict or the remaining findings are non-actionable (out-of-scope, false positive) and explicitly documented in the summary comment you post in step 9.

## Verdict handling

- **Clean verdict** → proceed to the summary comment (step 9); the summary names the verdict and the SHA range of any follow-up commits.
- **Only non-actionable findings remain** → proceed, but list each remaining finding and why it is out of scope or a false positive in the summary comment.
- **Cannot run** (e.g., required checks not yet green, missing context) → stop here, leave `Status: in-progress` in the PR body, update `HANDOFF.md` + `NOTIFY.md` with the blocker, and tell the user how to re-enter.

## om-auto-continue-pr-loop specifics

- The claim check inside `om-auto-review-pr` will recognize that the current user already owns the `in-progress` lock (claimed in step 1), so it proceeds as re-entry without re-claiming.
- Each review fix lands as a **new lean Step** appended to the Tasks table under a fresh `X.Y-review-fix` id: one commit, flip the Tasks row in the same commit, no per-Step checks/handoff files, conventional-commit subject (e.g. `fix(ui): address review feedback on confirmation dialog focus trap`). Push immediately.
- After each batch of fixes:
  - Run a quick scratch sanity check (typecheck + affected tests, or the closest configured equivalent).
  - When the batch closes — or every 5 review-fix Steps, whichever comes first — run a checkpoint pass per step 6b (targeted validation, focused integration tests + screenshots if UI was touched, write `checkpoint-<N>-checks.md`, rewrite `HANDOFF.md`, append NOTIFY entry, commit as `docs(runs): checkpoint N — review fixes`).
  - Re-run the full final gate from step 7 whenever a fix touches code outside a single module/test file.
