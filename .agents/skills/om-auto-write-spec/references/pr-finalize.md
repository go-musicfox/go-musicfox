# PR finalize — open or reuse, labels, evidence, summary comment, markers

The single procedure for the "commit → push → open (or reuse) the PR → normalize labels → evidence → summary comment → chaining reference lines" mechanics (steps 6 and 7 of the skill body). The point is **one** implementation of PR opening + labeling, reused rather than copied, and never a second PR for work that already has one.

## Never open a duplicate PR

Before opening anything, check whether a PR already exists for this branch (or, in an issue-driven run, one that references the issue) via **search-prs** / **get-pr**. If one exists, **reuse it** — push new commits to its head branch and update its body/labels — never open a second PR. Only the skill that first opens the PR owns opening it; everyone else updates that same PR.

## Prefer the `om-open-pr` skill when installed

`om-open-pr` already implements exactly this: it commits the worktree, pushes the branch, opens a **ready-for-review** PR against `$BASE_BRANCH` (draft only with `--draft`) with the unified body template, applies the full SDLC label set (pipeline `review`, category, QA meta, one priority, one risk) through the descriptor guards with rationale comments, posts the caller's summary comment, and (in an issue-driven run) hands the issue back and releases the `in-progress` lock — emitting the `PR:` / `Issue:` chaining reference lines. When it is installed, **delegate to it** instead of re-deriving the steps:

- Issue-driven run: invoke `om-open-pr {issueId} documentation --title "docs(specs): ${TITLE}"` (add `--draft` when the high-stakes guard applies) and capture the PR number and URL from its `PR:` reference line.
- Brief-driven run (no issue): invoke it without `{issueId}`; the issue-handback and lock-release parts don't apply.

## Graceful fallback when `om-open-pr` is NOT installed

`om-open-pr` is an **optional** enhancement — a repo may install this skill without it, and it must still work. When `om-open-pr` is absent, perform the mechanics inline:

1. Commit the worktree changes with a conventional-commit subject; push the branch.
2. Open the PR via the tracker operation **create-pr** against `$BASE_BRANCH`, with the spec-PR body template below.
3. Normalize labels per the section below.

Detect availability simply: if invoking `om-open-pr` is not possible in this environment (skill not present), take the inline path. Behavior is identical either way — the same PR, the same labels — so installing `om-open-pr` only removes duplication, it never changes the outcome.

## Ready vs draft

Open the PR **ready for review** — the spec PR is the finished deliverable of this run, not a work-in-progress handoff. Draft only when the high-stakes guard applies: any assumption carrying `⚠ NEEDS HUMAN CONFIRMATION` converts (or keeps) the PR draft, and the body states that merge is gated on confirming those assumptions.

## PR body

Title: `docs(specs): ${TITLE}`. Use the spec-PR variant of the unified template:

```markdown
Refs #{issueId}            <!-- issue-driven runs only; never `Closes` —
                                merging a spec must not close the FR -->
Source doc: ${SPEC_PATH}
Status: complete           <!-- draft/gated runs state the ⚠ merge gate here -->

## 🎯 Goal
- {one-line feature summary from the brief/issue}

## What Changed
- Spec document at `${SPEC_PATH}`
- Mockups: {list | skipped — {reason}}

## 💥 Breaking Changes
- None — design only
```

## Label normalization

Apply labels from the config's taxonomy after opening the PR, always through the `apply_label` guard from the tracker descriptor (missing labels degrade to a logged skip; `labels.enabled: false` skips everything — note that in the summary comment).

For a spec PR the set is: `review` (pipeline), `documentation` (category), `skip-qa` (docs-only — a spec PR changes no runtime behavior), exactly one priority (inferred from the brief/issue), exactly one risk (typically `risk-low` for design-only changes). Never both `needs-qa` and `skip-qa`; never `qa-approved` from this skill. After applying the set, post exactly **one** marker-idempotent consolidated rationale comment — never one comment per label (that spams the PR timeline and multiplies tracker API calls). Labels are still applied individually through the `apply_label` guard; only the commentary consolidates. **One label per line**, each with its emoji and a full-sentence reason; on any later label change, find the marker via **list-issue-comments** and rewrite this same comment via **update-comment** — never post an additional per-change comment:

```markdown
🤖 `om-auto-write-spec` — 🏷️ label rationale

- 🔍 `review` — ready for specification review.
- 📚 `documentation` — lands a spec document, design-only.
- ⏭️ `skip-qa` — docs/design-only, no runtime behavior to QA.
- 🔹 `priority-medium` — {why this priority}.
- 🟢 `risk-low` — {why this risk}.
```

## Evidence comment

After the PR exists, publish the step-5 visuals via **attach-image-evidence**: `{prNumber}`, a short scenario report (a table mapping each image to its screen and current/proposed role), slug `spec-${SLUG}`, and the image paths — so they render inline on the PR, the same mechanism `om-auto-qa-pr` uses. When the descriptor cannot render inline, it still posts the comment with links — surface the limitation, don't fail.

## Summary comment

Every run ends with a single comprehensive summary comment the human reviewer can read top-to-bottom without clicking into the diff. Post it via the tracker operation **comment-pr** with a body file so multi-line formatting is preserved. Never claim a completion you did not reach, and never paste secrets into it. Structure:

```markdown
## 🤖 `om-auto-write-spec` — run summary

**Spec:** ${SPEC_PATH}
**Branch:** spec/${SLUG}
**Final status:** {complete | draft — merge gated on ⚠ assumptions}

### 📝 Assumptions applied
- {count of autonomous defaults + any ⚠ NEEDS HUMAN CONFIRMATION rows — or "spec had no Open Questions"}

### 📸 Visual evidence
- {mockup/screenshot inventory — or why visuals were skipped}

### 🔁 Hand-off
Implement with: `om-auto-implement-spec ${SPEC_PATH}` (or `om-auto-continue-pr {prNumber}` for spec-only continuation).
```

## Marker emission

End the run's final report with the chaining reference lines, one per line, exact shape — include `Issue:` only when the run is issue-driven:

```
Issue: #<issue number> (link: <full issue URL>)
PR: #<PR number> (link: <full PR URL>)
Spec: <repo-relative spec path>
```

Chained consumers (`om-auto-implement-spec`, `om-auto-review-pr`, orchestration scripts) parse these exact text markers — never rename, translate, or decorate them.
