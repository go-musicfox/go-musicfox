# Release window — which PRs belong in this entry

How `om-auto-update-changelog` builds the set of PRs for step 2. A release entry describes **what the release ships**, so the window is defined by the ref the release is cut from — not by the branch PRs happen to target.

## Resolve the release ref

```bash
RELEASE_REF="${RELEASE_REF:-$BASE_BRANCH}"   # --release-ref wins; default: the configured base branch
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || true)
git fetch origin "$RELEASE_REF" --tags --quiet
```

In a single-branch repo `RELEASE_REF` equals `BASE_BRANCH` and the two enumeration modes agree. In a repo with an integration branch that runs ahead of the released branch, they do **not**: a `baseRefName` filter on the integration branch lists unshipped work *and* misses the PRs that landed on the released branch with the cut. When `RELEASE_REF` differs from `BASE_BRANCH`, or the repo has any branch that merges into the released branch, use reachability mode.

## Reachability mode (default whenever a tag and the release ref are both readable)

```bash
git rev-list "${LAST_TAG}..origin/${RELEASE_REF}" > /tmp/window-commits.txt
```

Enumerate merged PRs for the calendar window across **all** base branches — no `baseRefName` filter — requesting `mergeCommit` in the field list (**list-prs** ships it for merged PRs, but only when asked). Keep a PR when its `mergeCommit.oid` appears in `window-commits.txt`; drop it otherwise.

Start the calendar window a few days **before** the previous release date. PRs merged into an integration branch shortly before the cut land on the released branch with it, and a window that starts exactly at the tag date silently drops them. Reachability then discards anything the early start over-collected, so an early bound costs nothing.

**Degradation:** in a shallow clone, with no tags, or when the tracker returns no `mergeCommit`, fall back to filtering on `baseRefName == $RELEASE_REF` and say so explicitly in the report — that mode is known to mis-window repos with an integration branch.

## Pagination — the silent truncation

**list-prs** caps at its `limit` (250 in the shipped descriptors) and an active repo exceeds that in a release window. A truncated list produces an entry that is *plausibly* complete and quietly missing work.

Detect it: when the returned count equals the requested limit, the window is truncated. Split the window into date chunks (`merged:>=A merged:<B`), run **list-prs** per chunk, and dedupe by PR number until every chunk comes back under the cap. Report the total number of PRs enumerated so a reader can sanity-check it against the entry.

## Exclusions

Drop from the window:

- PRs that touched only `${RUNS_DIR}/` — execution-plan commits, not release work.
- PRs whose entire body says `Update CHANGELOG.md for vX.Y.Z` — prior runs of this skill.
- Branch-sync and release plumbing — titles matching `^chore:\s*(sync|merge|prepare)\b.*\b(branch|release|back)\b` or an equivalent shape in this repo's history. Check what previous entries omitted and match that.
- Any PR number that already appears in an earlier `CHANGELOG.md` entry — a re-run, or a fix carried to two branches, must not be listed twice.

Do **not** auto-exclude sub-PRs that merged into an intermediate feature branch: they are real work, and the umbrella-merge rule (Path D in `references/supersede-credit-rule.md`) decides whether they get their own bullet or coalesce into the umbrella's.
