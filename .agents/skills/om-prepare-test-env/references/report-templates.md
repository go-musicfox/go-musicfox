# Report templates — run summary (step 1) and generation report (step 2.6)

How `om-prepare-test-env` reports back to the user. Reporting style contract:
`references/rules.md` (Reporting style) — full sentences, explain the why,
never compress; emojis structure the sections, the text carries the meaning.
This skill defines no chaining reference lines. The same shape serves both the
fast path (the saved entrypoint ran) and a generation run — a generation run
additionally fills the scripts and timings lines from the cold/warm
verification.

## Run report

```markdown
## 🧪 om-prepare-test-env — {✅ environment up | ✅ scripts generated and verified | ❌ blocked}

**Result:** {one full sentence: the environment was reused / rebuilt / freshly generated, and what that means for the caller}
**Base URL:** {baseUrl from the descriptor} — {one sentence: what answers there and which readiness probe confirmed it}
**Descriptor:** `{$ENV_DESCRIPTOR path}` — {one sentence: this is what QA and integration-test skills attach to; no real secrets inside}
**Scripts:** `{$UP_SCRIPT}` / `{$DOWN_SCRIPT}` ({sh | ps1 flavor}) — {one sentence: run the up script to relaunch the same environment on this platform}
**Timings:** {this run took Xs ({reused | rebuilt}); generation runs: cold Xs, warm Ys — the warm run reused instead of rebuilding}

### 📋 Services
{One bullet per service in a full sentence: what it is, the port it is bound to on 127.0.0.1, and whether it was started fresh or reused.}

### ⚠️ Repairs and lessons
{Only when this run repaired or improved the scripts: full sentences describing the failure or drift observed, the fix baked into the script (with its dated history line), the repo-local skill note added, and the recommendation to commit the updated scripts. Omit this section on a clean fast-path run.}
```

When the run is **blocked** (generation could not pass cold+warm verification,
or the browser provider could not be readied), replace the sections above with
a ⚠️ paragraph in full sentences: what was attempted, the exact blocker, the
fallback recorded, and what a human should do before the next attempt.

Teardown runs (`--stop` / `--down`) report in one or two full sentences: what
was stopped (only what this repo started), and that the descriptor is now
marked `"status":"stopped"`.
