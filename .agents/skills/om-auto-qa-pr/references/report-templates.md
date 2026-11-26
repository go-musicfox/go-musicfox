# Report templates

## Verification report artifacts (step 10)

The machine- and human-readable report artifacts `om-auto-qa-pr` writes
into `$ARTIFACTS_DIR` in every mode (step 10). These are the primary deliverable
in local mode and the source of the PR comment in PR mode.

`$ARTIFACTS_DIR/report.json`:

```json
{
  "runId": "<RUN_ID>",
  "mode": "pr | local",
  "target": { "prNumber": 1234, "title": "…", "branch": "…", "headSha": "…" },
  "verdict": "PASS | FAIL | PARTIAL",
  "environment": { "baseUrl": "…", "role": "…", "startedByThisRepo": true, "browserProvider": "agent-browser | playwright | custom" },
  "scenario": [
    { "step": 1, "priority": "P1", "action": "…", "expected": "…", "observed": "…", "result": "PASS", "screenshot": "step-01-…​.png" }
  ],
  "hasUiTest": false,
  "notes": ["…"]
}
```

`$ARTIFACTS_DIR/report.md` — a readable version with the verdict, environment,
the scenario table, embedded/linked screenshots, and notes for QA. Use this
template:

```markdown
## 📸 UI QA evidence — {verdict}

**Verdict:** {✅ PASS | ❌ FAIL | ⚠️ PARTIAL — environment-limited}
**Environment:** `{baseUrl}` · role `{role}` · browser `{provider}`
**Verified:** {branch} @ {headSha (short)}

### Scenario ({P0|P1|P2} — {area})
**Where to click:** `{route}`

| # | Step | Expected | Observed | Result |
|---|------|----------|----------|--------|
| 1 | {action} | {expected} | {observed} | ✅ |

### Screenshots
![Step 1 — {slug}]({path or url})

### Notes for QA
- {edge cases not covered; permission/empty/error observations}
```

Rules: report only what was observed; never paste secrets, tokens, `.env`
content, or non-demo credentials; redact sensitive values that leaked into a
screenshot before including it, or omit the screenshot and say so.

## Follow-up UI-test scenario (step 12)

Posted only when `HAS_UI_TEST` is false — as a second PR comment (**comment-pr**)
in PR mode, or appended to `report.md` in local mode:

```markdown
## 🧪 Follow-up: add a UI/integration test

This change ships no browser-level test. The UI QA above was manual; lock it in
with an automated test (run `/om-integration-tests`).

**Scenario (derived from the manual run above):**
1. Setup: {fixtures to create via API — prefer the repo's integration fixtures}
2. Act: {the UI steps exercised above}
3. Assert: {the expected outcomes verified above}
4. Teardown: delete every fixture created.
```

## Final run report (step 14)

The user-facing summary printed at the end of every run. Reporting style
contract: `references/rules.md` (Reporting style) — full sentences, explain the
why, never compress; emojis structure the sections, the text carries the
meaning.

```markdown
## 📸 om-auto-qa-pr — {PR #{n}: {title} | local worktree}

**Verdict:** {✅ PASS | ❌ FAIL | ⚠️ PARTIAL (environment-limited)} — {one full sentence: what was verified and which observation drove the verdict, or what limited the run}
**Environment:** {one full sentence: the base URL and login role driven, the browser provider used, and whether this run started the environment or reused a running one}
**📸 Evidence:** {one full sentence: the PR comment URL where the screenshots render inline, or the artifacts directory path in local mode}
**🧪 Follow-up test:** {one full sentence: the follow-up UI-test scenario was posted because the change ships no browser-level test, or was skipped because one already exists}
**🏷️ Labels:** {one full sentence: unchanged under the evidence-only default, qa-approved + qa-self-verified applied via the self-QA exception, qa-failed applied via --apply-failure, or n/a in local mode — and why}
```
