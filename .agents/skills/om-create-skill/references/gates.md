# The hard gate — lint + completeness + readability

`om-create-skill` must not hand back a skill until these pass. Run them before
reporting; on `--dry-run`, describe what would run and write nothing.

## Gate 1 — lint (both modes, mandatory)

Run the repo's content gate and require a clean exit:

```bash
bash scripts/lint.sh    # must print "Lint OK" and exit 0
```

It verifies frontmatter (`name` == directory, non-empty `description`), the
product-agnostic forbidden patterns, and the no-direct-tracker-CLI rule across
`skills/` (including `references/`). A failure names the offending file and
pattern — fix and re-run.

## Gate 2 — split-mode completeness (split mode only)

Prove the refactor lost nothing and preserved the risk surface. Set `BASE_REF` to
the ref holding the pre-split `SKILL.md` (e.g. the base branch, or `HEAD` before
you started editing), and `SKILL` to the skill name:

```bash
SKILL=<skill-name>
BASE_REF=<ref-with-pre-split-version>
dir="skills/$SKILL"

# 2a. Every fenced code-block line from the original body still exists in the
#     body+references union (commands/snippets/templates are the risk surface).
git show "$BASE_REF:$dir/SKILL.md" | awk '/^```/{f=!f;next} f{print}' | sort -u > /tmp/base_code.txt
cat "$dir/SKILL.md" "$dir"/references/*.md 2>/dev/null | awk '/^```/{f=!f;next} f{print}' | sort -u > /tmp/new_code.txt
missing_code=$(comm -23 /tmp/base_code.txt /tmp/new_code.txt | grep -vcE '^[[:space:]]*$')
echo "code lines lost: $missing_code   (must be 0)"

# 2b. Every non-blank line removed from the body reappears in references
#     (heading level normalized; substring match tolerates reworded pointers).
git show "$BASE_REF:$dir/SKILL.md" | grep -vE '^[[:space:]]*$' | sort -u > /tmp/base_lines.txt
grep -vE '^[[:space:]]*$' "$dir/SKILL.md" | sort -u > /tmp/new_lines.txt
cat "$dir"/references/*.md 2>/dev/null > /tmp/refs.txt
comm -23 /tmp/base_lines.txt /tmp/new_lines.txt | while IFS= read -r line; do
  norm=$(printf '%s' "$line" | sed -E 's/^#+[[:space:]]*//')
  grep -Fq -- "$norm" /tmp/refs.txt || echo "REVIEW (only intro/pointer prose is acceptable here): $line"
done

# 2c. Safety boundary still loads on every run (body, or the step-0
#     references/agentic-setup.md the body always loads first).
grep -q "Untrusted content boundary" "$dir/SKILL.md" "$dir/references/agentic-setup.md" 2>/dev/null \
  && echo "untrusted boundary: loads on every run" || echo "FAIL: safety boundary lost"

# 2d. Description unchanged (routing must not move).
diff <(git show "$BASE_REF:$dir/SKILL.md" | sed -n '/^---$/,/^---$/{/^description:/p}') \
     <(sed -n '/^---$/,/^---$/{/^description:/p}' "$dir/SKILL.md") \
  && echo "description: unchanged" || echo "FAIL: description changed"
```

Pass condition: `code lines lost: 0`, no `FAIL:` lines, and every 2b `REVIEW`
line is only a section-intro sentence you deliberately condensed into a pointer
(not a rule, template, or command). If any real content is missing, put it back.

## Gate 3 — readability (both modes)

Read the body alone. It passes only if you can still tell **what** the skill does,
**in what order**, and **where** to look for detail. If the body became a bare
list of links with no flow, pull some content back up (per
`references/philosophy.md`).

## On failure

Fix and re-run — never hand back a skill with a failing gate. Report the final
gate output (lint result, completeness counts) so the user sees the skill landed
clean.
