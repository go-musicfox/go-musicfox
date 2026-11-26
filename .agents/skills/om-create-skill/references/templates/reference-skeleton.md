# Template — a `references/<content>.md` file (the terrain)

Starting skeleton for a reference file. Name it after its content in kebab-case
(`pr-summary-template.md`, `fork-pr-flow.md`), not after a step number. It opens
with one sentence saying what it is and which skill/step calls it, then holds the
detail moved out of the body.

```markdown
# <Content title>

<One sentence: what this file is and which step of `<skill-name>` opens it —
e.g. "The PR-body template `<skill-name>` opens the PR with in step 9.">

<The detail: the output template / conditional branch / big table / detailed
sub-procedure — moved 1:1 from the body in split mode, or authored fresh in
author mode.>
```

Guidance:

- One coherent concern per file. If a file starts covering two unrelated steps, split it.
- In split mode, the content here is copied word-for-word from the original body — do not re-word instructions.
- Keep the opening sentence honest about *when* it loads, so the body's pointer and this header agree.
