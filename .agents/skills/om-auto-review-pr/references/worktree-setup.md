# Worktree setup — isolated worktree for the PR

Detailed procedure for step 5 (create) and step 12 (cleanup) of `om-auto-review-pr`. Never review directly in the repository's primary worktree.

## Create the worktree (step 5)

First detect whether you are already inside a linked worktree:

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON_DIR=$(git rev-parse --git-common-dir)
WORKTREE_PARENT="$REPO_ROOT/.ai/tmp/om-auto-review-pr"
CREATED_WORKTREE=0

if [ "$GIT_DIR" != "$GIT_COMMON_DIR" ]; then
  WORKTREE_DIR="$PWD"
else
  WORKTREE_DIR="$WORKTREE_PARENT/pr-{prNumber}-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$WORKTREE_PARENT"
  git fetch origin "pull/{prNumber}/head"
  PR_HEAD_SHA=$(git rev-parse FETCH_HEAD)
  git worktree add --detach "$WORKTREE_DIR" "$PR_HEAD_SHA"
  CREATED_WORKTREE=1

  cd "$WORKTREE_DIR"
  git switch -c "review/pr-{prNumber}"
fi
```

Then install dependencies with whatever the repository's lockfile implies (npm, pnpm, bun, cargo, etc.); skip when the project needs no install step. Always restore this install state before running build, test, or other validation commands in the worktree.

Rules:

- Reuse the current linked worktree when already inside one. Never nest worktrees.
- The main worktree must stay untouched.
- Review, testing, and any optional follow-up fixes must happen inside the isolated worktree.
- Always clean up the temporary worktree at the end, even on failure, but only if you created it this run.

## Cleanup sequence (step 12)

Run in a `trap`/finally so crashes also clean up:

```bash
cd "$REPO_ROOT"
if [ "$CREATED_WORKTREE" = "1" ]; then
  git worktree remove --force "$WORKTREE_DIR"
fi
git worktree prune
```

## om-auto-review-pr specifics

- The checkout target is the **PR head**, not a fresh task branch off the base: use the code host's PR head ref (`pull/{prNumber}/head`, as fetched above) so the checkout works for both same-repo PRs and fork PRs; if that ref cannot be fetched from `origin`, fall back to the tracker operation **checkout-pr** for `{prNumber}`.
- If you reused an existing linked worktree, repoint it deliberately to the PR branch or a fresh local branch for that PR before continuing. After selecting the worktree, ensure you are on the correct PR branch context:

```bash
cd "$WORKTREE_DIR"
git fetch origin "pull/{prNumber}/head"
git checkout -B "review/pr-{prNumber}" FETCH_HEAD
git fetch origin "{baseRefName}"
```
