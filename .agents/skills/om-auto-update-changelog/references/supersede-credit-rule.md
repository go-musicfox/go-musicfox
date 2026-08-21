# Supersede Credit Rule

The credited-author resolution algorithm `om-auto-update-changelog` runs in
step 4 to compute `primaryAuthor` and optional `viaAuthor` for every merged PR,
plus the verification pass that must clear every credit before the entry is
assembled.

The problem, in two shapes. **Carry-forward:** when `om-auto-review-pr` carries a fork contributor's PR forward, the **merged** PR's author field is the reviewer, not the original contributor — who did the work and must get the credit (Paths A–C). **Merge-time capture:** an umbrella PR that merges a long-lived feature branch, or an informal hand-off written in prose, records whoever pressed the merge button as the author of work they did not write (Paths D–E). A wrong credit is the most damaging error a changelog can carry, because it is published and attributed.

Five detection paths. When several fire, the earliest-lettered one wins (A > B > C > D > E). The verification pass runs after all of them and overrides any of them.

## Identities that are never credited

Exclude these from commit tallies, from `primaryAuthor` / `viaAuthor`, and from the Contributors block — handles matching any of:

```
\[bot\]$        dependabot        renovate        ^app/        ^web-flow$
```

…plus AI coding agents, which commit under their own identities and must never appear as a contributor: `claude`, `cursoragent`, `copilot`, `codex`, `devin` (match the whole handle, case-insensitively; a human handle that merely *contains* one of these words is a real contributor).

When a PR's credit resolves to nothing but excluded identities, render the bullet with **no** `*(@...)*` suffix at all — never credit the bot, and never fall back to the merger. When carried-forward work *originates* from a bot (a dependency bump reopened or rebased by a maintainer), the maintainer keeps the credit: there is no human original author to restore.

## Path A: `Supersedes #N` in the PR body

`om-auto-review-pr` writes this template when it carries a fork PR forward. Regex (anchored to the first 20 lines of the body, case-insensitive):

```
^Supersedes\s+#(\d+)\b
```

When matched, resolve the superseded PR's author via the tracker operation **get-pr** (field `author`) for `{supersededPrNumber}`. Set `primaryAuthor` to that author and `viaAuthor = mergedPrAuthor`. Emit `(supersedes #M)` in the summary text.

## Path B: `Credit: original implementation by @user` in the PR body

Same template, also written by `om-auto-review-pr`. Regex:

```
Credit:\s+original\s+implementation\s+by\s+@([A-Za-z0-9][A-Za-z0-9-]{0,38})
```

When matched, set `primaryAuthor` from the captured handle and `viaAuthor = mergedPrAuthor`. No `supersedes #M` suffix unless Path A also fires (it usually does).

## Path C: `Closing in favor of` comment on the superseded PR

When neither body regex on the merged PR matches, the carry-forward flow still leaves an authoritative trail on the **original** PR via `om-auto-review-pr`'s closing-comment template:

```
Closing in favor of #{newPrNumber} ({newPrUrl}).

Credit to @{originalAuthor} for the original implementation. ...
```

Detection is reversed compared to Paths A and B — you are walking *candidate superseded PRs*, not the merged PR itself. For each closed-unmerged PR in the same window (**list-prs**, state closed, `closed:>=${SINCE_DATE} is:unmerged`), check its body **and** its comments for a line matching:

```
^Closing in favor of #(\d+)\b
```

When the captured number equals the merged PR currently being credited, treat the merged PR as a carry-forward. Set `primaryAuthor` to the closed PR's author (via **get-pr**) and `viaAuthor` to the merged replacement's author.

Run Path C as a real sweep of the window's closed-unmerged PRs, not as a spot check — and when a detected replacement is *absent* from the entry (it merged after the release was cut), confirm that rather than assuming it.

## Path D: umbrella / feature-branch merge

A PR that merges a long-lived feature branch is authored by whoever pressed the button, not by whoever wrote the code. Detect it from commit authorship — request `commits` alongside `author` in the **get-pr** field list:

> tracker operation **get-pr** — `{prNumber}`, fields `number,title,author,commits,mergeCommit`

Tally commit authors, excluding the never-credited identities above. Path D fires when the PR author authored **zero** commits **and** a single other human authored a clear majority of the remaining commits. Then:

- `primaryAuthor` = that human.
- `viaAuthor` = **null**. A merge is not a carry-forward; do not disclose the merger as a `via`.

