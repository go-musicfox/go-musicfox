# Report templates — run report and failure analysis (steps 9–10)

How `om-integration-tests` reports back to the user, for authoring runs and
running-only mode alike. Reporting style contract: `references/rules.md`
(Reporting style) — full sentences, explain the why, never compress; emojis
structure the sections, the text carries the meaning. This skill defines no
chaining reference lines.

## Run report

```markdown
## 🧪 om-integration-tests — {scope: suite, category, or authored scenario}

**Result:** {✅ all {N} tests pass | ❌ {M} of {N} failed} — {one full sentence on the outcome}
**Environment:** {full sentence: the base URL used, whether you attached to the shared test-env descriptor or provisioned one, and the runner command executed}

### 🧪 Per-test outcomes
- ✅ `{path}::{test name}` — {full sentence: what the test verifies and how it was validated (observed flow, key assertions, screenshots taken)}
- ❌ `{path}::{test name}` — {one sentence naming the failure; the full diagnosis lives in the failure-analysis table below}

### 📋 Authored tests {authoring runs only}
{Full sentences: which test files were created or edited and where they live, which repository conventions they follow (naming, fixtures, helpers), which elements were observed in the live app rather than guessed, and the verification run that proved each new test passes. Omit this section in running-only mode.}
```

## Failure-analysis table (mandatory after any failed run)

Respond with this table **before** any narrative, one row per failing test.
The **Reasoning** column carries the concrete technical diagnosis in full
sentences — grounded in the artifacts you inspected, never a guess:

```markdown
| Failing test | Evidence used | Reasoning (why it failed) | Suggested owner | Next action |
|--------------|---------------|---------------------------|-----------------|-------------|
| `<path>::<test name>` | `output + screenshot + error context` | `Concrete technical diagnosis in a full sentence` | `User/Product team` / `Agent/QA` / `Shared` | `Concrete fix recommendation` |
```

Never give a generic "tests failed" summary without per-test reasoning. After
the table, close with a short paragraph in full sentences: the failure classes
seen (product regression vs test issue vs environment), what you already
fixed, and what needs a human decision.
