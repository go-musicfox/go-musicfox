# Claiming and releasing work — issues and PRs

The generalized claim/lock procedure for any tracker item (issue or PR) an autonomous run takes ownership of. All reads and mutations go through the tracker descriptor's operations; label mutations only through the `apply_label` guard. This complements the skill-specific slot check (run folder / branch / open PR — see the specifics section below) — the slot check decides whether the *work* exists, this procedure decides who *owns* it.

## Three-signal in-progress check

Resolve `CURRENT_USER` via the tracker operation **current-user**, then read the item (for PRs via **get-pr**; for issues via the descriptor's issue-read operation). The item counts as **already in progress** when ANY of these signals holds:

1. **Assignee** — the item is assigned to someone other than `CURRENT_USER`.
2. **`in-progress` label** — the label is present on the item.
3. **Recent robot claim comment** — a claim comment in the format below, posted within the stale window (default 24 h) by someone other than `CURRENT_USER`.

Decision:

- No signal → claim and proceed.
- Signals point at `CURRENT_USER` → re-entry into your own run; refresh the claim (idempotent) and continue.
- Signals point at someone else → **STOP** and report the owner — unless the lock is stale (below) or `--force` was passed.

**`ci-monitoring` is not a lock signal.** It is a meta label meaning the previous run's work is *finished and fully reported* — labels applied, review submitted, comments posted — and that run is only watching CI, so it still owes a CI-result follow-up comment. An item carrying `ci-monitoring` **and none of the three signals above** is **not** in progress: claim it and proceed normally, without `--force` and without an override comment. Never fold `ci-monitoring` into the three-signal check, and never treat it as a reason to back off; the whole point of the label is that a monitoring process which dies leaves an honest state rather than a lock nobody holds. When `in-progress` and `ci-monitoring` are both present, the `in-progress` signal decides — `ci-monitoring` neither adds to it nor cancels it.

## Stale-lock recovery

A claim is **stale** when the newest claim signal is older than the stale window (24 h) and the claimant has produced no commits, comments, or label changes on the item since. Recover by posting a takeover note first — `🤖 Previous claim by {owner} appears stale ({age}); taking over.` — then claim normally. Never silently overwrite a live claim.

## `--force` override

`--force` bypasses the conflict stop, never the transparency: post an override comment naming the previous owner and why the override happened (`🤖 --force override: taking over from {owner} — {reason}.`), then apply the claim. Document the override in the run's plan/report.

## Applying the claim

Idempotent — safe to re-run on re-entry:

1. Assign `CURRENT_USER` to the item via the descriptor's assign operation.
2. Apply the `in-progress` label via the `apply_label` guard (missing label → logged skip; `labels.enabled: false` → skip and note it in the report).
3. Post the claim comment, once (skip when an identical recent comment by `CURRENT_USER` already exists):

   `` 🤖 Claiming this {issue|PR} — starting `{skill-name}` run. Started: {ISO-8601 timestamp}. ``

## Release / handback

When the run finishes, hands off, or aborts:

- Remove the `in-progress` label via the guard — and when the run intends to follow up with CI results after reporting, **swap** rather than simply remove: `apply_label "ci-monitoring"` in the same breath, so the item is never observably unlabeled and the outstanding follow-up is visible. Remove `ci-monitoring` when the CI-result comment is finally posted, or when the CI wait is abandoned at its `ci.maxWaitMinutes` cap (nobody is monitoring then, and leaving the label would promise a follow-up that never comes). A run that does no CI follow-up at all just removes `in-progress` as before and never applies `ci-monitoring`.
- In issue-driven runs, hand the issue back: restore the original assignee/author when the pipeline convention expects it.
- Post a short release comment stating the outcome (PR opened with its number, blocked with the blocker, or no action needed).
- The claimant releases their own claim — never release a lock another agent holds. A sub-skill that claims for itself (e.g. a standalone `om-auto-review-pr` run) owns its own release; under a chained hand-off the sub-skill instead re-enters and retains the lock (see Chained hand-off below). Do not second-guess either.

## Chained hand-off — a live chain never drops its lock

When the same run (same `CURRENT_USER`) finishes one skill and continues on the same item with another — `om-open-pr` → `om-auto-review-pr`, review → UI QA, or any flow-runner chain — the lock is **transferred, never released and re-acquired**. A release-then-reclaim seam leaves the item observably unclaimed mid-run: any concurrent actor's three-signal check reads "not in progress" and legitimately starts duplicate work, and humans watching the tracker see no owner and no state.

- **Hand-off (finishing step):** keep the `in-progress` label and lock assignee in place; instead of the release comment, post a hand-off comment naming the next phase:

  `` 🤖 `{finishing-skill}` completed: {outcome}. Lock handed off to `{next-skill}` — chain continues on this {issue|PR}. ``

- **Take-over (next step):** the three-signal check finds the lock held by `CURRENT_USER` → re-entry. **Before any other work** — fetching diffs, running validation, posting findings — refresh the claim comment so the tracker always shows who holds the item and why:

  `` 🤖 `{next-skill}` taking over the chain lock — {phase}. Started: {ISO-8601 timestamp}. ``

- **Ownership:** a skill releases only a lock its own run opened. An inherited (handed-off) lock is annotated in the completion comment (`Lock retained — chain continues.`) and released by the chain's driving skill at the end of the run, or by its failure path — "the claimant releases their own claim" applies to the chain as a whole.
- **Crash recovery (adoption):** a hand-off lock is live only while its chain is running. A **standalone** run (one not invoked as a chain step) that re-enters a same-`CURRENT_USER` lock whose newest 🤖 claim/take-over/hand-off comment is older than the stale window treats the chain as dead: post an adoption note — `` 🤖 Adopting a stale chain lock ({age}) — previous run presumed dead. `` — then own the lock as if this run opened it, releasing it at the end. Chained invocations never adopt; their driver owns release.
- **Invariant:** an item under active automation is never observably unclaimed — the claim or take-over comment precedes any work product, and the hand-off or release is the step's last tracker mutation.

## om-auto-create-pr-loop specifics

### Pre-flight run-ownership (slot) check — step 2

Before writing anything, confirm no other run owns the slot. Resolve `CURRENT_USER` via **current-user**, then compute:

```bash
DATE=$(date +%Y-%m-%d)
SLUG="{slug-or-derived}"
RUN_DIR="${RUNS_DIR}/${DATE}-${SLUG}"
PLAN_PATH="${RUN_DIR}/PLAN.md"
HANDOFF_PATH="${RUN_DIR}/HANDOFF.md"
NOTIFY_PATH="${RUN_DIR}/NOTIFY.md"
# Verification is checkpoint-based: ${RUN_DIR}/checkpoint-<N>-checks.md every ~5 Steps.
# Optional artifacts (browser transcripts, screenshots) live at ${RUN_DIR}/checkpoint-<N>-artifacts/.
# Final gate log lives at ${RUN_DIR}/final-gate-checks.md at spec completion.
BRANCH_PREFIX="{fix for bugfix/remediation work; otherwise feat}"
BRANCH="${BRANCH_PREFIX}/${SLUG}"
```

Branch naming: use `fix/${SLUG}` when the brief is primarily a bug fix, regression fix, remediation, hardening task, or corrective follow-up; use `feat/${SLUG}` for new capability, scoped refactors, docs/process automation, or anything not primarily corrective.

A run is **already in progress** when ANY of the following is true:

- A folder at `$RUN_DIR` (or a legacy flat file `${RUN_DIR}.md`) already exists on `origin/$BASE_BRANCH` or any remote branch.
- A remote branch `origin/${BRANCH}` already exists.
- An open PR already references `$RUN_DIR` or `$PLAN_PATH` (check via **search-prs** with the run-folder path as the query, or by scanning open PRs via **list-prs**).

Decision tree:

| State | `--force` set? | Action |
|-------|---------------|--------|
| Nothing exists | — | Claim and proceed. |
| Run folder/branch exists, current user owns it | — | Treat as re-entry; hand off to `om-auto-continue-pr-loop` and stop. |
| Run folder/branch exists, someone else owns it | no | **STOP.** Ask the user: "Run folder/branch for `${SLUG}` already exists (owner: ${owner}). Override and continue?" Only continue when the user explicitly says yes. |
| Run folder/branch exists, someone else owns it | yes | Pick a new dated slug (`${SLUG}-v2` or append a time suffix) to avoid clobber; document in the new `PLAN.md` why the original was superseded. |

When an open PR already references the run folder, stop and tell the user to use `om-auto-continue-pr-loop {prNumber}` instead.

### PR lock lifecycle — steps 7 → 14

This skill holds the three-signal lock on its own PR from the moment the draft PR opens (step 7) until the very end of the run:

1. **Claim (step 7, immediately after opening the draft PR):**
   1. **assign-pr** — add `$CURRENT_USER` as assignee.
   2. **label-pr** — apply `in-progress` through the `apply_label` guard (when `labels.enabled` is `false`, the claim is the assignee plus the claim comment).
   3. **comment-pr** — post: `` 🤖 `om-auto-create-pr-loop` started by @{CURRENT_USER} at {UTC ISO-8601 timestamp}. Other auto-skills will skip this PR until the lock is released. ``
   Wire the release into a `trap`/finally from here (step 14) so a crash frees the PR.
2. **Temporary release before the autofix pass (step 12):** **unlabel-pr** removes `in-progress` through the guard, then **comment-pr** posts: `` 🤖 `om-auto-create-pr-loop` releasing lock so `om-auto-review-pr` can claim it. `` — `om-auto-review-pr` claims and releases per its own workflow.
3. **Reclaim when it returns (step 12):** **label-pr** re-applies `in-progress` through the guard, then **comment-pr** posts: `` 🤖 `om-auto-create-pr-loop` reclaiming lock to post the final run summary. `` — covering the summary + cleanup window.
4. **Final release (step 14, always — even on failure):** **unlabel-pr** removes `in-progress` through the guard (tolerate failure), then **comment-pr** posts: `` 🤖 `om-auto-create-pr-loop` completed. Status: {complete | in-progress}. Lock released. ``

Executor subagents (see `references/executor-dispatch.md`) MUST NOT claim or release this lock — the main session owns it.
