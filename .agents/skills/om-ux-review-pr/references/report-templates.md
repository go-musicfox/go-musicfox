# Report templates

The review comment is the deliverable, not a log. Fill this template exactly
and expand with detail; never improvise a terser variant. Complete sentences,
the why behind every finding, sections structured with the glossary emojis.

## Review comment

Posted as ONE marker-idempotent comment: find the marker via
**list-issue-comments** and update in place via **update-comment**; a re-run
never adds a second review comment to the same PR.

```markdown
🤖 `om-ux-review-pr` — evidence-first design review

## 🔍 Design review — <PR title>

**Contract**: <.uxproof/ found: framework, N tokens, N components | no contract — reviewed against tiers 2-6, no [PRODUCT] findings possible>
**Screens walked**: <list, with viewport(s) and the tasks performed>
**Not walked**: <screens skipped and why — missing data, no permissions, broken env; never omit this line when coverage is partial>

### 🔍 Findings (worst first)

#### 1. <one-line finding title> `<EVIDENCE-TAG>`
- **Where**: <screen and element, with the 📸 evidence reference>
- **Evidence**: <the tagged claim — cite the contract rule, name the standard, or name the heuristic; an assumption is labeled as an assumption>
- **Pattern**: <what the fix looks like — point at an existing screen in this repo that already does it right when one exists>
- **Trade-off**: <what the fix costs; "none" is almost never true>
- **Accept when**: <criterion someone else can verify>

#### 2. …

### 📸 Evidence

<screenshots attached via the attach-image-evidence operation, each captioned with the screen and state it shows>

### ✅ Summary

- **Strong**: <one sentence on what already works and should stay the reference>
- **Must change**: <one sentence naming the findings that clear the impact bar>
- **Opinion**: <one sentence naming what is assumption-tier and safe to overrule>

### 📋 Applied

<One line naming the checks that ran and the ones that did not apply: design
contract, state matrix, humane gate, repo-local rules. It lets the author tell
a check that passed from a check that never ran.>

_Advisory review: findings are input for the author, not a merge gate. This skill applies no labels and blocks no merge._
```

## Rules for filling it

- Rank by impact × frequency × reach, never by ease of fix.
- At most about seven findings; fold the tail into one "minor notes" line.
- Attach every screenshot a finding references; a finding pointing at an
  invisible screen cannot be verified by the author.
- ⚠️ marks a finding that needs a human decision rather than a fix (policy,
  legal, or a deliberate trade-off the team must own).
- When the walk ran without a contract, say so on the Contract line and emit
  no `[PRODUCT]` findings.
