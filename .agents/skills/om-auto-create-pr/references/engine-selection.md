# Engine routing — when a run hands off to the loop engine

How `om-auto-create-pr` decides whether to run plain or hand the run to `om-auto-create-pr-loop`. Orchestrating skills (`om-auto-implement-spec`, `om-auto-fix-issue`'s feature route) delegate to this rule by invoking `om-auto-create-pr`, forwarding `--loop` verbatim only when the user passed it — they never add the flag on their own.

## The choice: plain unless the loop is earned

Plain execution is the default — always. The loop engine's run-folder ceremony (PLAN/HANDOFF/NOTIFY, per-step commits, checkpoints) costs several times the tokens of a plain run, so it must be earned.

Hand off to `om-auto-create-pr-loop` only when at least one holds:

1. **`--loop` was passed** — by the user directly, or forwarded verbatim by a routing skill.
2. **The drafted plan exceeds `LOOP_STEP_THRESHOLD` Steps** (`engine.loopStepThreshold`, default 20). Count the Steps (not Phases) drafted in step 4 — the items that would become the Progress checklist rows. More than the threshold → loop.

Nothing else selects the loop: not UI work (plain runs post screenshots via `om-auto-qa-pr` / **attach-image-evidence**), not "might not finish in one pass" (`om-auto-continue-pr` resumes a plain run fine), not subjective "feels large".

## Handoff mechanics

Two trigger points, one decision, nothing written before it:

- **`--loop` passed** → hand off immediately after the step-1 slot check, before any plan is drafted.
- **Otherwise** → at the end of step 4, after drafting the execution plan but **before saving it**: count the Steps; over the threshold → hand off; else save and continue plain.

The handoff invokes `om-auto-create-pr-loop` verbatim with the original `{brief}`, forwarding `--spec`, `--slug ${SLUG}`, every `--skill-url`, and `--force` when given. The drafted Progress-format plan is discarded — it was never written to disk or committed — and the loop skill derives its own run-folder `PLAN.md` from the brief/spec. There is no double-claim: step 1 is read-only, and the forwarded slug keeps the slot identity identical, so the loop's own claim step finds a free slot. After the delegated run completes, relay its report and chaining reference lines unchanged, prefixed with the `Engine:` line.

## The Engine report line

State the decision in one line in the run report and PR summary comment: `Engine: <name> (steps: <N>, --loop: <yes|no>)`. `<N>` is the count that informed the decision — the drafted plan's Steps, or `n/a` when `--loop` forced the handoff before a plan was drafted. Never reworded; callers relay it verbatim.

## Consequence for the plan format

The engine is decided **before** any plan is committed, because the two engines consume different plan formats: plain → the standard Progress-tracked execution plan (per `om-auto-create-pr` step 4); loop → the run-folder in `om-auto-create-pr-loop`'s format (`PLAN.md` Tasks table + `HANDOFF.md`/`NOTIFY.md`). A run's format is therefore fixed at creation — resume plain runs with `om-auto-continue-pr`, run-folder runs (a `Tracking run folder:` line in the PR body, or a run-folder tracking path) with `om-auto-continue-pr-loop`. Never re-apply the threshold to a run that already has artifacts.
