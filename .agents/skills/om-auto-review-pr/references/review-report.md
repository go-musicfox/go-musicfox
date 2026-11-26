# Review and report — the om-code-review pass plus the autofix loop

Detailed procedure for step 8 (full review pass) and step 11 (autonomous autofix and fix-forward loop) of `om-auto-review-pr`. This skill is itself the automated second pass that other skills in this collection invoke; the review engine underneath is always the `om-code-review` skill, executed verbatim inside the isolated worktree.

## Review pass (step 8)

Use the `om-code-review` skill against the PR diff. Explicitly verify:

- No public contract was broken silently: exported APIs, HTTP routes and response shapes, event names, CLI flags, DB schema, config formats. When `BACKWARD_COMPATIBILITY.md` exists at the repo root, check every touched surface against it — violations of its protected surfaces are Blockers and the review must explicitly WARN the user; honor any other documented compatibility rules.
- No security-sensitive surface was weakened: authentication, authorization, data scoping, input validation, secrets handling.
- Test coverage and cross-cutting impact.

## om-auto-review-pr specifics — pass scope and gates (step 8)

- Scope changed files with the changed-file list from the tracker operation **get-pr-diff** for `{prNumber}`
- Gather context from the repository's agent-instruction and contributing docs covering the changed areas, plus the repo-local review checklist when the config's `reviewChecklist` points at one
- Run the full validation gate: every command in `validation.commands`, in order
- Apply the full review checklist and the breaking-change checklist above
- Merge findings from step 7 (diff-level auto-detections) into the final review report. Do not duplicate the same issue twice.

## om-auto-review-pr specifics — autonomous autofix flow (step 11)

The loop below runs only on **autofix-eligible** runs: `--autofix` was passed, or the PR author is `$CURRENT_USER` (the automation finishing its own work). On another author's PR without `--autofix`, skip it — the run ends with the review, labels, and author handoff, noting `autofix: skipped (not my PR — re-run with --autofix to fix it here)` in the completion comment and report. Never modify someone else's branch uninstructed.

**Work order — binding, in this sequence:** (1) **any merge conflict against the latest base, always first**, including one that arose since step 4a and never deferred to a later pass — a conflicted branch makes every downstream signal unreliable, because the diff under review is not the diff that will merge; (2) **then the findings** — the unit-test audit first, then fix batches with targeted validation, re-review, repeat; (3) **CI only once neither conflicts nor findings remain**, since resolving those changes what CI does on the next run anyway. The loop below implements stages 2 and 3.

When eligible: after posting a `changes_requested` review, **immediately proceed to fix all actionable findings** without asking the user — including the `INHERITED` ones collected in step 2 from the review feedback other reviewers already posted on the PR (`references/pr-metadata.md`), which are fixed on the same severity order as this run's own and reported per-item as fixed, deferred to a follow-up, or declined with a reason. The om-auto-review-pr skill must be fully autonomous — it reviews, fixes, re-reviews, and iterates until the PR is merge-ready or a real blocker remains.

Only stop and ask the user in these critical situations:

- Ambiguous product or architecture decisions that could go multiple valid ways
- Missing credentials, environment access, or infrastructure failures
- Changes that would break public contracts in ways the project's compatibility rules do not allow
- Scope expansion that would fundamentally change what the PR does

For everything else — missing tests, code style issues, i18n problems, type errors, lint failures, missing metadata exports, security hardening — fix them autonomously.

## om-auto-review-pr specifics — fix-forward loop (step 11)

Continue inside the isolated worktree. Do not stop after the first patch. Treat autofix as an iterative loop:

0. **Unit test audit**: Before fixing code findings, check whether the PR includes unit tests for the changed behavior. If the PR has no test files in the diff (`*.test.*`, `*.spec.*`, `__tests__/*`, or the repo's equivalent), add appropriate unit tests as the first autofix action. Every behavior change, bug fix, or new feature must have corresponding test coverage — this is non-negotiable in autofix mode.
1. Convert the current review findings into a concrete fix list.
2. If the PR is currently conflicted, resolve conflicts against the latest base branch first.
3. Implement the next batch of fixable findings.
4. Run validation for the updated code:
   - Run the targeted subset of `validation.commands` relevant to the changed scope (the test and typecheck commands for the affected packages when the toolchain supports scoping; otherwise unscoped).
   - If the review findings touched shared contracts or multiple packages, expand to the full `validation.commands` gate.
5. Re-run the code review on the updated diff in the same worktree.
6. If new or remaining actionable findings exist, repeat from step 1.
7. Stop only when:
   - the re-review outcome is `approved`, or
   - a real blocker remains that cannot be resolved autonomously in the current turn.

Examples of real blockers: ambiguous product or architecture decisions that require user input; environment or infrastructure failures unrelated to the changed code; missing credentials or missing external access.

Conflict-resolution rules for autofix mode:

- Resolve conflicts only inside the isolated worktree or carry-forward branch.
- Never attempt conflict resolution in the user's active worktree.
- Always fetch the latest `{baseRefName}` before resolving conflicts.
- After conflicts are resolved, rerun the relevant validation commands and the code review before deciding the branch is ready.
- If conflict resolution introduces additional findings, continue the autofix loop instead of stopping.

For autofix mode, the goal is not "submit one fix commit". The goal is "finish the PR". Keep iterating until the code review is clean and validation passes, unless a real blocker stops progress.

### Same-repo PRs

If the PR head branch is in the main repository and you have push access, implement the fixes on the checked-out PR branch, resolve any base-branch conflicts there if needed, run the autofix loop above, then commit and push to that branch only after the latest re-review is approvable.

Rules:

- Never force-push unless the user explicitly asked for it.
- Prefer a normal follow-up commit.
- Use conventional-commit-style messages scoped to the affected area: `fix(<area>): <summary>`, `feat(<area>): <summary>`, `refactor(<area>): <summary>`, etc.
- Before pushing, ensure the latest autofix cycle included tests, the targeted validation commands, and a fresh code review on the final diff.

### Fork PRs

For fork PRs, do not wait on the original author and do not push to the contributor's branch by default. Instead, keep the fetched PR head SHA (to preserve authorship), build a `carry/pr-{prNumber}-ready` branch in the main repo, run the autofix loop there, push it, open a replacement PR crediting the original author, and close the original only after the replacement exists. The full step sequence, the autofix validation requirements, the replacement-PR requirements, and the suggested body / handoff / closing-comment templates are in `references/fork-pr-flow.md`.
