# Per-Step loop (lean, no per-Step chatter)

The per-Step half of step 8. A Step is atomic: one Step = one code commit. Nothing more.

1. **Implement** only the work described by the Step. Never pull work forward from later Steps.
2. **Tests** — add or update tests for anything that changed behavior:
   - Unit tests are mandatory for any code change.
   - Escalate to integration tests for risky flows, permission checks, data scoping, workflows, or behavior that crosses component boundaries.
3. **Quick sanity check** — run the minimum needed to confirm the Step compiles and its own new tests pass (the typecheck and test commands from `validation.commands`, scoped to the affected package or test file when the toolchain supports scoping). Do NOT record these runs anywhere — they are scratch.
4. Re-read the diff and remove scope creep.
5. Re-check changed production files against the project's data-access and security conventions from its agent instructions (for example mandated repository/query helpers, scoping wrappers, or encryption-aware accessors).
6. **Flip the Tasks-table row in the same commit.** In `PLAN.md`'s `## Tasks` table, flip the Step's `Status` cell from `todo` to `done` and fill the `Commit` column with the short SHA (use a placeholder like `pending` in the first write, then amend before push with the real short SHA via `git commit --amend --no-edit` after `git commit` gives you the SHA — or write any unique sentinel and do a fixup). Do not reorder rows, do not rename titles, do not change `Exec` cells — per-Step commits touch only `Status` and `Commit`. No separate docs-flip commit.
7. **Commit** with a clear conventional-commit subject. Example subjects:
   - `feat(ui): add confirmation dialog primitive`
   - `test(ui): cover confirmation dialog focus trap`
8. **Push** after every Step so `om-auto-continue-pr-loop` always has the latest state on the remote.
   - **Step review (`engine.stepReview: per-step` only)** — after the push, review this Step's commit range per `references/step-review.md` before starting the next Step.
9. **Do NOT** write a `step-<X.Y>-checks.md`. **Do NOT** rewrite `HANDOFF.md`. **Do NOT** append to `NOTIFY.md`, unless the Step produced a blocker, a scope decision worth recording, or a subagent delegation. Routine progress is inferred from the Tasks table and the commit log.
