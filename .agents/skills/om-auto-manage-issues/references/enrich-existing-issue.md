# Enrich one existing issue

The per-issue procedure `om-auto-manage-issues` runs in step 2, for the single id
or for each issue in the batch. Every mutation goes through the tracker
descriptor's named operations and label guards; the whole procedure is idempotent
and claim-aware.

## 1. Fetch and decide whether to touch it

**get-issue** for the issue, requesting `number,title,body,state,author,url,labels,assignees,comments`.
Skip the issue (record the reason for the report, mutate nothing) when:

- A **different actor holds an active claim**: it carries `in-progress` with an
  assignee who is not `$CURRENT_USER` (resolve via **current-user**), or a fresh
  `🤖` claim comment (< ~30 min) from another actor. This is the three-signal
  check of `references/claim-pr.md`, used skip-only. Never collide with active
  work.
- It carries a repo-defined **human-hold** label (e.g. `do-not-close`, or any
  hold label the repo's `SDLC.md` / labels config marks as agent-off-limits).

This skill does **not** take its own `in-progress` lock — it is a fast, additive
housekeeping pass.

## 2. Apply missing SDLC labels

Read the issue's current labels. Infer the classification per `SDLC.md` (its
"When no priority label is set" and "When no risk label is set" lists, and the
category group) from the title, body, and any screenshot analysis from step 3:

- **Category** — add exactly one of `bug`, `feature`, `refactor`, `security`,
  `dependencies`, `documentation` when none is present.
- **Priority** — add exactly one `priority-*` when none is present.
- **Risk** — add exactly one `risk-*` when none is present.

Add each through the `apply_issue_label` guard (a missing label degrades to a
logged skip; `LABELS_ENABLED=false` skips all). **Only add what is missing** —
never remove or swap a label a human already set (a present label is treated as
the human's decision). After adding, post one short rationale comment covering the
labels applied (per `SDLC.md`'s "leaves a short comment explaining why"). If the
issue already carries a full category+priority+risk set, add nothing and note it.

## 3. Enrich a laconic issue (skip when `--relabel-only`)

Apply the laconic test and, when it trips, run the screenshot + wording procedure
in `references/screenshot-analysis.md`. In short:

- Analyze any attached screenshot(s) together with the terse text to reconstruct
  what is actually being reported/asked.
- Rewrite the issue body via the **update-issue** operation with a **clarified
  description**, preserving the reporter's original text verbatim in a collapsed
  section (non-destructive).
- Post the agent's **understanding** as a single comment — but **only if** no
  equivalent understanding comment from this skill already exists (scan via
  **list-issue-comments** for the skill's understanding marker). This is what
  makes re-runs a no-op.

## 4. Spec check (feature-category issues)

For an issue whose category is (or becomes in step 2) `feature`, check whether a
covering spec exists — the same approach as `om-prepare-issue` step 2, read-only:

- Read the repo's specs directory (`$SPECS_DIR`, plus the design-doc areas the repo
  uses) and judge by the spec's TLDR/overview whether its scope covers this issue's
  ask — not by filename alone.
- **search-prs** for an open PR that already adds a covering spec (a design/spec doc
  under `$SPECS_DIR` or the repo's design-doc areas) — a spec in flight counts as
  covered; note that PR.

Record `SPEC_STATUS` = `covered` (spec path or spec-PR link) | `missing`. A
`bug`/non-feature issue is `n/a`. This check mutates nothing on its own.

## 5. Handle a missing spec

Two branches on `SPEC_STATUS = missing`, chosen by the `--write-missing-specs` flag.

**Default (flag OFF) — post the spec-required comment.** Post a comment on the
issue, addressed to the issue author, saying the spec must be filled in before
implementation starts. Use
the standard idempotent marker and update in place on re-runs (never duplicate);
skip entirely when a spec-PR link from this skill is already on the issue (the gap
is being closed) or when the marker comment already reflects the current state:

```markdown
🤖 `om-auto-manage-issues` — spec required

@{author} 📝 this feature issue has no covering specification (checked `$SPECS_DIR`
and open spec PRs). Please fill up the spec before implementation starts:

- write it following the repo's spec conventions (the `om-spec-writing` skill), and
  link it here, **or**
- have it authored autonomously: run `om-auto-write-spec {issueId}` (or re-run
  triage with `--write-missing-specs`).

⛔ Implementation skills will treat this issue as not ready until a spec is linked.
```

The `missing` status plus the comment outcome (posted / updated / skipped) is
carried into the report. Under `--dry-run`, record the intent and post nothing.

**With `--write-missing-specs`** the comment above is not posted — the gap is closed
instead of announced: delegate to
**`om-auto-write-spec {issueId}`** verbatim — it claims the issue, writes the spec
via `om-spec-writing --autonomous`, opens the spec PR (`Refs #{issueId}`), and emits
the `Spec:` and `PR:` reference lines. Then link the result back on the issue via
**comment-issue** (spec path + spec-PR link), idempotently (skip when a spec-PR link
from this skill already exists). Under `--dry-run`, record the intent and mutate
nothing. This is the one place this housekeeping pass produces a PR — a design-only
spec PR, never implementation.

## Idempotency summary

Running this procedure twice on the same issue must change nothing the second
time: labels already present are left alone, the understanding comment /
clarified-body markers are detected and not duplicated, a spec-required comment is
updated in place (and removed from consideration once a spec is linked), and a
spec-PR link this skill already posted is not re-posted. Design every mutation as "add if missing",
never "post again".
