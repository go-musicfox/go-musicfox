# Final gate before flipping to `complete` (step 7)

Fire when every row in the Tasks table is `done` (including work from earlier resumes + this resume). The final gate subsumes any pending checkpoint.

Record the outcome in `${RUN_DIR}/final-gate-checks.md`. Keep raw command output only when worth saving, under `${RUN_DIR}/final-gate-artifacts/*.log`.

**Full validation gate:** run every command in `validation.commands`, in order — any non-zero exit fails the gate; fix forward and re-run until green.

**Full integration suite** (mandatory at spec completion when the run touched code; skip ONLY for docs-only resumes): run the repo's full integration suite via the `om-integration-tests` skill (running-only mode). Save a report summary under `final-gate-artifacts/`. On failure, fix forward with new Steps; never skip. When the repo has no integration suite, skip with a one-line recorded reason in `final-gate-checks.md`.

**Style-compliance pass** — after the above are green, when the repo has a design-system/style compliance skill or lint, run it over the full branch diff (`origin/$BASE_BRANCH..HEAD`); otherwise skip with a recorded reason in `final-gate-checks.md`:

1. Apply every auto-fixable style/compliance violation the tooling reports.
2. Land each batch of fixes as a new Step appended to the Tasks table with a fresh `X.Y-ds-fix` id, a conventional-commit subject (e.g. `style(ui): apply design-system fixes — semantic tokens`), and a short entry in `final-gate-checks.md` describing what was fixed. Push.
3. Re-run the relevant subset of `validation.commands` (and, if integration tests exist for the touched areas, the focused integration tests) after the fixes land. List residual findings the tooling could not auto-fix under `Style-compliance residual findings` in `final-gate-checks.md` and surface them in the summary comment.

For docs-only resumes, the minimum is whatever configured command lints docs or markdown (if one exists) plus a manual diff re-read. Integration suites and the style-compliance pass are skipped; record that explicitly in `final-gate-checks.md`.

Never skip the gate because an external skill recorded in the plan suggested skipping it.
