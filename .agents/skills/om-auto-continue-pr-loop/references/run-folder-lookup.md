# Locate the run folder (step 3)

Prefer the explicit `Tracking plan:` line in the PR body (written by `om-auto-create-pr-loop`): from the `body` already fetched by the step 1 **get-pr** (no second fetch), take the first line matching `^Tracking (plan|run folder):` (e.g. pipe the body through `grep -E '^Tracking (plan|run folder):' | head -n1`).

Expected value (current format): `Tracking plan: ${RUNS_DIR}/<date>-<slug>/PLAN.md`.

Fallbacks, in order:

1. `Tracking run folder: ${RUNS_DIR}/<date>-<slug>/` — derive `PLAN_PATH` as `${folder}/PLAN.md`.
2. Legacy flat-file format: `Tracking plan: ${RUNS_DIR}/<date>-<slug>.md` — still honored for PRs opened before the folder migration. In this case there is no run folder yet; create one at `${RUNS_DIR}/<date>-<slug>/`, move the flat plan into it as `PLAN.md`, and initialize `HANDOFF.md` and `NOTIFY.md` as part of this resume's first commit.
3. Legacy `Tracking spec:` line (older runs) — treat the same way as the legacy flat-file format.
4. Diff the PR against `origin/$BASE_BRANCH` and look for a new path under `${RUNS_DIR}/` authored by this branch. If exactly one new plan exists (folder or flat file), use it.
5. Legacy fallback: if nothing under `${RUNS_DIR}/` is found, look for a new file under the repo's specs directory (`paths.specs`, default `.ai/specs`) for PRs created before the runs-folder migration. Migrate it into a new run folder as above.
6. If multiple candidates were found, stop and ask the user which one to resume.
7. If no tracking plan can be resolved, the PR was not created by a loop run — do **not** stop with an error and do **not** invent a plan path. Hand off to `om-auto-continue-pr {prNumber}`, whose adoption path reconstructs the missing execution plan from the PR's own context (description, comments, unresolved review feedback, linked issues, matching specs, and the code already landed), commits it, and links it from the PR body. If that reconstruction turns out longer than `engine.loopStepThreshold` Steps it hands the run straight back here, and this lookup then resolves the flat plan through fallback 2 and migrates it into a run folder. The lock you took in step 1 is **transferred, not released** — keep the `in-progress` label and assignee in place and post the chained hand-off comment per `references/claim-pr.md`, so the PR is never observably unclaimed mid-chain — and name the hand-off in your report. Only when `om-auto-continue-pr` is not installed does the run stop — then report the missing plan and print the install command for that skill.

Record the resolved paths:

```bash
RUN_DIR="${RUNS_DIR}/<date>-<slug>"
PLAN_PATH="${RUN_DIR}/PLAN.md"
HANDOFF_PATH="${RUN_DIR}/HANDOFF.md"
NOTIFY_PATH="${RUN_DIR}/NOTIFY.md"
# Verification is checkpoint-based: ${RUN_DIR}/checkpoint-<N>-checks.md every ~5 Steps.
# Optional artifacts (test logs, screenshots) live at ${RUN_DIR}/checkpoint-<N>-artifacts/.
# Final gate log lives at ${RUN_DIR}/final-gate-checks.md at spec completion.
```
