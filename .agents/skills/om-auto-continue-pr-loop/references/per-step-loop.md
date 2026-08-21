# Per-Step loop (lean, no per-Step chatter) — step 6a

One Step = one code commit. Nothing more.

1. Implement only the work described by the current Step.
2. Add or update tests for anything that changed behavior. Unit tests mandatory for code changes; escalate to integration tests for risky flows, permissions, data scoping, workflows, or multi-module behavior.
3. Run a quick scratch sanity check (typecheck + new test file, or the closest configured equivalent) to confirm the Step compiles. Do NOT record this anywhere — checkpoints verify.
4. Re-read the diff to remove scope creep.
5. Re-check changed production files against the project's data-access and security conventions from its agent instruction files; fix violations before committing.
6. **Flip the Tasks-table row in the same commit.** In `PLAN.md`'s `## Tasks` table, flip the Step's `Status` cell from `todo` to `done` and fill the `Commit` column with the short SHA (amend the commit to capture the real SHA before pushing). Do not reorder rows, do not rename titles, do not change `Exec` cells — per-Step commits touch only `Status` and `Commit`.
7. **Commit** with a conventional-commit message for that single Step. No separate docs-flip commit.
8. **Push** after the commit so the remote always has the latest state.
   - **Step review (`engine.stepReview: per-step` only)** — after the push, review this Step's commit range per `references/step-review.md` before starting the next Step.
9. **Do NOT** write a `step-<X.Y>-checks.md`. **Do NOT** create a `step-<X.Y>-artifacts/` folder. **Do NOT** rewrite `HANDOFF.md` at the per-Step level. **Do NOT** append to `NOTIFY.md` unless the Step produced a blocker, a scope decision, or a subagent delegation.

Do not alter work already completed in earlier commits. Do not reorder or rewrite history on the PR branch.
