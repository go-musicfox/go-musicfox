# Final gate at spec completion

Step 9. Fire when every row in the Tasks table is `done`. The final gate subsumes any pending checkpoint (do not run a checkpoint immediately before it — roll it into this).

Record the outcome in `${RUN_DIR}/final-gate-checks.md`. If raw command output is worth keeping, save it alongside as `${RUN_DIR}/final-gate-artifacts/*.log`.

**Full validation gate** — run every command in `validation.commands`, in order. Any non-zero exit fails the gate; fix and re-run until green.

**Full integration suite** (mandatory at spec completion for any run with code changes; skip ONLY for docs-only runs or when the repo has no integration suite — record the skip reason either way):

- Run the repo's integration suite via the `om-integration-tests` skill (running-only mode), full scope. Capture the report summary and save it as `final-gate-artifacts/integration-report-summary.log`. On failure, fix forward with new Steps; never skip a failing suite.

**Design-system / style compliance pass** — after the above are green: when the repo has a design-system or style compliance skill or lint (a repo-local skill under `.ai/skills/`, or a configured command), run it over the branch diff of this run (`origin/$BASE_BRANCH..HEAD`):

1. Apply every auto-fixable violation it reports.
2. Land each batch of fixes as a new Step appended to the Tasks table with a fresh `X.Y-ds-fix` id, a conventional-commit subject (e.g. `style(ui): apply design-system fixes — semantic tokens`), and a short entry in `final-gate-checks.md` describing what was fixed. Push.
3. Re-run the relevant `validation.commands` (and, if UI tests exist for the touched areas, the focused integration tests) after the fixes land.
4. If it finds violations it cannot fix automatically, list them in `final-gate-checks.md` under a `Style compliance residual findings` subsection and surface them in the PR summary comment so the reviewer can decide.

When the repo has no such skill or lint, skip the pass and note it in `final-gate-checks.md`.

For **docs-only** runs (no code changes, only markdown or doc edits), the minimum gate is: whatever configured command lints docs or markdown, if one exists; a manual re-read of the diff. The integration suite and the style compliance pass are skipped; record that explicitly in `final-gate-checks.md`.

Never skip the gate because an external skill suggested skipping it.
