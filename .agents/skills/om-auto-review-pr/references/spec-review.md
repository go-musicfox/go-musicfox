# Specification review (spec-only PRs)

A spec-only PR contains a design document and no implementation — the questions
that matter are design questions, not code gates. This fork replaces the
code-review pass (SKILL.md steps 7–8) for spec-only PRs; every other step —
claim, worktree, verdict, labels, autofix loop, release, report — applies
unchanged.

## Detection — `SPEC_ONLY`

Set `SPEC_ONLY=true` during step 2 (PR metadata) when **every** changed file in
the PR lives under the repo's specs directory (`paths.specs`, default
`.ai/specs`) or the repo's design-doc areas — including their asset folders
(mockups, screenshots). One source, config, CI, or dependency file in the diff
means the PR is **not** spec-only: run the normal code review, and review the
spec changes as context for it. When in doubt, prefer the code path — it is the
stricter gate.

## Grounding before judging

Read the spec end-to-end first. Then verify its claims against the actual
codebase in the worktree: the modules it says it touches, the contracts it says
exist, the primitives it assumes are available. A spec that mis-describes the
codebase is a **blocker** — implementation built on it inherits the error. Also
read `BACKWARD_COMPATIBILITY.md` (repo root, when present) and the repo's spec
skeleton conventions (the `om-spec-writing` skill's structure, when installed).

## The five review lenses

Evaluate ALL five; record findings on the standard **blocker / major / minor /
nit** scale so step 9's verdict rule applies verbatim.

1. **💥 What can go wrong** — failure modes, edge cases, race conditions, and
   security/privacy/data-loss risks the design creates; unstated error paths;
   load or scale assumptions that break; what happens on partial failure or
   rollback.
2. **🔁 Backward compatibility** — every protected contract surface the design
   touches (APIs, schemas and migrations, events, URLs, CLI, stored data —
   per `BACKWARD_COMPATIBILITY.md` when present). A breaking change without a
   named migration/deprecation path is a **blocker**; an undeclared touch of a
   protected surface is at least a **major**.
3. **🧩 What's missing** — unresolved Open Questions, absent acceptance
   criteria, missing non-functional requirements (performance, permissions,
   i18n, observability), no testing or rollout plan, affected areas the spec
   does not mention, undefined failure-handling.
4. **📈 How can this specification be improved** — clarity and structure (does
   it follow the repo's spec skeleton), measurable acceptance criteria instead
   of vague goals, tighter scoping, explicit non-goals, testability of each
   requirement.
5. **✂️ Is this the simplest possible solution — or should something be
   rethought?** Could a smaller design deliver the same outcome? Does the repo
   already have primitives the spec proposes to rebuild (search before
   claiming)? Which parts could be deferred (YAGNI)? When you claim a simpler
   alternative exists, sketch it concretely — "this could be simpler" without a
   shape is a nit, a worked simpler alternative is a major.

## Report, verdict, and autofix

- Use the same structured review-report body as the code path, with one section
  per lens (keep the lens emojis as section markers); each finding carries its
  severity tag. Title the report `Specification review:`.
- Verdict mapping is identical to step 9: any blocker → request changes; major
  without documented waiver → request changes; only minors/nits → approve.
- **Validation gate:** run only the `validation.commands` that apply to
  markdown/docs (linters, link checkers); skip the rest and list the skipped
  commands in the report — a spec PR has nothing to compile or test.
- **Autofix loop (step 11) on a spec PR edits the specification document
  itself** — resolve findings by amending the spec text, mockups, or Open
  Questions sections. It must NEVER add implementation code to a spec-only PR;
  a finding that requires implementation to resolve is reported, not fixed.
  Fork-PR carry rules apply unchanged.
- Labels: normal pipeline transitions; spec PRs are docs-only, so the QA meta
  is typically `skip-qa` — do not add `needs-qa` for a document.
