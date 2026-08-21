# Reporting before CI, and the bounded CI follow-up

The rule this file exists for: **reporting is decoupled from CI verification.** The
moment the work is done, the labels go on, the review is submitted, the comments are
posted, and a draft is promoted to ready — whatever CI is doing at that instant.
Only then does the run look at CI, and only for a bounded window.

## Why — the failure this prevents

CI on a real repository runs for twenty minutes to several hours. A skill that
treats "required checks green" as a precondition for posting its verdict pays for
that twice. It burns the whole budget idling instead of finishing, and — the part
that actually hurts — if the monitoring process dies mid-wait, the pull request is
left **stranded**: still a draft, no labels, no review, not one comment recording
that any work happened at all. The analysis is gone and the next human to look at
the PR cannot tell whether anything was ever done to it.

Posting first makes the stranded case harmless. The worst outcome becomes a fully
reviewed, fully labeled PR whose CI result nobody summarized — which the tracker
shows anyway.

## What must land before any CI wait

In this order, and all of it before the run so much as reads a check status for
follow-up purposes:

1. The full label set through the `apply_label` / `set_pipeline_label` guards.
2. The review or verdict, submitted with its complete body.
3. The run's comments — label rationale, summary, handoff, evidence.
4. The draft → ready promotion, when the work is complete.
5. The claim released or handed off (below), so a dead process cannot strand the
   lock either.

## Disclosing pending CI in the review body

When a required check was still `PENDING` at the moment of the verdict, the review
body (or, for a skill that posts no review, its summary comment) MUST carry the
disclosure as its own short paragraph, so the PR is self-documenting and nobody
mistakes the approval for a green run:

```markdown
**CI is still running on this head at the time of approval.** Branch protection plus
the QA-approval gate hold the actual merge; this approval covers the code, not a
green run. Checks still pending: {names}. A follow-up comment will report the CI
outcome.
```

Adjust the first clause to the verdict (`at the time of this review` for
changes-requested). Drop the last sentence when the run will not follow up — never
promise a follow-up the run does not intend to make. When no required check was
pending, omit the paragraph entirely rather than writing a "CI was green" variant;
the review's own validation section already covers that.

## Swapping the lock to `ci-monitoring`

`in-progress` is a claim: it means an auto-skill is actively working the item and
others must back off. Once the work is done and reported, that is no longer true, so
the release step **swaps** instead of removing — `unlabel-pr` `in-progress`, then
`apply_label "ci-monitoring"` — when, and only when, the run intends to post a
CI-result follow-up. `ci-monitoring` is a meta label, not a pipeline label: it
coexists with `merge-queue`, `changes-requested` and the rest exactly as `needs-qa`
does, and it is **not** a lock — another agent or a human may act on the PR while it
is set. It says one thing: the CI-result follow-up comment is still owed.

A run that will not follow up on CI simply removes `in-progress` as before and never
applies `ci-monitoring`. Both mutations go through the descriptor's label guards, so
a repository that never created the label degrades to a logged skip.

## The bounded wait

Read `CI_MAX_WAIT_MINUTES` from `ci.maxWaitMinutes` (default `40`; `0` means do not
wait at all — report and stop, applying no `ci-monitoring`). Then poll the required
checks through **get-pr-checks** until they settle or the budget is spent.

**Checks settled inside the budget** — post the outcome as an idempotent
`` 🤖 `<skill-name>` — CI result `` comment (find the marker via
**list-issue-comments** and rewrite via **update-comment** on a re-run): which
required checks passed or failed, with links, and what it means for the verdict.
Then remove `ci-monitoring`.

- Green, verdict unchanged → say so; the PR keeps the pipeline label it already has.
- A required check went red on an **approved** PR → the verdict no longer holds:
  move the pipeline label to `changes-requested` via `set_pipeline_label`, explain
  which check failed and why in the CI-result comment, update the single
  `🏷️ label rationale` comment in place, and hand the PR back to its author.
- A red check that was already red at verdict time changes nothing — it was priced
  into the changes-requested verdict already.

**Budget exhausted with checks still running** — stop waiting. Do not keep the
process alive on a run that may take hours. Instead:

1. Run the repository's configured `validation.commands` gate in order, scoped to
   the change where the toolchain supports scoping, as the run's own completion
   evidence.
2. Post the bail-out comment under the same `` 🤖 `<skill-name>` — CI result ``
   marker, carrying: the local validation results command by command, the required
   checks that were still pending at bail-out with their links, and the explicit
   statement that **no further follow-up will come from this agent** — a human or a
   later skill run owns the CI outcome from here.
3. Remove `ci-monitoring`. Nobody is monitoring any more, and leaving it on would
   promise a follow-up that is never coming; that honesty is the whole point of the
   label.
4. Close the run out cleanly and report, rather than hanging.

## The guard that is never traded away

Bailing out of the CI wait is **not** permission to merge without CI. The local gate
substitutes for this agent's own confidence and reporting, never for branch
protection: required checks still gate the actual merge, and every merge skill still
refuses to merge until they are genuinely green. Reporting early is safe; merging
early is not. Nothing in this file relaxes the QA-approval gate, makes
`qa-approved` appliable from a diff, sets the human-only `qa` pipeline label, or
lets a pipeline label stop being mutually exclusive.
