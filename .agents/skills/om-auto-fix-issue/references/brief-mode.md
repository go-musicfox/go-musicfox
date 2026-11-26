# Brief mode (step 1) — no issue located

The path `om-auto-fix-issue` takes when its required argument is a free-form
problem description rather than a tracker issue reference. The chain never works
on unfiled work: the issue is filed first, then the run continues unchanged on
the resulting `{issueId}`.

## Detection

- **Issue reference** (normal flow, not brief mode): a bare number (`1234`),
  `#number`, or a tracker issue URL.
- **Brief** (this path): anything else — a sentence, an error message, a pasted
  problem description.
- A **numeric** `{issueId}` that **get-issue** cannot find is *not* brief mode:
  stop and report the bad reference instead of filing a new issue from a bare
  number — there is no brief to file from.

## Procedure

1. **Invoke the `om-prepare-issue` skill verbatim** with the description as its
   `{brief}`, passing through any user-provided images as `{images}` and the
   `{repo}` context when given. It dedupes, links or authors a covering spec when
   needed, and files one well-formed, SDLC-labeled issue — or reuses a credible
   duplicate. When `om-prepare-issue` is not installed, stop and name it as the
   missing chain skill.

2. **Run it under this chain's autonomous contract.** `om-prepare-issue` is an
   interactive skill; this chain is not. Where it would ask the user, make the
   recommended, most-reversible call yourself and document it in the comment it
   posts:
   - credible **open** duplicate → reuse it (comment the new detail) instead of
     filing a copy;
   - credible **closed** duplicate → file fresh with a link to the old issue.

3. **Consume its report.** Parse the `Issue: #<number> (link: <url>)` reference
   line and continue the run with that number as `{issueId}`. Record
   `Issue mode: {new | reused}` for the step 12 report line
   (`filed from brief — new | filed from brief — reused duplicate`).

4. **Respect an authored spec PR.** When `om-prepare-issue` took its
   feature-needs-spec path and landed a design-only spec PR, the feature route's
   spec resolution (`references/spec-resolution.md`) later picks it up as
   `SPEC_PR` — the spec PR stays design-only and implementation ships on its own
   PR referencing it; never author a duplicate spec.

After this, the run proceeds exactly as if the user had passed the filed
issue's number: the step 1 concurrency check runs against it (a claim made by
`om-auto-write-spec` under the same automation identity is re-entry), and
step 2 classifies it — `om-prepare-issue`'s category label feeds the
label-first classification.
