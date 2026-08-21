# Unified PR body template (step 5)

The one PR body shape every pipeline skill uses. `om-auto-create-pr`'s
`references/pr-body-template.md` is the inline-fallback copy of this contract —
the two files must stay in sync (see the cross-skill contract in this repo's
AGENTS.md). Sections marked *conditional* appear only when their input exists.

Title: `<prefix>(<area>): <one-line summary>${issueId:+ (#issueId)}` — prefix
from the category (`bug` → `fix`; else the category name; `feat` for features).

```markdown
Closes #{issueId}                       <!-- issue-driven implementing PR; use
                                             `Refs #{issueId}` for a spec-only
                                             design PR so merging it does not
                                             close the issue; omit when no issue -->
Tracking plan: {RUNS_DIR}/{DATE}-{SLUG}.md   <!-- conditional: --plan; MUST be
                                                  present when a plan exists so
                                                  om-auto-continue-pr can resume -->
Source doc: {SPECS_DIR}/{spec}.md            <!-- conditional: spec-driven runs -->
Status: in-progress                          <!-- flip to `complete` when every
                                                  Progress step is checked -->

## 🎯 Goal
- {1–2 sentences naming the user-visible outcome this PR delivers; for a fix, state the symptom it removes (root cause goes in the section below)}

## 🔍 Problem                                <!-- bug fixes only -->
{one-paragraph summary of the issue}

## 🔍 Root Cause                             <!-- bug fixes only -->
{why the bug occurred — from the om-fix summary}

## External References                       <!-- only if --skill-url was used -->
- {url — what was adopted, what was rejected}

## What Changed
- {one bullet per changed area — name the files/modules touched and the behavioral change; write for a reviewer who has not opened the diff, not a one-line restatement of the title}

## 🧪 Tests
- {commands run with result counts, e.g. `npm test — 2927 passed / 172 files`; note any skipped or failing gate}
- {the new/updated test cases and what behavior they lock in}

## 💥 Breaking Changes
- {None | describe affected contracts and migration notes}

## 📋 Progress                               <!-- conditional: --plan -->
See the Progress section in the tracking plan.
```
