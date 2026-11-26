# PR finalize — update the existing PR, labels, lock release (step 11)

This step **updates the existing PR** — it never opens a new one. The point is **one** implementation of PR updating + labeling, reused rather than copied, and never a second PR for work that already has one.

## Never open a duplicate PR

Before touching anything, remember the reuse guard: when a PR already exists for this branch (in a resume it always does — it is the `{prNumber}` argument), **reuse it** — push new commits to its head branch and update its body/labels — never open a second PR. Only the skill that first opened the PR owns opening it; everyone else updates that same PR.

## Prefer the `om-open-pr` skill when installed

**Prefer the `om-open-pr` skill when installed** — reuse its push + label normalization instead of duplicating it. `om-open-pr` is an **optional** enhancement: when it is absent, perform the mechanics inline via the tracker operations below so this skill works standalone. Behavior is identical either way — the same PR, the same labels.

## Spec-only PRs stay design-only

A PR whose branch changes only spec/design files (`$SPECS_DIR`, docs areas, the run folder) is a **spec PR** — an atomic design deliverable. A resume never lands implementation code on it (the skill-body step 6 guard stops before that happens): implementation ships on its **own PR** via `om-auto-implement-spec`, which links the spec PR with `Refs #{prNumber}`. A branch that already mixes spec and implementation code from an earlier run is an implementation PR and is finished as one.

## Update the PR body

- If every row in the Tasks table now has `Status: done`, flip the PR body's `Status: in-progress` to `Status: complete` **and flip the PR from draft to ready via mark-pr-ready** — `om-auto-create-pr-loop` leaves the PR a draft while unfinished, so completing the resume is what promotes it. A resume that stays `in-progress` leaves the PR a draft the user can watch and re-enter.
- Extend the `What Changed` / `Tests` sections with the new work from this resume.

## Label normalization (resume state machine)

Labels — every mutation goes through the `apply_label`/`label_exists` guards from the tracker descriptor; when `labels.enabled` is `false`, skip every label operation and say so in the summary comment:

- If the PR is still in a non-terminal pipeline state (`review`, `changes-requested`, `qa`, `qa-failed`, `merge-queue`, `blocked`, `do-not-merge`), keep it. Do NOT move a PR already in `merge-queue` back to `review` just because a resume happened.
- If the PR has no pipeline label (shouldn't happen, but may after an override), apply `review`.
- Add `needs-qa` if the resume introduces user-facing behavior. Add `skip-qa` only for clearly low-risk changes. Never both. If the resume newly introduces user-facing behavior on a PR previously in `merge-queue`, add `needs-qa` and drop any stale `qa-approved` — the QA sign-off no longer covers the new work; when `qaGate` is on, the QA-approval gate re-blocks the merge until QA re-approves. Do not set the `qa` pipeline label yourself; `qa` is applied manually by a QA reviewer when they re-test.
- Preserve the priority label; raise it only when the resume materially widens scope (e.g. now touches auth, money, data scoping, or a release-blocking area) and comment why. If the PR somehow has no priority label, infer and apply one per the config taxonomy.
- Preserve the risk label; raise it only when the resume materially widens the blast radius (e.g. now touches auth, money, data scoping, migrations, encryption, event reliability, shared contracts, or spans more areas) and comment why. If the PR somehow has no risk label, infer and apply one per the config taxonomy.
- Never add `qa-approved` and never set the `qa` pipeline label from this skill. When `qaGate` is on, a `needs-qa` PR may sit in `merge-queue` while the QA-approval gate blocks the merge until a QA reviewer adds `qa-approved`; when `qaGate` is off, `needs-qa` is advisory only.
- Maintain the single consolidated `🏷️ label rationale` comment instead of per-change comments: on any label change, find the `` 🤖 `om-auto-continue-pr-loop` — 🏷️ label rationale `` marker via **list-issue-comments** and rewrite that comment via **update-comment** to the current label state — one label per line with its emoji and a full-sentence reason; create it when absent, never post a second (the format and label emoji map are the shared contract in `om-open-pr`'s label normalization).

## Final tracking-file updates before releasing the lock

- Rewrite `HANDOFF.md` one last time with either "complete" or "still in-progress — next Step: X.Y".
- Append a closing `NOTIFY.md` entry with the final status, PR URL, and any carry-forward notes.
- Commit and push as `docs(runs): finalize handoff for ${SLUG}` (or a similar message).

## Release the lock

Release the in-progress lock — **always**, even on failure (use a trap/finally): when `$LABELS_ENABLED` is `true`, remove the `in-progress` label via **unlabel-pr**; then post the completion comment via **comment-pr** (preserve multi-line formatting; `${STATUS}` is the final PR status):

```text
🤖 `om-auto-continue-pr-loop` completed. Status: ${STATUS}. Lock released.
```

## Cleanup

```bash
cd "$REPO_ROOT"
if [ "$CREATED_WORKTREE" = "1" ]; then
  git worktree remove --force "$WORKTREE_DIR"
fi
git worktree prune
```

## Marker emission

End the run's final report with the chaining reference lines, one per line, exact shape — include `Issue:` only when the run has a subject issue:

```
Issue: #<issue number> (link: <full issue URL>)
PR: #<PR number> (link: <full PR URL>)
```

Chained consumers (`om-auto-review-pr`, `om-auto-qa-pr`, orchestration scripts) parse these exact text markers — never rename, translate, or decorate them.
