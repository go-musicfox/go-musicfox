# Step review — configurable review granularity (`engine.stepReview`)

```bash
STEP_REVIEW=$(jq -r '.engine.stepReview // "final"' "$CONFIG")
```

Three modes. The authoritative end-of-run `om-auto-review-pr` pass runs in **every** mode — it alone posts the PR review and drives labels. Step review is an internal quality gate: it posts nothing to the tracker.

- `final` (default) — no step review; the end-of-run pass is the only code review. Today's cost profile, unchanged.
- `checkpoint` — at every checkpoint pass, review the diff landed since the previous checkpoint (or run start).
- `per-step` — review every Step's commit range right after it lands, inline or dispatched, before the next Step starts.

## Procedure (`checkpoint` and `per-step` modes)

1. **Build the scoped diff**: the Step's commit range (`per-step`) or the range since the last checkpoint (`checkpoint`). The 1:1 Step↔commit discipline makes both ranges exact.
2. **Review it against the `om-code-review` skill's checklist and severity scale** (blocker / major / minor / nit) — the checklist only: do NOT run its mandatory full validation gate here. Scoped validation already ran in the per-Step loop; the full gate belongs to checkpoints and the final gate. The main session performs the review itself, or delegates a reviewer subagent within the cap-at-2 rule.
3. **Route findings by severity**:
   - **blocker / major** → fix now: append a new `X.Y-review-fix` Step to the Tasks table (`Exec` usually `inline`; a dispatched Step's fix MAY re-dispatch at the same tier), land it under the normal per-Step loop, then re-review the fix commit scoped to the findings. Bounded: at most **2 fix rounds per reviewed range**; blocker/major findings still open after that are a blocker — halt per the safety stops (rewrite `HANDOFF.md`, append a NOTIFY entry naming the open findings, report back).
   - **minor / nit** → record in `NOTIFY.md` (one line each) and defer to the final review. Never fix minors mid-run — they would inflate the Step count without moving the plan.
4. **Log only on findings**: append one NOTIFY entry per reviewed range that produced findings (mode, range, counts by severity, fix rounds used). A clean pass writes nothing.

## Cost

`per-step` multiplies review cost by the Step count — reserve it for high-risk work where an early defect propagating into later Steps costs more than the extra reviews. `checkpoint` is the middle ground: detection within ~5 Steps at a fraction of the cost. `final` stays the default so no run pays for a gate it did not ask for.
