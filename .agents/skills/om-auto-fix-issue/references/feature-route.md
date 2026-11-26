# Feature route — spec, then build on one implementation PR

The route `om-auto-fix-issue` takes (instead of the bug chain, steps 4–12) when
step 2 classifies the issue as a feature request. It specs-then-builds the feature —
the spec (when one is authored) lands on its own design-only spec PR, and the
implementation ships on exactly one implementation PR referencing it — autonomous
by default. The delegated skills own the worktree, the claim, the review autofix
loop, and UI verification.

## F1. FR triage gate

Run the read-only FR triage per `references/fr-triage.md`: confirm the feature is
**not already implemented** (code search) and not already specced + in flight.
Already built / already in flight → stop with `NO_ACTION_NEEDED` and cited evidence.
Nothing is claimed yet, so a stop leaves no lock behind.

## F2. Claim / resume decision

The step-1 three-signal in-progress lock decision applies, with its 60-minute
stale-lock recovery and `--force` override rules. Also treat the slot as taken
when an open PR already references this issue
(**search-prs** for `#{issueId}`): stop and point at
`om-auto-continue-pr {prNumber}` — **unless** that PR is a **spec-only design PR**
(draft, `Refs #{issueId}`, a spec file but no implementation commits), which is this
route's resume point: continue at **F3b** with it as `SPEC_PR`. The delegated skills
perform their own claims, so a stopped run leaves no stray lock.

## F3. Resolve the spec and implement

a. **Resolve the spec** — follow the spec-resolution procedure in
   `references/spec-resolution.md` with `{spec}` = the issue
   id (checks `$SPECS_DIR` links in the issue body, name matches on the issue title,
   and open spec-PR branches).
b. **Spec found** (a path, or the spec-only `SPEC_PR` from F2) → invoke
   **`om-auto-implement-spec {SPEC_PATH-or-SPEC_PR} [--no-ui] [--loop] [--force]`** verbatim.
   Include `--loop` **only when the user passed it to `om-auto-fix-issue`** — never
   add it on your own; without it the plain engines are the default and
   `om-auto-create-pr` escalates to the loop on its own when the drafted plan
   exceeds the configured Step threshold (`engine.loopStepThreshold`, default 20).
   The spec PR stays design-only: it implements on a
   **separate implementation PR** that references the spec (`Refs #{SPEC_PR}` +
   `Source doc:`), resuming an existing implementation PR rather than opening a
   second, runs the review autofix loop and UI verification with screenshots, and
   leaves a ready PR. Ensure the implementing PR body carries `Closes #{issueId}`
   so the merge auto-closes the issue.
c. **No spec** → invoke **`om-auto-write-spec {issueId} [--slug …] [--force]`**
   verbatim (run its spec-writing step interactively when `--interactive` was
   passed). It claims the issue, writes the spec autonomously, attaches
   mockups/screenshots, opens the spec PR (`Refs #{issueId}`), posts the assumptions
   comment, and emits the `Spec:` and `PR:` reference lines. Then chain straight into
   **`om-auto-implement-spec {SPEC_PATH}`**, which opens the **implementation PR**
   (`Refs` the spec PR, `Closes #{issueId}`) — the spec PR stays design-only. For a
   spec **without** implementation, users run `om-auto-write-spec` directly.

## F4. Confirm the contract, report

The delegated skills own the machinery; verify the contract held:

- **Exactly one implementation PR** references the issue (a spec PR may
  additionally `Refs` it; an existing implementation PR means resume, never a
  duplicate).
- The implementation PR is **ready** (not draft) unless a `⚠ NEEDS HUMAN
  CONFIRMATION` assumptions guard applies.
- The **full label set** is present — pipeline `review`/current state, `feature`
  category, QA meta, one priority, one risk — re-run the label normalization in
  `references/pr-finalize.md` (the `om-open-pr` step-6 contract) on anything
  missing.
- **Linkage matches what ships**: `Closes #{issueId}` on the implementation PR,
  `Refs #{issueId}` on a spec-only design PR (and `Refs #{specPr}` on the
  implementation PR when a spec PR exists); the implementation PR body carries
  `Source doc:` and `Tracking plan:` so continuation can resume.
- The summary comment and, for user-facing changes, the UI screenshots are on the
  PR.

Report the route taken, spec path, branch, PR URL, and verification outcome, then
end with the chaining reference lines passed through from the delegated
skill. Then stop — do not continue to the bug chain.
