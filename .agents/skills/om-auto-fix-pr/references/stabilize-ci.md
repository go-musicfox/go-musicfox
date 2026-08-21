# The stabilization loop and the CI procedure

Two related things `om-auto-fix-pr` runs: the **outer loop** that drives a PR to
approvable + green + QA-evidenced, and the **CI stabilization procedure** the loop's
CI stage (and the skill's `--ci-only` mode) uses to drive checks green. Each
sub-skill named below is invoked verbatim; this file sequences them and defines the
exit criteria and the CI-fixing detail.

## The outer stabilization loop (full PR mode)

Run in this order each cycle. The order is mandatory, not a convenience: a stage
depends on every earlier one having landed, and reaching CI on a branch that is
still conflicted or still carries review findings measures a diff that will never
merge.

1. **Conflicts, then review findings — `om-auto-review-pr {prNumber} --autofix`**
   (verbatim; one engine, invoked in the right order — this skill never
   re-implements conflict resolution or fixing). Inside that skill the autofix pass
   resolves **merge conflicts against the latest base first, always**, and only then
   works the code-review findings — blockers, majors, mediums, missing tests. For a
   fork head it runs the carry-forward supersede/credit flow (if it opens a
   replacement PR, switch `{prNumber}` to it for all later stages). Record its
   verdict and, crucially, the findings it did **not** fix (nits, low-severity,
   out-of-scope) — step 5 files those as follow-ups.

   **Do not proceed to stage 2 while the branch is conflicted or actionable findings
   remain.** CI stabilization is reached only once both are clear.
2. **Stabilize CI — the CI procedure below.** Pull check status through the
   tracker, classify each failure (real bug / test bug / flake / infra), fix the
   real ones with tests, push, re-check. Never go green by weakening a test or
   disabling a check. Honor `--max-iterations` for the inner fix→push→re-check loop.
3. **Verify UI — `om-auto-qa-pr {prNumber}`** — only when the diff touches a
   user-facing surface (inspect **get-pr-diff** / **get-pr-files** for UI paths)
   and `--no-ui` was not passed. It boots the app, drives the changed flow, and
   posts screenshots + a pass/fail report as a PR comment. Invoke it in its default
   **evidence-only** mode (no flags) so it sets **no** QA labels — this orchestrator
   never self-approves QA. When there is no runnable UI surface, skip and note
   "UI: n/a".
4. **Re-merge base** if it advanced this cycle (see `references/base-merge.md`).

## CI stabilization procedure

Given the PR (or, in `--ci-only` branch mode, a branch) this drives CI green:
read the failing checks, diagnose from the actual failure logs, apply minimal fixes
with regression coverage, push, wait for the re-run, repeat. The one thing it never
does is *fake* green — deleting tests, loosening assertions, skipping steps, or
disabling checks to pass is forbidden, and a repo-local override cannot relax that.

### Baseline: what is failing?

- PR mode: **get-pr-checks** for `{prNumber}` (name, state, link) and
  **get-required-checks** for the base branch; when required checks are unreadable,
  treat all reported checks as required.
- Branch mode (`--ci-only --branch <name>`): **list-runs** for the branch; take the
  newest run per workflow; **get-run** for the per-job breakdown.

Classify every check as passing, pending, or failing. Nothing failing and nothing
pending → report "already green". Checks pending → **watch-run** (or poll) until they
settle before diagnosing, under the wait budget below.

### The fix → push → re-check loop (up to `--max-iterations`, default 5)

Per iteration:

1. **Collect evidence.** For every failing check, fetch the failed-step logs:
   resolve the run via **list-runs** / **get-run**, then **get-run-failed-logs**.
   Extract the first real error per job (assertion, compile error, stack trace),
   not the last noise line.
2. **Classify each failure** into one primary bucket:
   - **Real code bug** — the change (or branch state) genuinely breaks behavior.
     Fix the code, add or extend a regression test.
   - **Test bug** — stale locator/fixture, wrong assumption, nondeterminism in the
     test itself. Fix the *test's correctness*; never weaken what it proves.
   - **Flake** — suspected when the failure is unrelated to the diff,
     timing-dependent, or historically intermittent. Before touching code,
     **rerun-failed** once. If it passes on rerun, record it as a flake — do not
     "fix" it blindly; note it for a follow-up issue.
   - **Infra / out of scope** — runner outages, missing secrets, base branch already
     broken (verify by checking whether the same check fails on `origin/$BASE_BRANCH`
     via **list-runs** on the base branch). These are blockers, not fixables —
     record and move on; if every remaining failure is out of scope, stop with
     `Status: blocked`.
3. **Reproduce locally** whenever the check maps to a local command — match the
   check name against `validation.commands` and the repo's scripts, and run the
   matching command in the worktree. Fix against the local reproduction; it is
   faster and proves the diagnosis.
4. **Fix minimally.** No refactors, no scope creep, no drive-by cleanups. Respect
   the project's conventions from its agent instructions. Changes to CI workflow
   files are allowed only when the workflow itself is the bug (e.g. a wrong path
   filter) — never to remove, skip, or soften a check; flag any workflow edit
   prominently in the summary.
5. **Validate locally**: run the targeted commands for what changed, and the full
   `validation.commands` gate when fixes span more than one area.
6. **Commit and push**: one commit per logical fix, conventional subject
   (`fix(ci): …`, `fix(test): …`, `fix(<area>): …`). Never rewrite published
   history, never `--no-verify`, never force-push.
7. **Wait for CI — bounded**: **watch-run** on the new runs (or poll
   **get-pr-checks**) until they settle, for at most `CI_MAX_WAIT_MINUTES`
   (`ci.maxWaitMinutes`, default 40). Green → done. Still failing → next iteration
   with the new evidence. A check that fails the same way twice after a targeted fix
   means the diagnosis is wrong — re-diagnose from scratch instead of stacking
   guesses. Budget exhausted with checks still running → take the bail-out path in
   `references/ci-followup.md`: stop waiting, run `validation.commands` locally as
   this run's evidence, post it with the still-pending check names and the explicit
   statement that no further follow-up will come from this agent, drop
   `ci-monitoring`, and close the run out. Bailing out is **not** permission to merge
   without CI — the hand-off to `om-approve-merge-pr` still requires genuinely green
   required checks.

### The wait budget, and reporting before it

Everything this skill owes the PR — normalized labels, the draft→ready promotion, the
run summary comment, the lock — lands **before** any CI wait, never after it (step 6,
and `references/ci-followup.md` for why). A monitoring process that dies mid-wait must
leave a fully reported PR behind, not a stranded draft. Once reporting is done, the
outer lock is swapped from `in-progress` to the `ci-monitoring` meta label so the PR
is honestly described — work complete, CI follow-up still owed — and no other agent
treats it as claimed.

### CI exit conditions

- **Green**: every required check passes (non-required failures are reported but do
  not block success — say so explicitly).
- **Blocked**: remaining failures are all infra/out-of-scope, or `--max-iterations`
  is exhausted. Report `Status: blocked` with the per-check analysis — never leave
  it at "CI is still red".
- **Unverified**: the wait budget ran out with checks still running. Report the local
  validation gate as this run's evidence, name the pending checks, and state that no
  further follow-up will come from this agent. Never report this as green, and never
  hand it on as merge-ready.

For each confirmed flake (failed, passed on rerun, unrelated to the diff), file a
tracked follow-up (via `om-followup-issue-from-pr` or **create-issue**): test name,
failure signature, run link, rerun evidence.

## Outer-loop exit criteria

Stop the loop and go to step 5 when **all** hold:

- `om-auto-review-pr` returns approvable — no conflicts and no actionable blockers.
- Every **required** check is green (**get-required-checks** / **get-pr-checks**), or
  the wait budget expired and the run reports `Unverified` rather than merge-ready.
- UI verification passed, or is not applicable / was skipped by `--no-ui`.
- No unmerged base advance is pending.

Stop early — without declaring merge-ready — when `--max-iterations` is reached or a
genuine blocker remains (an unresolvable conflict, a real failure the fix for which
is out of scope, a design objection from review). In that case leave the PR in the
honest pipeline state (`changes-requested` or `blocked`), do not file it as
merge-ready, and report the specific blocker so a human or a later run can take it.

## Convergence guardrails

- If a cycle makes **no** forward progress (same findings, same red checks) twice
  in a row, stop — looping again will not help. Report the stuck state.
- Watch for a fix that regresses another stage (a review fix that breaks CI, a CI
  fix that breaks the UI). The ordered re-run each cycle catches it; if two stages
  keep undoing each other, stop and report the tension rather than thrashing.
- CI fixes are minimal and evidence-based; a fix that does not change the failure
  signature on the next run means the diagnosis was wrong — re-diagnose, do not
  stack guesses. Failures that also occur on the base branch are out of scope —
  report them, never fix unrelated base breakage from this branch.
