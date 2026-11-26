# Split mode — full procedure (conservative, behavior-preserving)

The full procedure `om-create-skill` follows to refactor an existing oversized
`SKILL.md` into layered `references/` **without changing behavior**. This is the
§9 process from `om-filozofia.md`. The body enters this file for split mode.

## 0. Should it be split at all?

Refuse — and say why — when:

- the skill is under ~150 lines (one coherent thought; splitting adds only a load hop);
- it has no dominant template, big table, or conditional branch to extract;
- every candidate fragment loads on every run anyway.

Only proceed when there is real terrain to move down. Confirm with the user if borderline.

## 1. Map sections to layers

Read the whole `SKILL.md`. Using the up/down rule in `references/philosophy.md`,
tag each section: **stays in body** (contract, workflow skeleton, decision points,
safety rules) vs **moves to references** (output templates, conditional branches,
big tables, detailed sub-procedures).

## 2. Move text 1:1

Copy the moved sections **word-for-word** into `references/<content-name>.md` —
**no re-wording of instruction content**. Each reference file opens with one
sentence: what it is and which step of the skill calls it. Name files after their
content in kebab-case.

## 3. Leave a one-liner + pointer

In the body, at the exact spot the text left, leave a short one-liner that keeps
pointing to *the same moment in the flow*, plus a sentence saying *when and why*
to open the reference. The workflow step must still read as the same step.

Only the section's *intro sentence* may be condensed into the pointer; the
instruction content itself lives verbatim in the reference.

## 4. Preserve the description exactly

**Never change the meaning — ideally not a byte — of the frontmatter
`description`.** It is the routing surface. The gate checks it is unchanged.

## 5. Verify completeness

Confirm every branch, marker, and rule from the original now has a home (body or
reference) — nothing vanished. The mechanical checks are in `references/gates.md`
(fenced code blocks preserved; every moved line reappears in references;
untrusted-content boundary still loads on every run — body or the step-0
`references/agentic-setup.md`; description unchanged).

## 6. Readability + lint gate

Run the readability test (body alone still reads as a recipe) and
`references/gates.md`. Fix and re-run until green.

## 7. Commit convention

One commit per split skill, Conventional Commit style:
`refactor(<skill>): split SKILL.md into references`, with a body noting which
sections moved and the before→after body line count. Do the behavioral
"works-the-same" check (running the skill on a real case) when feasible; note in
the report if it was left for after review.
