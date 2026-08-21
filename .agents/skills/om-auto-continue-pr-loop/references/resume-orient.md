# Orient via HANDOFF.md, then parse PLAN.md's Tasks table (step 5)

**Read `HANDOFF.md` first.** It is the authoritative short-form snapshot of what the previous agent (or this agent's previous session) was doing. It tells you: the current phase/step; the last commit SHA and what it delivered; the next concrete action; open blockers, environment caveats, and worktree details.

Then open `PLAN.md` and find the `## Tasks` table at the top of the file. It is a markdown table with these columns: `Phase`, `Step`, `Title`, `Exec`, `Status`, `Commit`. Plans committed before the `Exec` column existed have five columns (no `Exec`) — parse them identically and **never rewrite an old table to add the column**. Example shape written by `om-auto-create-pr-loop`:

```markdown
## Tasks

> Authoritative status table. `Status` is one of `todo` or `done`. On landing a Step, flip `Status` to `done` and fill the `Commit` column with the short SHA. The first row whose `Status` is not `done` is the resume point for `om-auto-continue-pr-loop`. Step ids and `Exec` cells are immutable once the plan is committed.

| Phase | Step | Title | Exec | Status | Commit |
|-------|------|-------|------|--------|--------|
| 1 | 1.1 | {step title} | dispatch | done | abc1234 |
| 1 | 1.2 | {step title} | dispatch:cheap | done | def5678 |
| 2 | 2.1 | {step title} | group:A | todo | — |
| 2 | 2.2 | {step title} | group:A | todo | — |
```

Parse rules:

- The **first row whose `Status` column is not `done`** is the resume point. `Status` values are `todo` or `done` only.
- The Step id comes from the `Step` column (`X.Y` or `X.Y-review-fix`). That id drives the Step commit and any checkpoint bookkeeping that references it.
- `Title` is informational and must match the Step title in the Implementation Plan section; if it drifts, trust the Implementation Plan title and fix the table.
- `Exec` is consumed by the dispatcher (`references/executor-dispatch.md`) to decide per-Step placement and model tier; a missing column or an empty/unrecognized cell falls back to the legacy heuristic described there.
- If `HANDOFF.md` names a different resume point than the table implies, trust `HANDOFF.md` and reconcile the table (a previous session may have crashed mid-Step). Log the reconciliation in `NOTIFY.md`.
- If the `## Tasks` table is missing, fall back to a legacy `## Progress` checkbox section (PRs opened before the table migration used checkboxes — first `- [ ]` is the resume point). When you hit a legacy Progress section, migrate it to a Tasks table as part of the resume's first commit — written in the six-column format, `Exec` filled per the planning heuristics.
- If neither the table nor a legacy Progress section can be parsed, stop and ask the user — unless `--from <phase.step>` was passed, in which case use that as the resume point and log a note in `NOTIFY.md`.
- Cross-check the most recent `done` row's `Commit` SHA against `git log` on the PR head. If the recorded SHA is not reachable, warn the user and ask whether to continue (or accept `--force`).
- Skim the tail of `NOTIFY.md` (e.g. last 30 entries) for recent blockers or decisions so you do not repeat or contradict prior work.

Append a NOTIFY entry announcing the resume:

```
## <UTC ISO-8601 timestamp> — om-auto-continue-pr-loop resume
- Resumed by: @<current-user>
- Resume point: <phase.step> (source: HANDOFF.md / Tasks table / legacy Progress / --from)
- PR head SHA: <sha>
```
