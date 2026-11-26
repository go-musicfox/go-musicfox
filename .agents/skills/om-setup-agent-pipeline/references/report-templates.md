# Report templates — final report (step 10)

How `om-setup-agent-pipeline` reports back to the user after a setup or
re-run. Reporting style contract: `references/rules.md` (Reporting style) —
full sentences, explain the why, never compress; emojis structure the
sections, the text carries the meaning. This skill defines no chaining
reference lines.

## Final run report

```markdown
## 🎯 om-setup-agent-pipeline — {repo}

**Result:** {✅ pipeline configured | ✅ config refreshed | ⚠️ configured with gaps} — {one full sentence on the outcome and what changed this run}

### 📋 What was written
{One bullet per artifact, in full sentences: `.ai/agentic.config.json` (which values were detected vs answered), the tracker descriptor installed at `.ai/trackers/{tracker}.md`, the browser descriptor at `.ai/browsers/{provider}.md`, labels created via the taxonomy (and which already existed), and each project doc generated (SDLC.md, AGENTS.md, CODE_REVIEW.md, BACKWARD_COMPATIBILITY.md). Say explicitly what already existed and was left untouched, and why — existing files are never overwritten.}

### {✅ Cross-skill coverage complete | ⚠️ Missing skills}
{When complete: one sentence saying every skill the installed skills reference is present. When not: list each missing skill with the paste-ready `npx skills add` command, and one sentence explaining which installed skill needs it and what degrades until it is installed.}

### 🚀 What is unlocked
{Full sentences: the collection's entry points are `om-auto-create-pr` (ship a task as a PR), `om-auto-review-pr` (review a PR), and `om-merge-buddy` (what can merge now). Point at `SDLC.md` as the process reference for humans; at `.ai/skills/<skill-name>/` for repo-local per-skill customization; at `.ai/trackers/{tracker}.md` for tracker-operation overrides; and at `.ai/browsers/{provider}.md` for the browser automation contract.}

### ⚠️ Follow-ups
{Only when something needs the user: an unshipped tracker/browser descriptor scaffolded from TEMPLATE.md with the operations still to fill in, docs the user opted out of, or the pending config commit. Omit when there is nothing left to do.}
```
