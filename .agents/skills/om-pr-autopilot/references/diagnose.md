# Diagnose — the ten state signals

Read-only. Collect every signal below before classifying; a signal you did not
read is `unknown`, never an assumption. Every read goes through a **named
tracker operation** from `$TRACKER_FILE` — this skill never issues tracker CLI
commands directly, so it works against any configured tracker.

## 1. Identity and repository

Run **current-user** and **repo-info**. Record the active account and the
repository the run targets.

The run must operate as the identity this repository's automation runs are made
from. Resolve that identity — never hard-code an account name; a repo-local
override may name it, otherwise **current-user** is authoritative. An active
account that does not match is a **stop condition**: report it and mutate
nothing.

## 2. PR core

Run **get-pr** for `{prNumber}` with the fields
`number,title,url,body,state,author,isDraft,baseRefName,headRefName,headRepository,headRepositoryOwner,isCrossRepository,maintainerCanModify,mergeable,mergeStateStatus,reviewDecision,labels,latestReviews,assignees,commits,files,closingIssuesReferences`.

Derive:

- `STATE` — stop unless the PR is open.
- `IS_MINE` — the PR author is `$CURRENT_USER`.
- `IS_FORK` — the head branch lives in a different repository.
- `PUSHABLE` — `!IS_FORK || IS_MINE` (you can push to your own fork). This, not
  `IS_FORK` alone, decides whether the carry-forward flow applies.
- `DRIVABLE` — `IS_MINE`, or the user explicitly asked for the autofix chain on
  this PR. Separate from `PUSHABLE` on purpose: `PUSHABLE` answers *can I push*
  and `DRIVABLE` answers *may I*. A colleague's PR on a branch in the main
  repository is `PUSHABLE` and not `DRIVABLE`, and every chain step that puts
  commits on the head branch requires `DRIVABLE`.
- `IS_DRAFT`.

## 3. Plan progress (the "how far is the implementation" signal)

Take the first `Tracking plan:` / `Tracking run folder:` line from the PR body
returned by step 2. Two contracts exist and they are **not** interchangeable —
record which one applies:

- **Plan file under `$RUNS_DIR`** (the config's `paths.runs`) → read its
  `## Progress` section; count `- [ ]` (pending) against `- [x]` (done) and note
  the first pending step. This is the `om-auto-continue-pr` contract.
- **Run folder** (holds a `PLAN.md` plus a `HANDOFF.md`) → parse the
  top-of-file `## Tasks` table in `PLAN.md`; the first row whose `Status` is not
  `done` is the resume point. This is the `om-auto-continue-pr-loop` contract.
- **No tracking line** → look in the PR's changed files (step 4) for a new file
  under `$RUNS_DIR`. Exactly one candidate → use it. Several → pick none and
  record the ambiguity. None → record `plan: none`: the PR did not come from the
  auto-create chain, no continue step can run, and implementation completeness
  must be judged from the diff against the linked issue instead.

## 4. Diff scope

Run **get-pr-files** (or **get-pr-diff** when the patch itself is needed) for
`{prNumber}`.

Classify the diff as **spec-only** (touches only `$SPECS_DIR`, the config's
`paths.specs` — read it, never hard-code the path), **docs-only**,
**UI-touching** (rendered components, pages, or portal surfaces as the
repository's agent instructions define them), **migration/schema**, or
**contract surface** (exported types, event ids, API routes, DI keys, permission
ids — per the repository's backward-compatibility document when it has one).
Spec-only and UI-touching change the chain; the rest drive the label set.

## 5. Review decision and unresolved conversations

Take `reviewDecision` from step 2, then establish whether review conversations
are still open. Prefer the tracker's own review-thread data when the descriptor
exposes it (resolution and outdated state per thread); count only threads that
are **unresolved and not outdated**, and record their file paths. When the
descriptor exposes no thread-level operation, do not invent one: derive the open
set from the `reviews` / `latestReviews` already fetched in signal 2 — the
newest review per author, with a `CHANGES_REQUESTED` decision standing in for
the unresolved threads — and record `conversations: approximate (no thread-level
operation)`. When even that is unavailable, record `conversations: unknown`.
Never resolve a thread id through **get-review-comment**: that operation maps an
id you already hold to its body, and with no thread-level operation there is no
id to hand it.

A `CHANGES_REQUESTED` decision, or any unresolved non-outdated thread, means the
review loop still has work — even when CI is green.

## 6. CI

Run **get-pr-checks** for `{prNumber}` and **get-required-checks** for the PR's
base branch. When required checks are unreadable (the operation reports the
branch protection as unavailable), treat every reported check as required and
say so. Record `ci: green | red(<names>) | pending | none`, and keep the link to
each failing check for the chain step that will fix it. **Pending is a state to
record and disclose, never a state to wait out here** — the diagnosis is
read-only and the chain reports on what it finds; any waiting happens once,
bounded, after step 6 has published (`references/ci-followup.md`).

## 7. Mergeability

`mergeable` plus `mergeStateStatus` from step 2. A conflicting state means the
base merge must run first; a behind state means the base advanced; a blocked
state means a gate (review or a required check) is unmet — not necessarily a
conflict.

## 8. Labels

From step 2's `labels`. Record the current pipeline label (they are mutually
exclusive), the category, meta, priority, and risk labels present, and which are
**missing** relative to what the diff warrants — the repository's agent
instructions carry the inference rules, and `references/report-templates.md`
reproduces the derivation this skill uses. A non-draft PR with no priority or no
risk label is a finding.

## 9. QA evidence

The QA meta labels from step 2, plus a scan of the PR conversation via
**list-issue-comments** for attached screenshots or a written self-QA
confirmation. When `qaGate` is on, a PR requiring QA without its approval label
is a **hard merge block**; so are the failed-QA and do-not-merge/blocked states.

## 10. Claim state

The `in-progress` label, assignees other than `$CURRENT_USER`, and any 🤖 claim
comment newer than the stale window from another actor. This feeds the claim
decision in the skill body, not the classification. `ci-monitoring` is **not** a
claim signal: it says an earlier run finished and reported its work and owes only
a CI-result comment, so record it as context and claim the PR normally.

---

## Output: the PR State Report

Fill this verbatim — it is what the classification step consumes and what the
report step publishes.

```markdown
### PR State Report — #{number} {title}

- Author / fork: {author} · {same-repo | own fork | other's fork} · {draft|ready} · pushable: {yes|no} · drivable: {yes|no}
- Plan: {plan path | run folder | none} — {done}/{total} steps done, next: {step}
- Diff scope: {spec-only|docs|UI|migration|contract|code} ({n} files)
- Review: {NONE|REVIEW_REQUIRED|CHANGES_REQUESTED|APPROVED} · {n} unresolved conversations
- CI: {green|red: names|pending|none} ({n} required checks)
- Mergeability: {mergeable|conflicting|behind|blocked|unknown}
- Labels: pipeline={x} category={…} meta={…} priority={x} risk={x} · missing: {…}
- QA: {n/a | required, no evidence | required + approved | skipped | failed}
- Blockers: {hard blockers, or none}
```
