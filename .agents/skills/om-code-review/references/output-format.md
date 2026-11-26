# Review report output format

The report skeleton `om-code-review` produces at workflow step 8. Callers
(`om-auto-review-pr`, `om-review-prs`, the self-review steps of
`om-auto-create-pr` / `om-auto-continue-pr`) consume the verdict and the
blocker/major findings from this structure — keep the section headings and the
severity names exactly as written (emojis decorate; parsers key on the text).

This report is a **deliverable read by humans on the PR**, not a log line.
Write complete sentences everywhere. Every finding explains what is wrong,
where, why it matters, and how to fix it — a reviewer who has not opened the
diff must understand each point on its own. Never compress the report to save
tokens, and never shrink a section to a bare verdict when there is reasoning
worth showing.

Use this structure for every review:

```markdown
# 🔍 Code Review: {PR title or change description}

## 🎯 Summary
{A short paragraph, in full sentences: what the change does, how it does it,
and your overall assessment — including what is good about it, not only the
problems. State the reviewed scope (files/areas) when it is not obvious.}

## Verdict
{✅ approve | ❌ request changes} — {a full sentence justifying the verdict:
which findings drive it, or why the change is safe to merge}

## 🧪 Validation Gate

| Command | Status | Notes |
|---------|--------|-------|
| {validation.commands[0]} | ✅ PASS / ❌ FAIL | {what failed and the relevant output, or blank} |
| {…one row per configured command, in order} | | |

## Findings

### ⛔ Blocker
{Must fix before merge — security, data integrity, data scoping, contract
breaks, failing gates. For each: `file:line` — what is wrong, the concrete
failure it causes, and the fix to make.}

### ⚠️ Major
{Correctness bugs, architecture violations, missing regression tests, weakened
assertions. Same shape: `file:line`, what, why it matters, fix.}

### 🔹 Minor
{Convention violations, suboptimal patterns, readability — each with a concrete
suggestion.}

### 💅 Nit
{Style suggestions, optional polish. Author's call.}

## 💥 Breaking Changes
- [ ] No exported/public symbol removed or renamed without a deprecation path
- [ ] No function signature changed in a breaking way (required params removed or reordered, return type changed)
- [ ] No required type field removed or narrowed
- [ ] No HTTP route URL removed or renamed; no method changed for an existing operation
- [ ] No field removed or retyped in an existing response shape
- [ ] No event or message name renamed or removed; no payload field removed
- [ ] No CLI command or flag renamed or removed; no machine-parsed output format changed
- [ ] No database table or column renamed or removed; no column type narrowed
- [ ] No config key renamed and no default changed silently
- [ ] Where a contract had to change: old surface kept working through a deprecation window, with migration notes

## 🧪 Test Coverage
{What the tests cover and how; then the gaps, with the exact test cases to add
and where they belong. "Covered" alone is not a coverage statement — say what is
covered.}
```

Rules:

- Omit empty severity sections. Mark passing checklist items with `[x]` and
  failing with `[ ]` plus an explanation in full sentences.
- Findings carry severity, `file:line`, rationale, and a concrete fix
  suggestion — a finding the author cannot act on is not a finding.
- The Summary and Verdict lines are what a skimming maintainer reads first;
  they must stand on their own without the sections below.
- Never paste secrets, tokens, or credentials into the report, even when
  quoting offending lines — redact the values.
