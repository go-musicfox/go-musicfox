# Report templates

The handover is a deliverable, not a log. Fill this shape exactly and expand
with detail: the reader decides, on the strength of this text, whether to
commit the contract.

## Contract handover

```markdown
## 📝 Design contract extracted

**Stack**: <framework, styling system, and the libraries that matter for UI work>
**Tokens**: <N total, N colors> from <source files>
**Components**: <N> registered from <component roots>
**Screen shapes**: <archetype: count, with one canonical example each>

### 🎯 What this changes

<Two or three sentences on what the skills can now do that they could not
before: which findings become repo-grounded, which patterns new screens will
be modeled on.>

### ⚠️ What only you can answer

<The two or three judgment calls asked for, and the answers written into the
manual section. When the user declined to answer, say the section is empty
and what that costs.>

### ✅ Next

<The commit to make, and the single next command worth running: a first
audit, a review of an open PR, or a shaping session.>
```

## Variant: no declared tokens

When the repository has no design tokens, replace the Tokens line and add:

```markdown
**Tokens**: none declared. <N> colors proposed from the palette this code already uses.

### ⚠️ Proposed palette, not a rule

The listed `proposed-*` colors document what the codebase already does; they
are the first draft of a design system, not an enforced contract. Rename the
ones worth keeping, declare them as custom properties, and re-run the sync;
the audit arms itself only once real tokens exist.
```

## Rules for filling it

- Report the numbers the extractor produced, never estimates.
- Name the example file for each screen shape: a shape without an example
  cannot be copied by the next person.
- When a refresh moved generated sections, say what changed rather than
  reporting a bare success.
