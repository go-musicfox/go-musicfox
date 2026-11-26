# Duplicated / already-merged change detection

Detailed procedure for step 6 of `om-auto-review-pr`. Before proceeding with the full review, verify that the PR does not duplicate work already present in the base branch. This catches: the base branch already contains the same fix (e.g., merged via a different PR); a parallel PR landed the same feature while this one was open; the PR's changes are a subset of recently merged work.

Steps:

1. Get the list of changed files from the PR diff: run the tracker operation **get-pr-diff** for `{prNumber}` in changed-file-list mode (names only, no patch content).

2. For each changed file, compare the PR version against the base branch version to identify overlap:
   ```bash
   git diff origin/{baseRefName} -- <file>
   ```

3. Check recent commits on the base branch that touch the same files:
   ```bash
   git log origin/{baseRefName} --oneline -20 -- <files>
   ```

4. Look for semantic duplication — the same logic, function, or fix already present in the base branch even if the code differs slightly.

If the PR's core changes are already present in the base branch:
- Submit a changes-requested review explaining that the changes duplicate already-merged work.
- List the specific commits or PRs in the base branch that already contain the equivalent changes.
- Set the pipeline label to `changes-requested` (which also removes `merge-queue`) and stop.

If partial overlap exists (some changes are new, some are redundant):
- Note the redundant parts as a finding in the review.
- Continue reviewing the genuinely new changes.
