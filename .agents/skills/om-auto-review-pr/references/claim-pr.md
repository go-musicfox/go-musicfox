# Claiming and releasing work — issues and PRs

The generalized claim/lock procedure for any tracker item (issue or PR) an autonomous run takes ownership of. All reads and mutations go through the tracker descriptor's operations; label mutations only through the `apply_label` guard.

## Three-signal in-progress check

Resolve `CURRENT_USER` via the tracker operation **current-user**, then read the item (for PRs via **get-pr**; for issues via the descriptor's issue-read operation). The item counts as **already in progress** when ANY of these signals holds:

1. **Assignee** — the item is assigned to someone other than `CURRENT_USER`.
2. **`in-progress` label** — the label is present on the item.
3. **Recent robot claim comment** — a `🤖` claim comment posted within the stale window by someone other than `CURRENT_USER`.

Decision:

- No signal → claim and proceed.
- Signals point at `CURRENT_USER` → re-entry into your own run; refresh the claim (idempotent) and continue.
- Signals point at someone else → **STOP** and report the owner — unless the lock is stale (below) or `--force` was passed.

**`ci-monitoring` is not a lock signal.** It is a meta label meaning the previous run's work is *finished and fully reported* — labels applied, review submitted, comments posted — and that run is only watching CI, so it still owes a CI-result follow-up comment. An item carrying `ci-monitoring` **and none of the three signals above** is **not** in progress: claim it and proceed normally, without `--force` and without an override comment. Never fold `ci-monitoring` into the three-signal check, and never treat it as a reason to back off; the whole point of the label is that a monitoring process which dies leaves an honest state rather than a lock nobody holds. When `in-progress` and `ci-monitoring` are both present, the `in-progress` signal decides — `ci-monitoring` neither adds to it nor cancels it.

## Stale-lock recovery

A claim is **stale** when the newest claim signal is older than the stale window and the claimant has produced no commits, comments, or label changes on the item since. Never silently overwrite a live claim.

## `--force` override

`--force` bypasses the conflict stop, never the transparency: post an override comment naming the previous owner and why the override happened (`🤖 --force override: taking over from {owner} — {reason}.`), then apply the claim. Document the override in the run's report.

## Applying the claim

Idempotent — safe to re-run on re-entry:

1. Assign `CURRENT_USER` to the item via the descriptor's assign operation.
2. Apply the `in-progress` label via the `apply_label` guard (missing label → logged skip; `labels.enabled: false` → skip and note it in the report).
3. Post the claim comment, once (skip when an identical recent comment by `CURRENT_USER` already exists).

## Release / handback

When the run finishes, hands off, or aborts:

- Remove the `in-progress` label via the guard — and when the run intends to follow up with CI results after reporting, **swap** rather than simply remove: `apply_label "ci-monitoring"` in the same breath, so the item is never observably unlabeled and the outstanding follow-up is visible. Remove `ci-monitoring` when the CI-result comment is finally posted, or when the CI wait is abandoned at its `ci.maxWaitMinutes` cap (nobody is monitoring then, and leaving the label would promise a follow-up that never comes). A run that does no CI follow-up at all just removes `in-progress` as before and never applies `ci-monitoring`.
- Post a short release comment stating the outcome.
- The claimant releases their own claim — never release a lock another agent holds.

## Chained hand-off — a live chain never drops its lock

When the same run (same `CURRENT_USER`) finishes one skill and continues on the same item with another — `om-open-pr` → `om-auto-review-pr`, review → UI QA, or any flow-runner chain — the lock is **transferred, never released and re-acquired**. A release-then-reclaim seam leaves the item observably unclaimed mid-run: any concurrent actor's three-signal check reads "not in progress" and legitimately starts duplicate work, and humans watching the tracker see no owner and no state.

- **Hand-off (finishing step):** keep the `in-progress` label and lock assignee in place; instead of the release comment, post a hand-off comment naming the next phase:

  `` 🤖 `{finishing-skill}` completed: {outcome}. Lock handed off to `{next-skill}` — chain continues on this {issue|PR}. ``

- **Take-over (next step):** the three-signal check finds the lock held by `CURRENT_USER` → re-entry. **Before any other work** — fetching diffs, running validation, posting findings — refresh the claim comment so the tracker always shows who holds the item and why:

  `` 🤖 `{next-skill}` taking over the chain lock — {phase}. Started: {ISO-8601 timestamp}. ``

- **Ownership:** a skill releases only a lock its own run opened. An inherited (handed-off) lock is annotated in the completion comment (`Lock retained — chain continues.`) and released by the chain's driving skill at the end of the run, or by its failure path — "the claimant releases their own claim" applies to the chain as a whole.
- **Crash recovery (adoption):** a hand-off lock is live only while its chain is running. A **standalone** run (one not invoked as a chain step) that re-enters a same-`CURRENT_USER` lock whose newest 🤖 claim/take-over/hand-off comment is older than the stale window treats the chain as dead: post an adoption note — `` 🤖 Adopting a stale chain lock ({age}) — previous run presumed dead. `` — then own the lock as if this run opened it, releasing it at the end. Chained invocations never adopt; their driver owns release.
- **Invariant:** an item under active automation is never observably unclaimed — the claim or take-over comment precedes any work product, and the hand-off or release is the step's last tracker mutation.

## om-auto-review-pr specifics

This skill claims the PR itself (step 1 of the skill body) and uses tighter windows than the generic defaults.

**Detection.** Run **current-user** to fill `CURRENT_USER`, then **get-pr** for `{prNumber}` requesting `assignees`, `labels`, `number`, `title`, and `comments`. The PR is **already in progress** when it carries the `in-progress` label, has at least one assignee whose login is not `$CURRENT_USER`, or has a claim comment newer than **30 minutes** from another actor (look for the `🤖` start marker).

Decision tree:

| State | `--force` set? | Action |
|-------|---------------|--------|
| Not in progress | — | Claim and proceed |
| In progress, current user owns the lock | — | Re-entry (incl. a chain hand-off lock from `om-open-pr --handoff` or a flow runner's outer claim): post the take-over comment naming this skill **before any review work**, then proceed without re-claiming |
| In progress, someone else owns the lock | no | **STOP**. Ask the user: "PR #{prNumber} is in progress (owner: {owner}, signal: {label/assignee/comment}). Override and continue?" Only continue when the user explicitly says yes. |
| In progress, someone else owns the lock | yes | Post a force-override comment naming the previous owner, then claim and proceed |

**Stale lock**: if the `in-progress` label is older than **60 minutes** and the assignee did not push or comment in that window, treat it as expired. Still ask the user before overriding unless `--force` was set.

**Claim** (only after the check passes) — three tracker operations:

1. **assign-pr**: add `$CURRENT_USER` as an assignee on `{prNumber}`.
2. Run `apply_label "in-progress" {prNumber}`.
3. Post the claim comment via **comment-pr**, filling in `$CURRENT_USER` and the current UTC timestamp (ISO-8601, e.g. `date -u +%Y-%m-%dT%H:%M:%SZ`):

```text
🤖 `om-auto-review-pr` started by @{CURRENT_USER} at {timestamp}. Other auto-skills will skip this PR until the lock is released.
```

When `labels.enabled` is `false`, the claim consists of the assignee plus the claim comment — other skills detect those two signals.

**Release** (step 12 of the skill body — always, even on failure; use a `trap` or equivalent finally-block so a crash or early stop still clears the lock):

1. When `LABELS_ENABLED` is `true`, remove the `in-progress` label from `{prNumber}` via the tracker operation **unlabel-pr**.
2. Post the lock-release comment via **comment-pr**, where `VERDICT` is the decision from steps 9–10 (`APPROVED` or `CHANGES REQUESTED`):

```text
🤖 `om-auto-review-pr` completed: {VERDICT}. Lock released.
```

The completion comment carries the verdict plus a short summary (and, when autofix ran, how many fix iterations completed). For `changes-requested`, the assignee is already handed back to the author before release; for approved outcomes, keep the current assignee unless a handoff changed it.

**Inherited chain lock (re-entry).** When step 1 found the lock already held by `$CURRENT_USER` because a chain handed it off (see the chained hand-off section above), this run did not open the claim and must not release it: skip the **unlabel-pr**, keep the assignee, and post the completion comment as:

```text
🤖 `om-auto-review-pr` completed: {VERDICT}. Lock retained — chain continues.
```

The chain's driving skill (e.g. `om-auto-fix-issue`) releases the lock at the end of its run.
