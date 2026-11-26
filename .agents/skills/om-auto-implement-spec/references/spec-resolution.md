# Spec resolution (step 1)

The one procedure for turning a `{spec}` reference into a concrete spec file. Shared contract: `om-auto-create-pr --spec` and `om-auto-fix-issue`'s feature-route spec lookup follow this same order — update all consumers together.

Try in order; first hit wins:

1. **Path** — `{spec}` is an existing repo-relative file path (under `$SPECS_DIR` or elsewhere): use it directly.
2. **Name/slug in `$SPECS_DIR`** — case-insensitive match of `{spec}` against filenames (with or without the `YYYY-MM-DD-` prefix and `.md`) and against each spec's `# {Title}` line. One match → use it. Multiple matches → pick the newest by filename date **only** when its title matches unambiguously; otherwise treat as not found and list the matches as candidates.
3. **Issue id** — when `{spec}` is numeric, **get-issue**: scan the issue body and comments for links/paths to `$SPECS_DIR` files or spec-PR references. Record `ISSUE_ID` for PR linkage.
4. **Spec PR** — when `{spec}` is numeric and step 3 found nothing (or points at a PR), **get-pr**: if the PR's branch adds a file under `$SPECS_DIR`, use that file and set `SPEC_PR`. Also run **search-prs** for open PRs referencing the resolved spec path — an open spec PR means the spec is in flight (set `SPEC_PR`; the spec PR stays design-only — implementation ships on its own PR referencing it).

Validate the resolved file: it must contain an `## Implementation Plan` (or `## Phasing`) section to be implementable. When it lacks one, stop with the not-found notification variant: "spec found but has no implementation breakdown — run om-spec-writing to complete it first."

## Not-found notification

```
Status: blocked
Spec not found for "{spec}".
Searched: path, $SPECS_DIR name/title match, issue body links, spec-PR branches.
Closest candidates:
- {path} — {title}
- …
Next: pass an exact path, or create the spec first with om-auto-write-spec "{spec}".
```
