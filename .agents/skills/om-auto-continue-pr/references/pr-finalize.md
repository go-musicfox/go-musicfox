# PR finalize — update the resumed PR, labels, summary comment, markers

The single procedure for the "push → update the existing PR → normalize labels → summary comment → chaining reference lines" mechanics (steps 9 and 10 of the skill body). The point is **one** implementation of PR updating + labeling, reused rather than copied, and never a second PR for work that already has one.

## Never open a duplicate PR

Before opening anything, check whether a PR already exists for this branch via **search-prs** / **get-pr**. If one exists, **reuse it** — push new commits to its head branch and update its body/labels — never open a second PR. Only the skill that first opens the PR owns opening it; everyone else updates that same PR. For this skill the check is already decided: the run resumes PR `#{prNumber}`, so this step **always updates that existing PR and never opens a new one**.

## Prefer the `om-open-pr` skill when installed

`om-open-pr` implements the push + label-normalization mechanics (through the descriptor guards, with rationale comments). When it is installed, **reuse it** instead of re-deriving the steps — this is a resume with no `{issueId}` in scope, so its PR-opening, issue-handback, and lock-release parts do not apply: this skill updates the existing PR and holds its own lock until step 10.

## Graceful fallback when `om-open-pr` is NOT installed

`om-open-pr` is an **optional** enhancement — a repo may install this skill without it, and it must still work. When absent, perform the mechanics inline via the tracker operations: commit with a conventional-commit subject, push the head branch, normalize labels per the section below. Behavior is identical either way — the same PR, the same labels — so installing `om-open-pr` only removes duplication, it never changes the outcome.

## Spec-only PRs stay design-only

A PR whose branch changes only spec/design files (`$SPECS_DIR`, docs areas) is a **spec PR** — an atomic design deliverable. A resume never lands implementation code on it (the skill-body step 5 guard stops before that happens): implementation ships on its **own PR** via `om-auto-implement-spec`, which links the spec PR with `Refs #{prNumber}`. A branch that already mixes spec and implementation code from an earlier run is an implementation PR and is finished as one.

## PR body updates

- If all Progress steps are now `- [x]`, flip `Status: in-progress` to `Status: complete` **and flip the PR from draft to ready via mark-pr-ready** — `om-auto-create-pr` leaves the PR a draft while unfinished, so completing the resume is what promotes it. A resume that stays `in-progress` leaves the PR a draft the user can watch and re-enter.
- Extend the `What Changed` / `Tests` sections with the new work from this resume.

## Label normalization — resume semantics

Every mutation goes through the `apply_label`/`label_exists` guards from the tracker descriptor; when `labels.enabled` is `false`, skip every label operation and say so in the summary comment.

- If the PR is still in a non-terminal pipeline state (`review`, `changes-requested`, `qa`, `qa-failed`, `merge-queue`, `blocked`, `do-not-merge`), keep it. Do NOT move a PR already in `merge-queue` back to `review` just because a resume happened.
- If the PR has no pipeline label (shouldn't happen, but may after an override), apply `review`.
- Add `needs-qa` if the resume introduces user-facing behavior. Add `skip-qa` only for clearly low-risk changes. Never both. If the resume newly introduces user-facing behavior on a PR previously in `merge-queue`, add `needs-qa` and drop any stale `qa-approved` — the QA sign-off no longer covers the new work; when `qaGate` is on, the QA-approval gate re-blocks the merge until QA re-approves. Do not set the `qa` pipeline label yourself; `qa` is applied manually by a QA reviewer when they re-test.
- Preserve the priority label. If the resume materially widens the scope (e.g. now touches auth, money, or data scoping), raise the priority accordingly and comment why; otherwise leave it. If the PR somehow has no priority, infer and apply one: outage, data loss, or a security incident → `priority-extreme`; security hardening or a release-blocking regression → `priority-high`; ordinary bug or feature → `priority-medium`; cosmetic, docs, dependency bumps, or cleanup → `priority-low`.
- Preserve the risk label. If the resume materially widens the blast radius (e.g. now touches auth, money, data scoping, migrations, encryption, event reliability, shared contracts, or spans more areas), raise the risk accordingly and comment why; otherwise leave it. If the PR somehow has no risk label, infer and apply one: auth/session/data scoping/money, migrations, encryption, event reliability, shared contract surfaces, or broad cross-cutting edits → `risk-high`; ordinary single-area change with tests → `risk-medium`; docs, dependency bumps, test-only, or isolated cleanup → `risk-low`.
- Never add `qa-approved` and never set the `qa` pipeline label from this skill. When `qaGate` is on, a `needs-qa` PR may sit in `merge-queue` while the QA-approval gate blocks the merge until a QA reviewer adds `qa-approved`; when `qaGate` is off, `needs-qa` is advisory only.
- Maintain the single consolidated `🏷️ label rationale` comment instead of per-change comments: on any label change, find the `` 🤖 `om-auto-continue-pr` — 🏷️ label rationale `` marker via **list-issue-comments** and rewrite that comment via **update-comment** to the current label state — one label per line with its emoji and a full-sentence reason; create it when absent, never post a second (the format and label emoji map are the shared contract in `om-open-pr`'s label normalization).

## Summary comment

Every resume ends with a single comprehensive summary comment the human reviewer can read top-to-bottom without clicking into the diff, posted in step 9 via the tracker operation **comment-pr** with a body file so multi-line formatting is preserved. Full structure and rules: `references/summary-comment-template.md`. Never post it before the automated review loop (step 8) finishes, never claim a completion you did not reach, and never paste secrets into it. The summary's "Verification phases completed" section is where this resume's validation-gate and integration/UI outcomes land on the PR; when a verification runs mid-flight and is worth surfacing early, post it as its own idempotent `` 🤖 `om-auto-continue-pr` — verification `` comment (screenshots via **attach-image-evidence** when UI changed).

## Marker emission

End the run's final report with the chaining reference lines, one per line, exact shape — include `Issue:` only when the run has a subject issue:

```
Issue: #<issue number> (link: <full issue URL>)
PR: #<PR number> (link: <full PR URL>)
```

Chained consumers (`om-auto-review-pr`, `om-auto-qa-pr`, orchestration scripts) parse these exact text markers — never rename, translate, or decorate them.

## om-auto-continue-pr specifics

- Release the `in-progress` lock at the very end of step 10 — **always**, even on failure (trap/finally). Mechanics and the completion-comment text: `references/claim-pr.md`.
- Remove the worktree you created and prune, in the same trap/finally: `references/worktree-setup.md`.
