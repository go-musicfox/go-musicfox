# Run folder layout

Every run lives in its own folder (never a flat file). Verification is **checkpoint-based** — one combined `checkpoint-<N>-checks.md` for every ~5 Steps, not per Step. Per-Step verification logs are NOT produced; the per-Step commit flips its own row in the Tasks table and nothing else.

```
${RUNS_DIR}/<YYYY-MM-DD>-<slug>/
├── PLAN.md                       # Tasks table (top), goal, scope, phases/steps (1:1 step↔commit)
├── HANDOFF.md                    # Rewritten at each checkpoint and at run end (not per Step)
├── NOTIFY.md                     # Append-only UTC log — checkpoint events, blockers, decisions only
├── checkpoint-<N>-checks.md      # Required every ~5 Steps — cumulative verification log
├── checkpoint-<N>-artifacts/     # Optional — screenshots + browser-automation transcripts from this checkpoint
│   ├── browser-session.log
│   ├── screenshot-<desc>.png
│   └── typecheck.log
├── final-gate-checks.md          # Written at spec completion — full gate + integration suite + style pass
├── final-gate-artifacts/         # Optional — retained only when raw output is worth keeping
└── ...
```

Rules:

- `<X.Y>` is the exact Step id from the `Step` column of `PLAN.md`'s `## Tasks` table.
- `<N>` is a monotonically increasing checkpoint index starting at `1`. A checkpoint fires after every 5 consecutive Steps and again at spec completion (as part of the final gate).
- **There is NO `step-<X.Y>-checks.md` and NO `step-<X.Y>-artifacts/`.** Do not create them. Per-Step chatter (individual check logs, NOTIFY entries, HANDOFF rewrites) is deliberately dropped to reduce noise.
- `checkpoint-<N>-artifacts/` is optional — create it only when the checkpoint produced real artifacts (browser transcripts, screenshots, captured command output). Never create an empty folder.

This section is the layout contract; `om-auto-continue-pr-loop` parses these files to resume a run.

## Commit the run folder as the first commit (step 7)

```bash
mkdir -p "$RUN_DIR"
git add "$RUN_DIR"
git commit -m "docs(runs): add execution plan for ${SLUG}"
git push -u origin "$BRANCH"
```

Do not pre-create `checkpoint-*-checks.md` or `checkpoint-*-artifacts/` — each checkpoint writes its own files when it fires. This guarantees that if anything later crashes, `om-auto-continue-pr-loop` can find `PLAN.md`, `HANDOFF.md`, and `NOTIFY.md` via the remote branch.
