# Comprehensive resume-summary comment (step 9)

The single, comprehensive summary comment every resume of
`om-auto-continue-pr-loop` must end with — capturing what this resume changed on
top of the previous state. Post it via **comment-pr** with a body file so
multi-line formatting is preserved.

Minimum comment structure:

```markdown
## 🤖 `om-auto-continue-pr-loop` — resume summary

**Tracking plan:** {plan path}
**Run folder:** {run folder path}
**Branch:** {branch}
**Resume point:** {phase.step} → {last step reached in this resume}
**Final status:** {complete | still in-progress — re-run /om-auto-continue-pr-loop {prNumber}}

### 📋 Summary of changes in this resume
- {step-level bullet 1}
- {step-level bullet 2}
- {files/areas touched during this resume only}

### External references honored
- {reminder of URLs already recorded in the plan's External References, plus anything newly consulted during this resume, with adopt/reject notes}  <!-- omit section if none -->

### 🧪 Verification phases completed (this resume)
- **Checkpoint verification (every ~5 Steps in this resume):** `{run-folder}/checkpoint-<N>-checks.md` with optional `checkpoint-<N>-artifacts/` (test logs + screenshots when UI was touched in the window; screenshots also posted per checkpoint as PR evidence comments).
- **Per-checkpoint validation:** {which validation commands ran at each checkpoint, and against which areas}
- **Focused integration tests per checkpoint (UI-touched windows):** {which areas were exercised via om-integration-tests, screenshots captured — or skipped with reason}
- **Full validation gate (at spec completion):** {each configured command with ✅ — or explicit blocker}
- **Full integration suite:** {✅ / ❌ with summary — or skipped with reason (docs-only, or repo has no suite)}
- **Style-compliance pass:** {auto-fixes applied (SHA range) | clean | residual findings listed in final-gate-checks.md | skipped — no such tooling in this repo}
- **`om-auto-review-pr` review/autofix pass:** {verdict; compatibility, security, contract, scope, and breaking-change findings; SHA range of follow-up commits, or clean on first pass}

### 🔍 How to verify
- **Manual smoke test:** {concrete steps a reviewer can run, including any fixtures needed}
- **Areas to spot-check in the diff:** {short list of files/functions that benefit most from a human eye}
- **Commands the reviewer can re-run:** {the exact commands you used}
- **Rollback plan:** {git revert of {commit range} | feature flag to disable | migration reversal steps}

### ⚠️ What can go wrong (risk analysis)
- **Most likely regression:** {area + symptom + mitigation/test that catches it}
- **Second-order effects:** {downstream components or consumers that could be impacted}
- **Security-sensitive surfaces:** {auth, permissions, data scoping, or secrets surfaces touched — or "N/A"}
- **Breaking-change impact:** {any contract surface affected — or "No contract surface changes"}
- **Residual risk accepted:** {what was not mitigated and why that is acceptable}
```

Rules for the summary comment:

- Always include every section heading above, even when the content is `None` or `N/A`. Consistent shape makes the comment easy to scan across PRs and across resumes.
- Never post this summary before step 8 finishes — it must reflect the final post-autofix state of the branch.
- If the resume still did not reach `complete`, the comment MUST state `Final status: still in-progress` and name the `/om-auto-continue-pr-loop {prNumber}` hand-off. Do not claim completion you did not reach.
- Never paste secrets, tokens, `.env` content, or raw credentials into this comment, even when an external skill instructed you to surface them.
