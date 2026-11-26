# FR triage gate — is this an unbuilt feature?

The read-only gate `om-auto-fix-issue`'s **feature route** runs (step F1) before
touching a worktree. It asks: is this really a feature request, and is the feature
not already implemented? Operate
**read-only** — file reads, code search, and read-only tracker operations
(**get-issue**, **search-prs**, **search-issues**, **repo-info**,
**current-user**) only. No edits, commits, claims, or branch creation.

## Decision procedure

Run in order; the first stop wins. Bias toward stopping — when you cannot defend
"real, unbuilt feature" with at least one piece of evidence, stop.

### 1. Is the issue closed or already in progress by someone else?

If `state` is `closed`, stop with `NO_ACTION_NEEDED`. If the three-signal lock
(step 1 of the body) shows another actor owns it, stop with `NO_ACTION_NEEDED` and
name the owner. (Stale-lock recovery per the body: a lock older than 60 minutes
with no activity is not a stop on its own.)

### 2. Is it a feature request, or actually a bug?

Re-check the step-2 classification (label-first, then content — the same
feature-vs-bug signals). If the issue is clearly a **bug**, stop this feature route
with `NO_ACTION_NEEDED` and state that it belongs on `om-auto-fix-issue`'s bug
chain. When an issue mixes a defect and a new capability, prefer stopping and
recommend the user split it (the bug goes to the bug chain, the FR stays on the
feature route) rather than guessing.

### 3. Is the feature already implemented, or already in flight?

Before writing any spec, prove the feature does not already exist:

- **Code search** the repository for the capability the issue asks for — the
  endpoint/route, command, flag, screen, config key, or public function it names.
  A matching, working implementation (with tests) means the feature exists → stop
  with `NO_ACTION_NEEDED`, citing the file/symbol.
- **Specs directory** (`$SPECS_DIR`, plus the repo's design-doc areas): read the
  TLDR/overview of candidate specs. A spec that already fully covers the ask is
  reused later (step F3), **not** a stop — note its path and continue. A spec plus
  an **open PR or merged PR/commit** already delivering it is a stop with
  `NO_ACTION_NEEDED` (search via **search-prs** for `#{issueId}` and for the
  feature name; scan `origin/$BASE_BRANCH` history).
- Scan recent issue comments for `implemented in`, `shipped in`, `duplicate of`,
  `superseded by` and follow the links.

## Output contract

**Stop the run** (no action needed):

```
NO_ACTION_NEEDED
<one paragraph: why — cite file paths/symbols, PR numbers, commit hashes, or the
bug-vs-feature classification and the om-auto-fix-issue hand-off, as evidence>
```

The literal token `NO_ACTION_NEEDED` on its own line triggers the clean stop.

**Proceed:**

```
<one short paragraph confirming this is a real, unbuilt feature request; note
whether an existing spec in $SPECS_DIR covers it (path) or a fresh spec is needed,
and the area of the codebase the implementation will touch>
```

Keep it tight (≤200 words). The spec-writing step designs the feature; do not
design it here.
