# PR finalize — open or reuse, labels, markers

The single procedure for the "commit → push → open (or reuse) the PR → normalize labels → chaining reference lines" mechanics (step 8 of the bug chain, and the feature route's F4 contract check). The point is **one** implementation of PR opening + labeling, reused rather than copied, and never a second PR for work that already has one.

## Never open a duplicate PR

Before opening anything, check whether a PR already exists for this branch (or one that references `#{issueId}`) via **search-prs** / **get-pr**. If one exists, **reuse it** — push new commits to its head branch and update its body/labels — never open a second PR. Only the skill that first opens the PR owns opening it; everyone else updates that same PR. On the feature route an open PR referencing the issue means resume/continue, never a duplicate.

## Prefer the `om-open-pr` skill when installed

`om-open-pr` already implements exactly this: it commits the worktree, pushes the branch, opens a **ready-for-review** PR against `$BASE_BRANCH` (draft only with `--draft`) with the unified body template, applies the full SDLC label set (pipeline `review`, category, QA meta, one priority, one risk) through the descriptor guards with rationale comments, posts the caller's summary comment, and — in an issue-driven run — transfers the chain lock onto the PR when `--handoff` is passed, then hands the issue back and releases the issue's `in-progress` lock, emitting the `PR:` / `Issue:` chaining reference lines. When it is installed, **delegate to it** instead of re-deriving the steps: invoke `om-open-pr {issueId} {category} --handoff om-auto-review-pr` (add `--draft` only for spec-only or incomplete hand-offs; the `--handoff` keeps the PR continuously locked for the step-9 review — see `references/claim-pr.md`, chained hand-off) and capture the PR number and URL from its `PR:` reference line.

## Graceful fallback when `om-open-pr` is NOT installed

`om-open-pr` is an **optional** enhancement — a repo may install this skill without it, and it must still work. When `om-open-pr` is absent, perform the mechanics inline:

1. Commit the worktree changes with a conventional-commit subject; push the branch.
2. Open the PR via the tracker operation **create-pr** against `$BASE_BRANCH`, with a conventional-commit-prefixed title scoped to the primary area and a body that carries the linkage line (`Fixes #{issueId}`), the fix summary, tests added, and the breaking-changes statement.
3. Normalize labels per the section below.
4. Transfer the chain lock onto the PR — **assign-pr** `$CURRENT_USER`, `apply_label "in-progress" {prNumber}`, and the 🤖 hand-off comment naming `om-auto-review-pr` — exactly as `om-open-pr --handoff` would (`references/claim-pr.md`, chained hand-off).
5. Then hand the issue back and release the issue's `in-progress` lock per `references/claim-pr.md`.

Detect availability simply: if invoking `om-open-pr` is not possible in this environment (skill not present), take the inline path. Behavior is identical either way — the same PR, the same labels, the same lock hand-off.

## Ready vs draft

Open the PR **ready for review**. Draft only when the run is explicitly handing off incomplete work (a spec-only design PR, or an interrupted run).

## Label normalization

Apply labels from the config's taxonomy after opening the PR, always through the `apply_label` guard from the tracker descriptor (missing labels degrade to a logged skip; `labels.enabled: false` skips everything — note that in the report). This is the canonical label contract for every PR-opening skill; `om-open-pr` carries the same rules and the two must stay in sync.

- Apply the `review` pipeline label. New PRs always start in `review` unless the run terminated early with an explicit blocker.
- Add `skip-qa` **only** for clearly low-risk non-user-facing changes (docs-only, dependency-only, CI-only, test-only, trivial typos, single-file maintenance).
- Add `needs-qa` when the run touches UI or other user-facing behavior that requires manual exercise.
- Never add both `needs-qa` and `skip-qa`.
- Add additive category labels when they clearly apply: `bug`, `feature`, `refactor`, `security`, `dependencies`, `documentation`.
- Apply exactly one priority label. Infer it from the issue and the diff: outage, data loss, or a security incident → `priority-extreme`; security hardening or a release-blocking regression → `priority-high`; ordinary bug or feature → `priority-medium`; cosmetic, docs, dependency bumps, or cleanup → `priority-low`.
- Apply exactly one risk label. Infer it from the diff: changes to auth, session handling, data scoping, money, DB migrations, or shared contract surfaces, or broad cross-cutting edits → `risk-high`; an ordinary single-area change with tests → `risk-medium`; docs, dependency bumps, test-only, or isolated cleanup → `risk-low`.
- After applying the label set, post **one** consolidated rationale comment covering every applied label — never one comment per label (that spams the PR timeline and multiplies tracker API calls). Labels are still applied individually through the `apply_label` guard; only the commentary consolidates. The comment carries the standard idempotent marker, so a re-run updates it in place.
- When `qaGate` is `true`, a `needs-qa` PR will not be mergeable until QA signs off with `qa-approved`. Do not add `qa-approved` from this skill — it is earned by manual QA or the self-QA exception. State in the PR summary that manual QA is still pending.

Consolidated label-rationale comment — exactly **one** marker-idempotent comment from this skill per PR, listing only the labels actually applied: **one label per line**, each with its emoji from the map below and a full-sentence reason (drop lines for labels not applied; never compress into a `·`-concatenated one-liner). On any later label change, find the marker via **list-issue-comments** and rewrite this same comment via **update-comment** so it always describes the current label state; never post an additional per-change comment (no **update-comment** in the descriptor → post a replacement stating it supersedes the previous rationale):

```markdown
🤖 `om-auto-fix-issue` — 🏷️ label rationale

- 🔍 `review` — ready for code review.
- 🐛 `bug` — {why this category fits, one full sentence}.
- 🧪 `needs-qa` — {why manual QA is needed}.
- 🔺 `priority-high` — {why this priority}.
- 🟡 `risk-medium` — {why this risk}.
```

Label emoji map (decoration only — parsers key on the backticked label text): 🔍 `review` · ❌ `changes-requested` / `qa-failed` · 🧪 `qa` / `needs-qa` · 🚀 `merge-queue` · ⛔ `blocked` / `do-not-merge` · 🐛 `bug` · ✨ `feature` · ♻️ `refactor` · 🔒 `security` · 📦 `dependencies` · 📚 `documentation` · ⏭️ `skip-qa` · ✅ `qa-approved` · 📸 `qa-self-verified` · 🔥 `priority-extreme` · 🔺 `priority-high` · 🔹 `priority-medium` · 🔽 `priority-low` · ⚠️ `risk-high` · 🟡 `risk-medium` · 🟢 `risk-low` · 🤖 `in-progress` · ⏱️ `ci-monitoring`.

## Marker emission

End the run's final report with the chaining reference lines, one per line, exact shape — include `Issue:` only when the run has a subject issue:

```
Issue: #<issue number> (link: <full issue URL>)
PR: #<PR number> (link: <full PR URL>)
```

Chained consumers (`om-auto-review-pr`, orchestration scripts) parse these exact text markers — never rename, translate, or decorate them.

## om-auto-fix-issue specifics

- **Bug route (step 8).** Provide `om-open-pr` the implementer's final summary in the exact block shape it expects (`— PREVIOUS STEP (om-fix) said —` followed by the verbatim summary) and pass `--handoff om-auto-review-pr`. `om-open-pr` transfers the chain lock onto the PR, then hands the issue back to its original author and releases the issue's `in-progress` lock; if it ends with `Status: blocked`, it has already released the issue lock and no PR lock exists — go to the final report and state the blocker.
- **Feature route (F4).** The delegated skills own PR opening; this skill only verifies the contract — exactly one PR references the issue; ready unless a `⚠ NEEDS HUMAN CONFIRMATION` guard applies; full label set present (pipeline state, `feature` category, QA meta, one priority, one risk — re-run the normalization above on anything missing); linkage matches what ships (`Closes #{issueId}` implementing, `Refs #{issueId}` spec-only) — and passes the chaining reference lines through.
- Carry a `LOW_CONFIDENCE` flag from `om-root-cause` into the PR body so a human reviewer looks harder.
