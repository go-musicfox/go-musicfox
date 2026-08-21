# Multi-Step runs: executor-dispatch pattern (step 8)

Applies only to **Spec-implementation runs** whose Tasks table marks Steps `dispatch`/`group` (or that hit the legacy fallback below) — Simple runs have at most one code commit and never dispatch.

The main session acts as a **dispatcher** and spawns one **executor subagent** per dispatched Step (foreground `Agent` tool call, `subagent_type: "general-purpose"`). The executor implements exactly that Step end-to-end (code commit + Tasks-table flip + push). The main session waits for the executor to return, verifies the commits landed and pushed, then dispatches the next Step.

Placement is decided at planning time: follow the `Exec` column of PLAN.md's Tasks table mechanically — `inline` → drive the Step in the main session with the default per-Step loop; `dispatch[:tier]` → spawn one executor for the Step; `group:<id>[:tier]` → spawn one executor for the whole contiguous group (see Group dispatch below). Do not re-litigate the plan's placement at run time. Plans without an `Exec` column (written before it existed) fall back to the **legacy heuristic**: dispatch each Step when ≥3 Steps are expected to land in this invocation, else run them inline; an empty, `—`, or unrecognized cell falls back the same way for that row, at the default tier.

Hard constraints:

- Subagents do NOT have access to the `Agent` tool. A coordinator subagent **cannot** spawn executors. Dispatch MUST live in the main session.
- Dispatch is **sequential** (one executor at a time). This is not parallelism — the cap-at-2 rule above still applies to the rare case where you want an implementer and a reviewer running side-by-side; an executor-dispatch run is a sequence of one-at-a-time executors.
- The main session claims the PR's three-signal `in-progress` lock **once** when the draft PR opens and releases it at run end. Executors MUST NOT claim or release the lock, and MUST NOT post PR comments — checkpoint evidence, verification comments, and the final summary are the main session's job.

## Model tier (best-effort)

When the harness's `Agent` tool supports model selection for subagents, map the Step's tier — `cheap` / `standard` / `capable` — onto the closest model class the harness offers and spawn the executor on it. When it does not, dispatch normally and ignore the tier. Tiers are abstract hints, never model names and never a blocker. Tier resolution: the `Exec` cell's suffix → the `engine.executorTier` config key → `standard`:

```bash
EXECUTOR_TIER=$(jq -r '.engine.executorTier // "standard"' "$CONFIG")
```

The rationale: a Step whose plan text is a complete spec is transcription work — the cheap tier removes the cost reason to keep small independent Steps inline. When a tier was actually applied, name it in the NOTIFY delegation entry.

Executor prompt template — the main session writes this into each spawned `Agent` call:

```markdown
You are an executor for om-auto-create-pr-loop run {SLUG}. Implement exactly one Step.

Working directory: {absolute worktree path}
Branch: {branch} (already checked out from origin/{BASE_BRANCH}; origin tracking set up)
Run folder: {absolute run folder path}

Step to implement:
- Step id: {X.Y}
- Title: {step title from Tasks table}
- Full description: {paste the Step's bullets from PLAN.md Implementation Plan}

Spec anchors:
- PLAN.md: {plan path}
- Source spec (if any): {spec path}
- External References adopted: {list from PLAN.md Overview}

Rules:
- One Step = exactly one code commit. Nothing more, nothing less. No docs-flip commit.
- Run a quick scratch sanity check (typecheck + new test, from the configured validation commands) to confirm the Step compiles. Do NOT record it anywhere — the checkpoint pass verifies.
- Do NOT write a `step-{X.Y}-checks.md`. Do NOT create a `step-{X.Y}-artifacts/` folder. Verification is checkpoint-based.
- Flip the `Status` cell of row `{X.Y}` in PLAN.md's Tasks table from `todo` to `done` and fill the `Commit` column with the short SHA as part of the same commit (amend if needed to capture the real SHA before push).
- Do NOT rewrite `HANDOFF.md` at the per-Step level. Do NOT append to `NOTIFY.md` unless you hit a blocker, make a scope decision worth logging, or are delegating to another subagent.
- Push after the commit so the remote always has the latest state.
- Do NOT claim or release the PR's `in-progress` lock, and do NOT post the final summary PR comment. The main session owns both.
- Do NOT rewrite or reorder prior history. Do NOT split into multiple code commits. If this Step truly needs splitting, stop and return early with a report asking the main session to split the Step in PLAN.md first.

Return format (concise report, < 300 words):
- Step id
- Code commit SHA
- Files touched
- Brief note on what changed (one line)
- Push confirmation (`origin/{branch}` now at {sha})
- Blockers or decisions worth escalating
```

## Group dispatch

A `group:<id>` set is dispatched as **one executor for the whole group**: when the dispatcher reaches the group's first `todo` member, it spawns a single executor that receives all remaining `todo` members in table order — replace the template's "Step to implement" block with an ordered "Steps to implement" list. Every rule in the template applies **per member Step**: one code commit + Tasks-row flip + push each, in order. Groups are contiguous rows sharing the identical cell value (same tier — one executor, one tier). On resume, dispatch only the group's remaining `todo` members. If any member truly needs splitting, the executor stops and returns early exactly as the single-Step rule says.

Verification the main session MUST run after each executor returns — before dispatching the next Step:

- `git status` is clean in the worktree.
- Exactly **one** new commit **per dispatched Step** exists on HEAD since the dispatch (a group executor returns one commit per member Step).
- Local HEAD == `origin/{branch}` (push actually landed; fetch if in doubt).
- The PLAN.md Tasks-table row for each dispatched Step is flipped to `done` with the correct short SHA in the `Commit` column.

When `engine.stepReview` is `per-step`, the main session reviews each dispatched Step's commit range per `references/step-review.md` after this verification passes, before dispatching the next Step (a group executor's Steps are reviewed as one range).

Every 5 successfully landed Steps (a group counts each member Step; or when a Phase with ≥3 Steps closes), the main session MUST run a full **checkpoint pass** per step 8 (`references/checkpoint-pass.md`) before dispatching the next Step.

## Escalation before the stop

A problematic executor result — a returned blocker, failing tests, an error, or a failed post-executor verification — gets **one rescue attempt** before the safety stops fire: restore a clean worktree state (discard the failed executor's uncommitted leftovers; committed prior Steps are never touched), then dispatch a **fresh executor for the same Step one tier above** the failed one (`cheap` → `standard`, `standard` → `capable`), with the failed executor's report and the concrete failure appended to its prompt. The rescue result goes through the same post-executor verification. Escalation is bounded: one rescue per Step, never above `capable` — a Step that already ran at `capable`, or whose rescue also fails, halts per the safety stops below. Append a NOTIFY delegation entry for every rescue, naming both tiers and the failure it answers.

Safety stops — the main session MUST halt dispatch (leave `Status: in-progress` in the PR body if the PR is open, rewrite `HANDOFF.md`, append a NOTIFY entry naming the blocker, release the lock, and report back) when, after the one rescue attempt above where applicable, any of the following is true:

- An executor returns a blocker, failing tests, or an error — and its rescue also failed or was unavailable.
- `git status` is not clean after an executor returns.
- The Tasks-table row was not flipped to `done` with the correct SHA.
- Local HEAD ≠ `origin/{branch}` (push did not land).
- Two consecutive Steps needed a rescue, even when both rescues landed.
- **Safety checkpoint:** after ~20 consecutive successful Steps, stop and let the user review before plowing on.

Sibling auto-skills (`om-auto-continue-pr-loop`, `om-auto-update-changelog`) inherit this pattern when driving multiple Steps in a single invocation.
