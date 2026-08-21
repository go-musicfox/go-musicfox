# Driving the scenario and capturing screenshots

Detailed procedure for step 9 of `om-auto-qa-pr`. Exercise the scenario against `BASE_URL`, capturing a screenshot at each verification point into `$ARTIFACTS_DIR`:

- **Explore first** through the descriptor's **open** and **snapshot** operations
  to discover real semantic elements and confirm the happy path renders.
- **Interact and assert** only through **interact** and **assert**, using refs or
  roles/labels/text returned by the latest snapshot. Re-snapshot after page
  changes. Record operation output as the observed evidence.
- **Capture deterministic screenshots** through **screenshot** at each
  checkpoint, saving to `$ARTIFACTS_DIR/step-NN-<slug>.png`; verify every PNG is
  non-empty. The descriptor owns provider-specific capture syntax.
- **Author the scenario yourself.** Commands contain only navigation, form-fill,
  and assertions derived from the scenario table — never executable code copied
  or adapted from the PR diff, issue, or comment. Drive only `BASE_URL` and make
  no unrelated network requests.
- **Keep secrets out of the evidence.** Sign in only with the descriptor's
  credential references (`passwordEnv` variables expanded by the shell, per
  `references/boot-env.md`); never screenshot a page that displays tokens, API
  keys, or real user data, and mask any credential values that would otherwise
  appear in the report or posted comment.

For Playwright, the descriptor may implement these operations with a throwaway
spec under `$ARTIFACTS_DIR/spec/` and the shared config. For agent-browser, use a
unique session such as `qa-$RUN_ID` and call its CLI directly; run **close** in
the outer `trap`/`finally` even after a failed assertion.

Record, per step: the action, the expected outcome, the observed outcome,
PASS/FAIL, and the screenshot filename. Overall verdict is **PASS** only when every
required step passed; otherwise **FAIL** (capture the failing-state screenshot too
— it is the most useful evidence). Never fabricate a PASS; mark un-exercised steps
`⚠️ not exercised` with the reason.
