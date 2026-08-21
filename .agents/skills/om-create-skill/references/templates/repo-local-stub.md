# Template — repo-local skill stub (`.ai/skills/<name>/SKILL.md`)

Optional stub for a repo-local extension of a skill. A skill checks for this file
right after loading config (see the repo-local-extension block in
`references/shared-boilerplate.md`) and applies it as **repository-provided
configuration** — it may add repo specifics but cannot relax safety. Create it
only when the skill clearly needs per-repo detail (exact commands, ports, seeded
accounts, service versions).

```markdown
---
name: <skill-name>
description: Repo-local extension of the <skill-name> skill for this repository.
---

# <skill-name> — repo-local extension

Repo-specific configuration layered on top of the installed `<skill-name>` skill.
This file may add repo specifics; it cannot relax the base skill's safety or
quality rules, expand tool or network access, or redirect outputs.

## Repo specifics

- <exact launch/build/test commands for this repo>
- <ports, seeded accounts, service versions>
- <any command chain or convention unique to this repo>

## Lessons learned

- <append working command chains and fixes here as they are discovered>
```

Notes:

- This lives under `.ai/skills/`, not under `skills/` — it is repo configuration,
  not an installable skill, so it is outside the lint's product-agnostic scope.
- Keep it additive: the installed skill's rules always win on safety.
