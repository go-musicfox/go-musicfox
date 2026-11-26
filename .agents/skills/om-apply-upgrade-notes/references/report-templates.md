# Report templates — user-facing output (step 5)

How `om-apply-upgrade-notes` reports back to the operator. Reporting style
contract: `references/rules.md` (Reporting style) — full sentences, explain the
why, never compress; emojis structure the sections, the text carries the
meaning. This skill defines no chaining reference lines (see the specifics
section of `references/rules.md`), so the report ends with the **Next** line,
not a `PR:`/`Issue:` marker.

## Final run report

```markdown
## 🔁 om-apply-upgrade-notes — tracker `{tracker}` · browser `{provider}`

**Result:** {✅ already current — nothing changed | ✅ changes applied, left uncommitted for review | ⚠️ gaps need operator action} — {one full sentence describing the overall outcome and why}
**Mode:** {applied | dry-run — everything below is what a real run would do; nothing was changed}

### ✅ Tracker descriptor — `.ai/trackers/{tracker}.md`
{Full sentences: which operations were added and why the upgraded skills need them; which stock sections were replaced with the shipped version; which local customizations were detected and kept untouched; which conflicts the operator resolved and how. When nothing changed, say the descriptor is already current instead of listing "none" rows.}

### ✅ Browser descriptor — `.ai/browsers/{provider}.md`
{Same treatment: operations added, replaced, and kept, each with the reason. When the descriptor had to be created, or `browser.provider` defaulted to the legacy provider, explain what was installed and why that default preserves existing behavior.}

### 📋 Config changes — `.ai/agentic.config.json`
{One sentence per key added, naming the documented default used and the UPGRADE_NOTES entry that introduced it. When no keys were added, say the config already carries everything the current skills expect.}

### ⚠️ Custom-provider gaps
{Only when a custom tracker or browser descriptor lacks operations the shipped TEMPLATE now requires: one bullet per missing operation with its contract text, plus a sentence stating that you reported the gap rather than inventing an implementation for someone else's provider. Omit this section entirely when there are no gaps.}

### 📋 Notable-upgrade entries
{One short paragraph: how many entries were checked, how many were applied, how many were already current — and, concretely, which entries still need an operator decision and what that decision is.}

**Next:** {Full sentence: review the diff and commit (for example via the `om-check-and-commit` skill), then re-run the skill that degraded. After a dry run: re-run without `--dry-run` to apply.}
```

Fill every section with what actually happened in this run — real operation
names, real paths, real UPGRADE_NOTES entries — and expand with detail rather
than compressing to keywords. A reader who did not watch the run must be able
to see exactly what state their installation is in and what, if anything, they
still have to do.