An umbrella PR usually ships alongside the sub-PRs that built the feature branch, and those sub-PRs land through the same merge commit, describing the same work. Detect them by checking whether another windowed PR's merge commit appears in the umbrella's `commits` list, and **coalesce the umbrella with its sub-PRs into one bullet** listing every number — `(#1820, #1655)` — instead of emitting the work twice.

> The failure this path exists to prevent: a 192-commit "implementation of X" PR credited to the maintainer who merged the feature branch. Zero of those commits were theirs — a contributor wrote 100 and an AI agent wrote 92 — and a sub-PR listed the same work a second time.

## Path E: free-text attribution in the PR body

Contributors get handed off in prose that matches none of the `om-auto-review-pr` templates. Scan every merged PR body (case-insensitive) for:

```
Original author:? .*@([A-Za-z0-9][A-Za-z0-9-]{0,38})
Carries (the )?.* from #(\d+)
[Cc]redits? (to|goes to) @([A-Za-z0-9][A-Za-z0-9-]{0,38})
(takes?|took) over .*@([A-Za-z0-9][A-Za-z0-9-]{0,38})
[Bb]ased on (the )?work (of|by) @([A-Za-z0-9][A-Za-z0-9-]{0,38})
```

A captured handle becomes `primaryAuthor` with the merged PR author as `viaAuthor`. A captured PR number resolves its author via **get-pr** and additionally emits `(supersedes #N)` in the line text.

> Real failures this path caught: `Original author: … (@handle) — assigning for review/ownership` and `Carries the registry from #N (original commit preserved)`, both credited to the maintainer who opened the follow-up.

## Fallback

If none of A–E match, `primaryAuthor = mergedPrAuthor` and `viaAuthor = null` — no supersede. Most PRs fall here, and the verification pass below is what keeps that default honest.

## Mandatory verification pass

Run this after resolving every credit and **before** assembling the entry. It is not a spot check: compare each bullet's `primaryAuthor` against that PR's commit authorship (already fetched for Path D) and review every mismatch by hand.

| Signal | Verdict |
|--------|---------|
| Credited author wrote **0** commits **and** a `Credit:` / `Supersedes` template is present | ✅ correct — the original commits live on the closed PR; the template is authoritative over commit counts |
| Credited author wrote **0** commits **and** no template | ❌ **wrong** — apply Path D or E, or investigate before shipping |
| Credited author wrote a minority of commits, the rest by a maintainer | ✅ usually correct — normal review-fixup or rebase-and-fix carry |
| …and the maintainer's majority commits are titled "address review findings" or similar | ✅ correct — a review fix is not authorship |
| PR author differs from the dominant commit author on a large (50+ commit) PR | ❌ investigate as an umbrella merge (Path D) |
| Every commit author is an excluded identity | ✅ render the bullet with no author suffix |

Report the pass in the run report: how many credits were checked, how many mismatches were reviewed, and what each mismatch resolved to. When a mismatch cannot be resolved confidently, keep the fallback credit and leave an `<!-- credit unverified for #N -->` HTML comment above the entry.

**Residual gap, stated so nobody assumes coverage:** a hand-off with no template, no attribution wording, *and* commits re-authored under the merging maintainer's name is invisible to all five paths and to the verification pass. Only a human who knows the work can catch it.

## Worked example

Given merged PR `#1555` with body:

```markdown
Supersedes #1421

Credit: original implementation by @contributor-a. This follow-up PR carries that work forward with the requested fixes so it can merge without waiting on the original branch.
```

...and PR author `@reviewer-b` (the reviewer), the changelog entry becomes:

```markdown
- 🐛 Validate event names against module registry (supersedes #1421). (#1555) *(@contributor-a, via @reviewer-b)*
```

The Contributors block lists `@contributor-a` first (primary author) and `@reviewer-b` once (only if they did not already appear as a primary author for some other PR in the same release).

Contrast with Path D — merged PR `#1820` "implementation of the warehouse module", authored by `@maintainer-c`, whose 192 commits are 100 by `@contributor-d`, 92 by an AI agent, and none by `@maintainer-c`; sub-PR `#1655` is in the same window and its merge commit is inside `#1820`:

```markdown
- ✨ Implementation of the warehouse module. (#1820, #1655) *(@contributor-d)*
```

No `via @maintainer-c` — pressing merge is not a contribution.
