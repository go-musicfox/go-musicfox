# Report templates — user-facing output (step 6)

How `om-prepare-issue` reports back to the user. Reporting style contract:
`references/rules.md` (Reporting style) — full sentences, explain the why,
never compress; emojis structure the sections, the text carries the meaning.

## Final run report

```markdown
## 🎯 om-prepare-issue — {one-line restatement of the brief}

**Issue mode:** {new issue filed | reused existing issue — comment added | filed fresh with a link to the closed duplicate} — {one full sentence: why this was the right call, e.g. which duplicate was credible or why nothing matched}

### 🏷️ Labels
{One label per line with its emoji and a full-sentence reason — mirror the
rationale comment posted on the issue; when labels are disabled in config, say
so here instead of listing labels:}
- 🐛/✨ `{category}` — {why the brief clearly is this category, full sentence}
- 🔥/🔺/🔹/🔽 `{priority-*}` — {why this priority was inferred or overridden, full sentence}
- ⚠️/🟡/🟢 `{risk-*}` — {why the eventual change carries this blast radius, full sentence}

### 📝 Spec
{One short paragraph: which spec covers the task and where it lives (repo path,
plus the spec PR number when one was authored in step 3 or found in flight), or
that no spec covers it and step-level analysis was embedded in the issue body
instead — and why that was the appropriate depth for this task.}

### 📸 Evidence
{One or two full sentences: how many images were attached and where they render
on the issue, that local paths were referenced because inline upload was
unavailable, or that the brief came with no images.}

### 🔍 Duplicates checked
{Full sentences: which search queries ran (issues and open PRs), which top
candidates were read and considered, and why each was rejected — or which one
was reused and what new detail the comment added.}
```

End the report with the chaining reference lines on their own lines, exact
shape — the one part never decorated, reworded, or wrapped in extra formatting.
The `Issue:` line is machine-parsed by `om-auto-fix-issue`'s brief mode, so it
must appear exactly like this:

```text
Issue: #<number> (link: <full issue URL>)
Spec: <repo-relative spec path>            <- only when a spec was linked or authored
PR: #<number> (link: <full PR URL>)        <- only when step 3 authored a spec PR
```
