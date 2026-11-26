# Manual-QA instructions comment (step 10)

The additive comment `om-auto-review-pr` posts when the verdict is approved AND
the PR carries `needs-qa` without `skip-qa` — i.e. it was just routed to
`merge-queue` with `needs-qa` retained. It tells the QA reviewer exactly what to
exercise. It does not replace the pipeline-label, claim, or completion comments
— keep all of them. Do not set the `qa` label yourself; the QA reviewer applies
it when they start testing. Skip this step entirely when `labels.enabled` is
`false`.

Build the instructions from the actual diff, not from generic boilerplate:

- Scope the changed surfaces with the changed-file list from **get-pr-diff** for `{prNumber}` and the PR title/body.
- Translate each user-facing change into concrete click paths (routes or screens), the exact actions to take, and the expected outcome to verify.
- Group areas by priority tag: **P0** auth/sessions/data scoping/money/event reliability, **P1** primary user-facing features and UI, **P2** docs/tooling/DX. Use the three-block layout **Where QA should click** / **What human QA should verify** / **What can go wrong** per area.
- For PRs touching web UI surfaces, add perceived-performance checks: cold-load the changed route (screenshot evidence where possible), first useful shell/loading state, interaction responsiveness, mobile viewport.
- Call out edge cases and data-scoping/permission boundaries explicitly (cross-account isolation, permission-gated actions, empty/error states).

Post it as a single comment via the tracker operation **comment-pr** (preserving multi-line formatting):

```markdown
## 🧪 Manual QA instructions (`needs-qa`)

This PR is approved and requires manual QA (`needs-qa`, no `skip-qa`). It is queued in `merge-queue` but the QA-approval gate holds it until `qa-approved` is added. QA reviewer: when you pick it up, move it to `qa` by swapping the labels (remove `merge-queue`, add `qa`), then run the routes below.

### P0 — {area}
**Where to click**
- {route or screen}
- {route or screen}

**What to verify**
- {concrete action → expected outcome}
- {concrete action → expected outcome}

**What can go wrong**
- {concrete regression symptom}
- {data-scoping/permission/edge-case to probe}

### P1 — {area}
**Where to click**
- {route or screen}

**What to verify**
- {concrete action → expected outcome}

**What can go wrong**
- {concrete regression symptom}

### Pass/fail
- All routes pass → remove the `qa` label and add `merge-queue` plus `qa-approved` (this clears the QA-approval gate)
- Any route fails → remove the `qa` label, add `qa-failed`, and leave a comment describing the failure.
```

Rules for this comment:

- Only post it when approving a `needs-qa` PR (approved + `needs-qa` + no `skip-qa`, routed to `merge-queue`). Never post it for a PR with no QA requirement, or one routed to `changes-requested` or any other state.
- When `qaGate` is `false`, keep the routes but replace the gate sentence with a note that `needs-qa` is advisory in this repository.
- Never invent routes, fields, or behavior that the diff does not contain. If a change is hard to exercise manually, say so and give the closest observable check.
- Keep it scoped to THIS PR's changes; do not turn it into a full-app regression script.
- Never paste secrets, tokens, `.env` content, or real credentials into the instructions.
