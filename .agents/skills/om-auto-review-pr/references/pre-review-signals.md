# Pre-review signals (conflicts and CI)

Detailed procedure for step 4 of `om-auto-review-pr`. Run these checks before the worktree is created.

**Neither check ends the run.** They are signals gathered up front so the review that follows can account for them — not gates that replace it. One invocation must produce the most complete, most informative picture of the PR it can: a run that stops at the first red signal spends a whole cycle telling the author something the tracker already showed them, while the review that would have told them something new never happens. The author fixes the one thing they already knew about, pushes, and pays for another full cycle to discover the rest. Collect the signal, review anyway, report both in the same verdict.

Signals gathered here become **blocker findings** in step 9, so they drive `changes-requested` exactly as before — the verdict is unchanged; only its completeness improves.

## 4a. Check for merge conflicts

Run the tracker operation **get-pr** for `{prNumber}`, requesting `mergeable`, `mergeStateStatus`, and `baseRefName`.

If `mergeable` is `CONFLICTING` or `mergeStateStatus` is `DIRTY`, set `CONFLICTED=true`. What happens next depends on `AUTOFIX_ELIGIBLE` (step 2).

- On a **pure review pass** (`AUTOFIX_ELIGIBLE` false — another author's PR without `--autofix`), this run may not touch the branch, so the conflict cannot be resolved here. Carry it as a **blocker finding** — "the head conflicts with `{baseRefName}` and cannot merge until it is rebased or merged forward", naming the conflicting paths when the tracker reports them — and **continue with the review** of the head as it stands. The review body leads with that blocker and states the caveat plainly: the diff was reviewed as pushed, and resolving the conflict may change it, so whatever the resolution touches is worth a second look. A conflicted head is still a real commit with real code in it, and reviewing it now is worth strictly more to the author than a bare "please rebase".
- On an **autofix-eligible pass**, conflicts are **the first work item, not a deferral**. Go to step 5, create the worktree, and resolve the conflicts against the latest base *before any review work* — before the duplicate check, the diff scan, `om-code-review`, the tests, and certainly before CI — then resume the review at step 6 on the resolved branch. Here the diff under review is one this run controls, so resolving first makes every downstream signal measure the code that will actually merge. Conflicts that appear later (the base advances mid-run) are handled the same way, at the head of the step 11 loop. Fork heads resolve on the carry-forward branch instead (`references/fork-pr-flow.md`).

## 4b. Check CI status

Discover required checks first: run the tracker operation **get-required-checks** for the PR's base branch (`{baseRefName}`). If branch protection is not readable (the operation reports 404/no data), treat all reported PR checks as required.

Fetch the actual PR check results with the tracker operation **get-pr-checks** for `{prNumber}`, requesting each check's `name`, `state`, and `link`.

Treat these states as failing: `FAILURE`, `ERROR`, `CANCELLED`, `TIMED_OUT`. Ignore these as non-failing: `PENDING`, `SUCCESS`, `SKIPPED`, `NEUTRAL`.

**A failing required check never stops the review.** Record each one in `FAILING_CHECKS` (name, state, link) and carry it as a **blocker finding** into step 9 — red CI is real evidence and still forces `changes-requested` on its own — then continue to step 5 and run the full review regardless. A failing check is one finding among however many the review turns up, and the author deserves all of them in one pass rather than the cheapest one first. The review body lists the failing checks in their own subsection, each with its name and link, alongside the ordinary findings.

Use the review to explain the failure rather than restate it:

- The step-8 validation gate (`validation.commands`, run inside the worktree) usually reproduces the same failure locally. When it does, report the concrete cause once — the failing test, the type error, the `file:line` — and tie the red check to it instead of filing the same problem twice.
- When the local gate is **green** while a required check is red, say so explicitly: the failure is environmental, flaky, or specific to CI's configuration, and that is a materially different message to the author than "CI is red".
- Pull the failing job's log or summary through the tracker when the descriptor exposes an operation for it. Naming the failing test beats repeating the check name.

On an **autofix-eligible pass**, failing checks are stage 3 of the step 11 work order — fixed once neither conflicts nor review findings remain, because resolving those changes what CI does on the next run anyway.

**Pending is not failing, and pending never delays the review.** A required check still queued or running is not a reason to wait, and never a reason to hold back the verdict, the labels, or the review body. Review the code that is in front of you, submit, label, and report — then let `references/ci-followup.md` handle the CI outcome afterwards. Record which required checks were pending at review time, in `PENDING_CHECKS`, so step 10 can disclose them in the review body and the follow-up knows what it is waiting on. Waiting for green before reporting is the failure this skill must not reproduce: a monitoring process that dies mid-wait leaves a PR with no labels, no review, and no record that any work happened.
