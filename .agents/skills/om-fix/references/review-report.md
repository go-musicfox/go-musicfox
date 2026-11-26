# Review hand-off — authoritative review runs after PR creation

`om-fix` validates and hands the change to the PR-opening step. The chain performs its single authoritative code-review pass later through `om-auto-review-pr`.

## om-fix specifics

The later `om-auto-review-pr` pass must apply these checks to a fix produced from an analyzer brief:

- No API response fields removed.
- No data-scoping or permission-check rules weakened; the project's data-access conventions followed in every changed production file.
- Fix remains minimal — edit only what the analyzer named plus tests; refactors belong in their own PR.

Do not run `om-code-review` directly in `om-fix`; `om-auto-review-pr` invokes it once after `om-open-pr`, driven by `om-auto-fix-issue` or the external flow runner.
