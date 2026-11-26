# HANDOFF.md and NOTIFY.md templates (step 5)

The two tracking-file templates `om-auto-create-pr-loop` seeds in step 5 (a
Spec-implementation run), alongside `PLAN.md`. Save all three under `$RUN_DIR`;
create the directory if it does not exist.

`HANDOFF.md` (rewritten at every checkpoint and at run end):

```markdown
# Handoff — <date-slug>

**Last updated:** <UTC ISO-8601 timestamp>
**Branch:** <branch>
**PR:** <url or "not yet opened">
**Current phase/step:** <e.g. Phase 1 Step 1.2>
**Last commit:** <sha> — <short subject>

## What just happened
- <one or two bullets>

## Next concrete action
- <one bullet: the exact next Step to start on>

## Blockers / open questions
- <or "none">

## Environment caveats
- Dev runtime runnable: <yes|no|unknown>
- Browser / UI checks: <enabled|skipped because ...>
- Database/migration state: <clean|dirty — describe>

## Worktree
- Path: <worktree path>
- Created this run: <yes|no>
```

`NOTIFY.md` (append-only):

```markdown
# Notify — <date-slug>

> Append-only log. Every entry is UTC-timestamped. Never rewrite prior entries.

## <UTC ISO-8601 timestamp> — run started
- Brief: <one-line task summary>
- External skill URLs: <list or "none">
```
