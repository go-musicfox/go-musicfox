# Worktree setup — isolated worktree at the PR head

Detailed procedure for the check-out and cleanup steps of `om-auto-qa-pr` (PR mode). Never verify in the user's primary worktree.

## Create the worktree at the PR head

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON_DIR=$(git rev-parse --git-common-dir)
WORKTREE_PARENT="$REPO_ROOT/.ai/tmp/om-auto-qa-pr"
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
fi
cd "$WORKTREE_DIR"
```

Then restore the dependency install state with whatever the repository's lockfile implies (npm, pnpm, bun, cargo, etc.); skip when the project needs no install step.

Rules:

- Reuse the current linked worktree when already inside one (repoint it deliberately to the PR head first). Never nest worktrees.
- The main worktree must stay untouched.
- Always clean up the temporary worktree at the end, but only if you created it this run — record `CREATED_WORKTREE` for that.

## Cleanup sequence

Run in a `trap`/finally so crashes also clean up:

```bash
cd "$REPO_ROOT"
if [ "$CREATED_WORKTREE" = "1" ]; then
  git worktree remove --force "$WORKTREE_DIR"
fi
git worktree prune
```

## om-auto-qa-pr specifics

- **Fork PRs:** use the code host's PR head ref (`pull/{prNumber}/head`, as fetched above) so the checkout works for both same-repo and fork PRs; if that ref cannot be fetched from `origin`, fall back to the tracker operation **checkout-pr** for `{prNumber}`.
- **Read-only:** this worktree exists only to build and run the app for verification — never edit source, never commit, never push from it.
- **Local mode:** no worktree work at all. Verify the current worktree as-is; do not stash, reset, or switch branches — the user wants their in-progress changes tested.
