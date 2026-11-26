# Report templates — publishing the complete information

Two outputs carrying the same content: one **summary comment on the PR**, so the
state is readable on the tracker without the session, and the **session report**.
Both are deliverables, not logs — full sentences that explain the why behind
every outcome, never a compressed key:value dump.

## 1. Summary comment on the PR

Post once per run via **comment-pr**. When a previous `om-pr-autopilot` comment
exists, find its marker via **list-issue-comments** and rewrite that comment via
**update-comment** instead of stacking duplicates. When the tracker descriptor
defines no **update-comment** operation, post a replacement that states it
supersedes the previous `om-pr-autopilot` report.

```markdown
🤖 `om-pr-autopilot` — run at {UTC timestamp}

**Diagnosis on entry**

{the PR State Report from references/diagnose.md}

**Chain executed**

| Step | Skill | Outcome |
|---|---|---|
| 1 | `om-auto-continue-pr` | Finished plan steps 2.3–2.5; 3 commits pushed |
| 2 | `om-auto-fix-pr` | Review approvable after 1 autofix cycle; CI green |
| 3 | `om-auto-qa-pr` | 4 screenshots attached; the order flow was verified |

**State now**

- 🔍 Review: {verdict, and by whom}
- 🧪 CI: {green, or which checks are red and why}
- 🚀 Mergeability: {status}
- 📸 QA: {evidence, or what is still missing}
- 📋 Follow-ups filed: {issue links, or none}

**Labels this PR should carry**

{One label per line, each with its glossary emoji and a full-sentence reason.}

{Applied automatically | ⚠️ This account has no triage rights — the set above could not be applied; maintainer, please apply it.}

**Verdict:** {merge-ready | blocked on X | waiting for QA sign-off | needs a human decision on Y}
```

## 2. Label set

Derive the full intended set from the diff and the repository's agent
instructions, which own the taxonomy. This skill applies the derivation, never
its own vocabulary:

- **pipeline** (exactly one): the review state while under review, the
  changes-requested state on a failed review, the merge-queue state when
  approved and green, the blocked state on a genuine blocker. **Never set the
  manual-QA-in-progress state** — that one is driven by QA reviewers only.
- **category** (additive): whatever the repository defines (bug, feature,
  refactor, security, dependencies, documentation, …).
- **meta**: QA-required for user-facing behavior, QA-skipped for
  docs/dependency/CI/test-only changes. Never both. The evidence label when UI
  evidence was posted.
- **priority** (exactly one, inferred when absent): outage, data loss, or a
  security incident → the highest; security hardening, a release-blocking
  regression, or auth/session/tenant-scope/money/event-reliability work → high;
  an ordinary fix or feature → medium; cosmetic, docs, dependency bumps, or
  cleanup → low. Conflicting signals → the higher one, and say why.
- **risk** (exactly one, inferred when absent): auth/session/tenant-scope/money,
  migrations or schema, encryption, event reliability, shared contract surfaces,
  or broad cross-module edits → high; an ordinary single-module change shipping
  with tests → medium; docs, dependencies, test-only, typo, or cosmetic → low.

Every label the run adds or changes needs its one-line reason in the comment —
that is the collection's label-commentary rule (`references/rules.md`).

**No triage rights:** an account without them cannot apply labels, and the guard
reports a permission error. Do not retry or work around it — list the intended
set in the summary comment, address the maintainer, and carry it in the session
report as an open item.

## 3. Session report

The same content, plus what the comment cannot carry: the operations run, the
files touched per step, and every judgment call the run made autonomously. End
with the chaining reference lines so a following skill can consume them:

```text
PR: #{number} (link: {url})
Issue: #{number} (link: {url})
```

The issue line appears only when the run has a subject issue.

## 4. What the report must never claim

- Never report a gate as passed that was not actually run — name every skipped
  step and why it was skipped.
- Never claim QA passed without attached evidence.
- Never report a merge unless `--allow-merge` was passed **and** the merge
  really happened.
- Never report a label as applied when the guard reported a permission error.
