# Author mode — full procedure

The full procedure `om-create-skill` follows to turn a brief into a new
`skills/<name>/` skill. The body enters this file for author mode.

## 1. Interview (interactive — ask only what changes the output)

Ask the minimum set of questions whose answers change the generated skill; infer
the rest from the brief and the repo. The ones that matter:

- **Goal + produced result** — one sentence: what the skill does and what it
  leaves behind (a PR, a report, a mutated tracker state, files on disk).
- **Trigger phrases** — how a user or agent will invoke it, in **both PL and
  EN** (this repo's skills carry bilingual triggers). Drives the `description`;
  see `references/description-guide.md`.
- **Tracker interaction** — does it mutate PRs/issues (then it needs the
  three-signal claim/lock and label discipline) or is it read-only?
- **Chain membership** — is it a step in the autofix chain (then it needs the
  `PR:` reference line and the `— PREVIOUS STEP said —` handoff), or
  standalone?
- **Arguments** — required/optional inputs, and any `--flags`.
- **Isolation** — does it need an isolated worktree (any run that builds/tests/
  commits) or does it work in place?

Confirm the derived name is kebab-case, `om-`-prefixed, verb-first (matching the
repo's house style), and not already taken under `skills/`.

## 2. Draft the router body

Start from `references/templates/skill-skeleton.md`. Fill:

- **Frontmatter** — `name` == directory; `description` crafted per
  `references/description-guide.md`.
- **Preamble** — give the new skill its own `references/agentic-setup.md`,
  built by pasting the needed blocks verbatim from
  `references/shared-boilerplate.md` (config load, repo-local extension check,
  untrusted-content boundary, value sanitization), and a two-line step 0 in
  the body that loads it and names the config vars / tracker operations the
  skill uses (the this-skill-uses list). Also give it a `references/rules.md`
  carrying the shared rules (same shape as this skill's
  `references/rules.md`), pointed to from a one-line "Shared rules:" bullet.
  Drop the tracker/label parts for a read-only skill; always keep the
  untrusted-content boundary.
- **Arguments / Contract** — the signature, concisely.
- **Workflow skeleton** — numbered steps as one-liners; each says *what* happens
  and *which reference to open* for detail.
- **Rules** — only the hard, global, safety rules.

## 3. Push detail down to references/

Apply the up/down rule from `references/philosophy.md`. Anything that is an output
template, a conditional branch, a big table, or a detailed sub-procedure starts in
`references/` — not in the body. Reuse existing reference shapes across the repo
(summary-comment, label-normalization, PR-body templates) rather than inventing
parallel ones; name files after their content in kebab-case, each opening with a
one-sentence "what/where-called".

## 4. Scaffold the files

Write:

- `skills/<name>/SKILL.md` (the router body).
- `skills/<name>/references/*.md` (the terrain), each from
  `references/templates/reference-skeleton.md`.
- Optionally `.ai/skills/<name>/SKILL.md` as a repo-local stub from
  `references/templates/repo-local-stub.md` when the skill clearly needs
  per-repo specifics (exact commands, ports, seeded accounts).

## 5. Optional — record a decision

When the new skill introduces a genuinely new capability worth logging, ask the
user whether to add a one-line entry to `DECISIONS.md`. Do not do it unprompted.

## 6. Run the gate

Run `references/gates.md` (lint + readability). Author mode has no split-
completeness step, but the lint and readability checks are mandatory. Fix and
re-run until green before handing back.
