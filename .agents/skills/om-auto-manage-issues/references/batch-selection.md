# Batch selection — resolving which issues to manage

How `om-auto-manage-issues` picks its target set in step 1 when no single
`{issueId}` was given. A specific id bypasses all of this (validate it is numeric
or a valid issue URL, then manage just that one).

## Default scope

With no id, select via the **search-issues** operation (the tracker's issue-list
command) using:

- `--state` → default `open`.
- `--limit` → default `25` (the most recent by creation).
- `--label <name>` → keep only issues carrying the label; `-<name>` keeps only
  issues **missing** it (e.g. `--label -priority-medium` to find unprioritized
  issues). Repeatable.
- `--author <login>` → keep only that author's issues.

Fetch a little more than `--limit` when ordering (below) needs it, then trim to
`--limit` after sorting.

## Ordering — worst-described first

Rank the fetched issues so the highest-value fixes run first, before the trim:

1. **Missing SDLC labels** — issues lacking a category, priority, or risk label
   rank first (they are the ones this skill most improves).
2. **Laconic** — a near-empty body, a body that is essentially just a screenshot,
   or a one-sentence body (see `references/screenshot-analysis.md` for the
   laconic test) ranks next.
3. Everything else, most-recently-created first.

An issue that is both unlabeled and laconic sorts to the very top.

## Safety

- **No id and no filter** is a request to touch the whole recent backlog. When a
  user is present, confirm the resolved scope ("about to manage the last 25 open
  issues, N of them missing labels — proceed?") before any mutation. Under
  `--dry-run`, skip the confirmation (nothing mutates). When unattended (no user
  to confirm) and no filter was given, process the default batch but state the
  scope prominently in the report.
- **Truncation is never silent.** When more issues match than `--limit`, process
  `--limit` and report how many matched but were left (`"37 open issues matched;
  processed the worst-described 25; 12 not processed — raise --limit to include
  them"`).
- Never widen `--state` to `closed`/`all` on your own — a closed issue is
  generally out of scope for enrichment unless the user explicitly asked.
