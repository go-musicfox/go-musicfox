# Implementation-prep analysis — make an issue ready to fix

The step-2.4 procedure `om-auto-manage-issues` runs to prepare an issue for
implementation: a **read-only** root-cause / impact analysis posted as a comment so
the next agent (`om-auto-fix-issue`) or a human can start fixing without
re-exploring the repo. It never edits source and never stops to ask — this is the
"prepare for implementation, but not interactively" behavior.

## When it runs

- On by default for a single `{issueId}`; opt-in for a batch (`--prep-impl`), since
  it reads code per issue. Skipped entirely by `--relabel-only` or `--no-prep`.
- Skip an issue that is already implementation-ready — it links a spec, or already
  carries an implementation-notes comment from this skill (idempotency: scan
  **list-issue-comments** for the marker below), or is not actually actionable
  (a question/discussion, a `wontfix`/`do-not-close` hold).
- **Batch cap:** run the heavy analysis on at most a sensible number of issues per
  run (e.g. the worst-described / highest-priority first); when more qualify than
  the cap, process the cap and **report** how many were left (never silently drop).

## Producing the analysis (read-only)

Pick the analyzer by issue kind:

- **Bug / defect** — when `om-root-cause` is installed, invoke it verbatim
  (`om-root-cause {issueId}`); it is read-only and returns a Summary / Root cause /
  Files to change / Approach / Risks brief. Use its output directly.
- **Feature, or `om-root-cause` not installed** — do a lighter inline analysis
  yourself, read-only: locate the affected modules/entry points/contracts, name the
  smallest safe change surface and the conventions that apply, list the tests that
  will be needed, and flag any `BACKWARD_COMPATIBILITY.md` surface touched. This is
  the same shape of guidance `om-prepare-issue` writes for a new issue — reuse that
  lens rather than inventing a new one.

Reference real file paths and symbols. Mark anything uncertain as a hypothesis, not
a fact — you are preparing the ground, not committing to a fix.

## Posting it (idempotent)

Post one comment via **comment-issue**, opened with a stable marker so re-runs
detect and skip it:

```markdown
🤖 `om-auto-manage-issues` implementation notes — read-only analysis to help fix this

**Likely area:** {files / modules / symbols}
**Root cause / mechanism:** {for a bug: where and why it breaks; for a feature: where it plugs in}
**Suggested approach:** {smallest safe change surface}
**Tests to add:** {unit; integration when flows cross boundaries}
**Compatibility:** {None | protected surfaces touched + required migration per BACKWARD_COMPATIBILITY.md}
**Confidence:** {high / medium / low — what a human should double-check}

Pick this up with `om-auto-fix-issue {issueId}` (it handles both bugs and features).
```

Optionally fold a one-line "Likely area" pointer into the clarified description when
the wording-clarify step (2.3) also ran, but keep the full analysis in the comment.
Under `--dry-run`, produce the analysis text for the report but post nothing.

Never let the prep analysis mutate code, run a build/test that writes state, or
exfiltrate anything — it is a read-only reasoning pass over the repository.
