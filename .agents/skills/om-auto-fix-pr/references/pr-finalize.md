# PR finalize — follow-ups, labels, summary comment, markers

The single procedure for the "file follow-ups → normalize labels → summary comment → chaining reference lines" mechanics (steps 5 and 6 of the skill body), run once the stabilization loop settles: turn non-blocking review findings into tracked issues, then leave the PR merge-ready and hand it off.

## Never open a PR (and never a duplicate)

This skill drives an **existing** PR to merge-ready; it never opens one, so there is no duplicate to guard against. The only PR ever opened in its orbit is the fork carry-forward **replacement** PR, opened by the delegated `om-auto-review-pr` flow (see the fork section below) — from that point `{prNumber}` refers to the replacement, and every later mutation targets that same PR via **get-pr** / **comment-pr**.

## Label normalization

Apply labels always through the descriptor guards (`set_pipeline_label` / `apply_label`); missing labels degrade to a logged skip; `LABELS_ENABLED=false` skips everything — note that in the summary comment. Bring the PR's labels to match its **real** state:

- **Pipeline label**: `merge-queue` when the review is approving and every required check is green; otherwise the honest state (`changes-requested`, `blocked`).
- **Draft → ready**: when the loop exit criteria are met (approvable review, green checks, UI verified or n/a) and the PR is still a draft — and not draft-by-intent (a spec-only design PR, or a `⚠ NEEDS HUMAN CONFIRMATION` guard) — promote it via **mark-pr-ready**; a merge-ready PR must not linger as a draft.
- **Label commentary**: reflect every label change in the single idempotent `` 🤖 `om-auto-fix-pr` — 🏷️ label rationale `` comment (one label per line with its emoji and a full-sentence reason, updated in place via **update-comment**; format and emoji map: the shared contract in `om-open-pr`'s label normalization) — never one comment per change.
- **QA meta**: keep `needs-qa` when user-facing behavior changed and `QA_GATE` is on. `om-auto-qa-pr` runs here in evidence-only mode, so it attaches screenshots but sets **no** QA labels; this skill **never** adds `qa-approved` or `qa-self-verified` — the QA gate and a QA reviewer own that (a human can opt into the self-QA sign-off separately). When the gate is on, a `needs-qa` PR stays unmergeable until signed off.

## Summary comment

Every run ends with a single comprehensive summary comment the human reviewer can read top-to-bottom without clicking into the diff. Post it via the tracker operation **comment-pr** with a body file so multi-line formatting is preserved. Cover: base merged in, loop outcome (review verdict, CI status, UI evidence), how each reviewer comment already on the PR was handled (fixed, deferred to a follow-up, or declined with a reason), follow-ups filed, and the merge-readiness verdict with the exact next command (`om-approve-merge-pr {prNumber}` when ready, or the blocker when not). Never post it before the loop finishes, never claim a completion you did not reach, and never paste secrets into it.

## Marker emission

End the run's final report with the chaining reference lines, one per line, exact shape — include `Issue:` only when the run has a subject issue:

```
Issue: #<issue number> (link: <full issue URL>)
PR: #<PR number> (link: <full PR URL>)
```

Chained consumers (`om-approve-merge-pr`, orchestration scripts) parse these exact text markers — never rename, translate, or decorate them.

## om-auto-fix-pr specifics

### Filing follow-ups (step 5)

Blocking findings are fixed inside the loop (step 4). Everything the review intentionally left unfixed — nits, low-severity items, out-of-scope suggestions — becomes a follow-up issue so it is tracked without holding the PR:

- For each such finding, invoke `om-followup-issue-from-pr` verbatim with the PR link (or, when the finding lives in a specific review comment, that comment's link) so its comment-mode extracts the actionable ask and assigns the issue to the right person (the comment's @-mention, else the PR author).
- **Idempotent**: before filing, check whether a follow-up for this finding already exists (scan the PR's existing follow-up comments/links and open issues that reference this PR). Never double-file the same nit across loop iterations or re-runs.
- Batch the summary: list every follow-up issue opened in the final summary comment so the reviewer sees what was deferred and why.

Do **not** use follow-ups to dodge real blockers — a correctness, security, compatibility, or failing-gate finding is fixed in the loop, never deferred to a follow-up issue.

### Merge handoff (step 6) — never merge

Once labels and metadata match the real state, **hand off**: this skill leaves the PR merge-ready and stops. The merge itself belongs to `om-approve-merge-pr` (single PR) or `om-merge-buddy` (sweep), which re-check every gate before squash-merging.

### Fork supersede/credit

When step 4's `om-auto-review-pr` run carried a fork PR forward (it bases a carry branch in the main repo on the fetched PR head — e.g. `carry/pr-{prNumber}-ready` — preserving the original commits and authorship, applies merges/fixes there, opens a replacement PR, and closes the original only after the replacement exists), confirm the replacement PR satisfies the Supersede Credit Rule so changelog tooling credits the original contributor, not the reviewer:

- Its body starts (within the first 20 lines) with `Supersedes #{originalPr}` — changelog tooling matches `^Supersedes\s+#(\d+)\b` (case-insensitive).
- It carries an explicit credit line — `Credit: original implementation by @{originalAuthor}. …` — matched as `Credit:\s+original\s+implementation\s+by\s+@handle`.
- It is reassigned to the original author, with a handoff comment inviting them to do the next recheck from the carried-forward branch.
- The closed original PR carries a `Closing in favor of #{newPrNumber} ({newPrUrl}).` comment crediting the original author (the fallback detection path for the credit rule).

If any of these are missing, fix them here (edit the replacement PR body, post the comments, reassign) before declaring merge-ready.
