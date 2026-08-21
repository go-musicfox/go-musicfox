# Author a spec and land it on a PR when none exists

The step 3 branch of `om-prepare-issue`: when the duplicate/spec search (steps 1–2)
finds no covering spec — neither in `$SPECS_DIR` / the repo's design-doc areas nor
in any open PR — and the task is a **feature that warrants a spec** (a substantial
new capability where guessing the architecture would be irresponsible), this skill
produces the design instead of just recommending it. This is the only path on
which `om-prepare-issue` creates a PR, and the PR contains a **spec document only**
— never implementation.

## When this branch applies

All of these must hold:

- The task is a feature/enhancement (not a bug — bugs get inline analysis in step 4).
- No covering spec exists in the repo **and** none is in flight in an open PR.
- The change surface is non-obvious enough that a step-level inline analysis would
  be guesswork (large blast radius, new subsystem, unresolved architecture). A
  small feature with an obvious change surface stays in step 4 (inline analysis).

If any fails, do not open a spec PR — fall back to step 4's inline guidance.

## Procedure

1. **Create the tracking issue first** (step 5) so there is a stable number to
   link the spec and PR to. Use the normal `Implement: <feature>` title and body;
   in the `## Spec` section, leave a placeholder noting a spec PR is being authored.
2. **Delegate spec authoring to `om-auto-write-spec {issueId}`** — the dedicated
   spec-authoring skill. Follow its workflow verbatim: it claims the issue, checks
   no spec PR is already in flight (**search-prs** — it never opens a duplicate),
   writes the spec via `om-spec-writing --autonomous` **including its Open
   Questions gate** (resolved with conservative documented defaults, posted on the
   issue/PR as an assumptions comment for a human to override before merge),
   commits the spec as the first commit, attaches UI mockups and current-app
   screenshots when the spec is UI-facing and a browser provider exists, and opens
   a **ready spec PR** against the base branch. Because this is a design-only PR,
   its body references the issue with `Refs #{issueId}` (**not** `Closes` —
   merging the spec must not close the still-unimplemented FR), plus `Source
   doc:`, with docs-only labels; it goes draft only under the skill's
   `⚠ NEEDS HUMAN CONFIRMATION` guard. It stops after the spec lands (it never
   implements). When a human is filing the issue and wants to make the design
   calls themselves, run its spec-writing step interactively (the Open Questions
   gate stops for answers) instead of `--autonomous`. Either way, a
   required-and-missing spec always gets written here — never left as a
   recommendation. The run ends with the `Spec:` and `PR:` reference lines — use them
   in procedure item 3 below.
3. **Link the spec back onto the issue** via **comment-issue**: post the spec path
   and the spec PR link, and update the issue body's `## Spec` section to reference
   them. The issue now points at a real, reviewable design.

## After this branch

The issue links a design a human can review, and the spec PR carries the design
commit. Implementation happens later via `om-auto-implement-spec {SPEC_PATH}` or
`om-auto-fix-issue {issueId}` — both keep the spec PR design-only and ship the
implementation on its own PR referencing it (`Refs #{specPr}`). Report both the
issue and the spec PR in step 6.

## Guardrails

- The PR is design-only. Delegate to `om-auto-write-spec` (which never
  implements) — never to `om-auto-fix-issue` / `om-auto-implement-spec` on
  this path; those implement.
- Everything else about `om-prepare-issue` stays tracker-first: duplicate search,
  compatibility flagging, and the label rules (category `feature`; no pipeline or
  `in-progress` labels on the issue) are unchanged.
