# Claiming and releasing work — issues and PRs

The generalized claim/lock procedure for any tracker item (issue or PR) an autonomous run takes ownership of. All reads and mutations go through the tracker descriptor's operations; label mutations only through the `apply_label` guard. This complements the skill-specific slot check in the skill body — the slot check decides whether the *work* exists, this procedure decides who *owns* it.

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
3. Post the claim comment, once (skip when an identical recent comment by `CURRENT_USER` already exists).

## Release / handback

When the run finishes, hands off, or aborts:

- Remove the `in-progress` label via the guard — and when the run intends to follow up with CI results after reporting, **swap** rather than simply remove: `apply_label "ci-monitoring"` in the same breath, so the item is never observably unlabeled and the outstanding follow-up is visible. Remove `ci-monitoring` when the CI-result comment is finally posted, or when the CI wait is abandoned at its `ci.maxWaitMinutes` cap (nobody is monitoring then, and leaving the label would promise a follow-up that never comes). A run that does no CI follow-up at all just removes `in-progress` as before and never applies `ci-monitoring`.
- In issue-driven runs, hand the issue back: restore the original assignee/author when the pipeline convention expects it.
- Post a short release comment stating the outcome (PR opened with its number, blocked with the blocker, or no action needed).
- The claimant releases their own claim — never release a lock another agent holds. A sub-skill that claims for itself owns its own release; do not second-guess it.

## Chained hand-off — a live chain never drops its lock

When the same run (same `CURRENT_USER`) finishes one skill and continues on the same item with another — `om-open-pr` → `om-auto-review-pr`, review → UI QA, or any flow-runner chain — the lock is **transferred, never released and re-acquired**. A release-then-reclaim seam leaves the item observably unclaimed mid-run: any concurrent actor's three-signal check reads "not in progress" and legitimately starts duplicate work, and humans watching the tracker see no owner and no state.

- **Hand-off (finishing step):** keep the `in-progress` label and lock assignee in place; instead of the release comment, post a hand-off comment naming the next phase:

  `` 🤖 `{finishing-skill}` completed: {outcome}. Lock handed off to `{next-skill}` — chain continues on this {issue|PR}. ``

- **Take-over (next step):** the three-signal check finds the lock held by `CURRENT_USER` → re-entry. **Before any other work** — fetching diffs, running validation, posting findings — refresh the claim comment so the tracker always shows who holds the item and why:

  `` 🤖 `{next-skill}` taking over the chain lock — {phase}. Started: {ISO-8601 timestamp}. ``

- **Ownership:** a skill releases only a lock its own run opened. An inherited (handed-off) lock is annotated in the completion comment (`Lock retained — chain continues.`) and released by the chain's driving skill at the end of the run, or by its failure path — "the claimant releases their own claim" applies to the chain as a whole.
- **Crash recovery (adoption):** a hand-off lock is live only while its chain is running. A **standalone** run (one not invoked as a chain step) that re-enters a same-`CURRENT_USER` lock whose newest 🤖 claim/take-over/hand-off comment is older than the stale window treats the chain as dead: post an adoption note — `` 🤖 Adopting a stale chain lock ({age}) — previous run presumed dead. `` — then own the lock as if this run opened it, releasing it at the end. Chained invocations never adopt; their driver owns release.
- **Invariant:** an item under active automation is never observably unclaimed — the claim or take-over comment precedes any work product, and the hand-off or release is the step's last tracker mutation.

## om-fix specifics

Claiming the issue is step 1 of this skill — the only tracker-state mutation before PR-open. Run the claim once, up front, so any parallel automation sees the lock immediately:

- The chain driver (`om-auto-fix-issue`, via `om-verify-in-repo`) already ran the three-signal in-progress check before this step started; om-fix does not re-litigate ownership — it applies the idempotent claim on `{issueId}`.
- The claim carries all three signals: **assign-issue** to `$CURRENT_USER`, **label-issue** applying `in-progress` through the guard (honors `labels.enabled` and label existence; a missing label degrades to a logged skip), and **comment-issue** posting the claim comment. The assignee and comment are applied regardless of `labels.enabled`.
- Claim comment format (exact):

  ```
  🤖 `autofix` started by @${CURRENT_USER} at <UTC timestamp>. Other auto-skills will skip this issue until the lock is released.
  ```

- Claim failures are non-fatal — log and continue.
- **Do not release the lock here.** The release happens in `om-open-pr` (success path) or via an external janitor (failure path); on `Status: blocked` the lock intentionally remains set so a human can pick it up.
