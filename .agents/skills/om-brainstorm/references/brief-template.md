# Brief template

The payload `om-brainstorm` writes in workflow step 6 (ramps 2–5), consumed cold by the routed skill. Fill every section; write "none" rather than deleting a heading — humans and downstream skills key on the structure.

```markdown
# {one-line goal}

- Date: {YYYY-MM-DD}
- Category: {feature | bug | refactor | security | dependencies | documentation}
- Priority signal: {low | medium | high | extreme} — {one-line why}
- Risk signal: {low | medium | high} — {one-line why}
- Routing: {the emitted Next line, verbatim}

## Problem

{2–5 sentences in the user's sharpened words; evidence it matters}

## Agreed direction

{what to pursue — and what was explicitly rejected, including why "build nothing" lost}

## Resolved unknowns

| Question | Answer (from the conversation) |
|----------|--------------------------------|
| {…} | {…} |

## Non-goals

- {explicit exclusions, so nobody gold-plates}

## Affected areas (if known)

- {only what the conversation established — never guessed}
```

Notes:

- **Category** follows the SDLC category taxonomy, so downstream label inference works unchanged.
- **Priority/Risk signals** feed `om-prepare-issue`'s `--priority`/`--risk` on ramp 2; on the other ramps they are context for the implementer.
- **Resolved unknowns** is the load-bearing section: on ramp 3 these answers replace `om-auto-write-spec`'s autonomous defaults, and on ramp 4 they pre-answer `om-spec-writing`'s Open Questions gate. An empty table on ramp 3 means the routing is wrong — go back to ramp 4 or keep talking.
- **Affected areas** stays honest: only what the conversation established. The routed skill re-derives the rest from the codebase.
