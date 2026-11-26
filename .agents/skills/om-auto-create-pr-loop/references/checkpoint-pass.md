# Checkpoint pass (every 5 Steps)

The checkpoint half of step 8. A checkpoint fires when any of these is true:
- 5 Steps have landed since the last checkpoint (or since the start of the run).
- The next Step would close a Phase and the Phase has ≥3 Steps.
- The run is about to hit the final-gate stage (step 9) — that final gate subsumes a checkpoint.
- A blocker stops the run mid-Phase.

At a checkpoint, run the following and record them in a single `${RUN_DIR}/checkpoint-<N>-checks.md`:

1. **Targeted validation for every area touched since the last checkpoint** — the relevant subset of `validation.commands` (typecheck and tests scoped to the affected packages when the toolchain supports scoping; otherwise unscoped), plus any configured codegen, localization-sync, or build commands whose inputs changed in the window.
2. **Step review (`engine.stepReview: checkpoint` only)** — review the diff landed since the previous checkpoint per `references/step-review.md`; blocker/major findings are fixed as `X.Y-review-fix` Steps before the checkpoint concludes.
3. **UI verification (conditional)** — if any Step in the window touched UI (pages, components, widgets, navigation):
   - Run the repo's integration suite via the `om-integration-tests` skill (running-only mode), scoped to the touched areas — prefer folder- or tag-scoped selection over the full suite at this stage.
   - If no existing test covers the touched area, fall back to browser automation tooling when available to drive a minimal smoke path against the running dev server.
   - Create `${RUN_DIR}/checkpoint-<N>-artifacts/` and save session transcripts (`browser-session.log`) and at least one screenshot per touched area (`screenshot-<short-desc>.png`). Reference filenames from `checkpoint-<N>-checks.md`.
   - **Post the checkpoint's verification outcome + screenshots to the PR.** The PR exists from step 7, so post immediately (never defer to the run folder alone): an idempotent verification comment — marker `` 🤖 `om-auto-create-pr-loop` — checkpoint <N> verification ``, summarizing the checks run with pass/fail/skip — and, when a Step in the window touched UI, this checkpoint's screenshots via **attach-image-evidence** — marker `` 🤖 `om-auto-create-pr-loop` — checkpoint <N> evidence ``, one line naming the Step range + touched areas, one line per image, slug `checkpoint-<N>`. Both are idempotent by marker: a re-run updates in place. Inline rendering is the tracker descriptor's job; when it cannot, the comment carries links — surface that, don't fail.
   - **UI checks MUST NOT block development.** If the repo has no integration suite, the dev env cannot be started, browser tooling cannot connect, or the scenario requires missing fixtures, skip the UI portion and record a single UTC-timestamped note in `checkpoint-<N>-checks.md` and `NOTIFY.md` explaining why. The checkpoint otherwise proceeds.
4. **Write `checkpoint-<N>-checks.md`** listing: checkpoint index, the Steps it covers (id range + SHA range), touched areas, every check run with pass/fail/skip + reason, and links to any artifacts.
5. **Rewrite `HANDOFF.md`** from scratch with the new state (next concrete action = the first `todo` Step).
6. **Append one NOTIFY entry** for the checkpoint: UTC timestamp, checkpoint index, Step range covered, one-line summary, any decisions/problems.
7. **Commit** the checkpoint files (`checkpoint-<N>-checks.md`, `checkpoint-<N>-artifacts/` if any, `HANDOFF.md`, `NOTIFY.md`) as a single commit: `docs(runs): checkpoint N — steps X.Y..X.Z verified`. Push.

If the checkpoint fails (typecheck/test/build/integration test regresses), halt dispatch, rewrite `HANDOFF.md` naming the failure, append a NOTIFY blocker entry, fix forward with new Steps appended to the Tasks table, and re-run the checkpoint before continuing.

## Subagent parallelism (optional, capped at 2)

- At your discretion, you MAY run up to **two** subagents concurrently — for example, one implementing the next Step while a second prepares targeted test evidence for the just-landed commit. Never exceed two.
- **Conflict avoidance is the top priority.** Two agents MUST NOT edit the same files in the same window. If conflicts are likely, serialize instead.
- Prefer serial execution whenever the gain is marginal. Parallelism is a tool, not a default.
- Record any subagent delegation in `NOTIFY.md` with timestamps so the reviewer can tell who did what.
