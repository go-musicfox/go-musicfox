# Worktree setup — isolated worktree from the PR head

Detailed procedure for step 4 (create) and step 11 (cleanup) of `om-auto-continue-pr-loop`. Never resume in the user's primary worktree.

## Create the worktree from the PR head (step 4)

`HEAD_REF` and `IS_CROSS` are filled via **get-pr** (fields `headRefName`, `isCrossRepository` — already part of the step 1 fetch). On the cross-repository path, use the **checkout-pr** operation to make the PR head available locally.

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON_DIR=$(git rev-parse --git-common-dir)
WORKTREE_PARENT="$REPO_ROOT/.ai/tmp/om-auto-continue-pr-loop"
CREATED_WORKTREE=0

# tracker: get-pr → HEAD_REF (headRefName), IS_CROSS (isCrossRepository)

if [ "$GIT_DIR" != "$GIT_COMMON_DIR" ]; then
  WORKTREE_DIR="$PWD"
else
  WORKTREE_DIR="$WORKTREE_PARENT/pr-{prNumber}-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$WORKTREE_PARENT"
  if [ "$IS_CROSS" = "true" ]; then
    # tracker: checkout-pr {prNumber}
    git worktree add --detach "$WORKTREE_DIR" "HEAD"
  else
    git fetch origin "$HEAD_REF"
    git worktree add "$WORKTREE_DIR" "origin/$HEAD_REF"
  fi
  CREATED_WORKTREE=1
fi

cd "$WORKTREE_DIR"
```

Then install dependencies with whatever the repository's lockfile implies (npm, pnpm, bun, cargo, etc.); skip when the project needs no install step.

Rules:

- Reuse the current linked worktree when already inside one. Never nest worktrees.
- The main worktree must stay untouched.
- Always clean up the temporary worktree at the end, but only if you created it this run.

## Cleanup sequence (steps 4 and 11)

Run in a `trap`/finally so crashes also clean up:

```bash
cd "$REPO_ROOT"
if [ "$CREATED_WORKTREE" = "1" ]; then
  git worktree remove --force "$WORKTREE_DIR"
fi
git worktree prune
```

## om-auto-continue-pr-loop specifics

- This skill resumes an existing PR, so the worktree is created from the **PR head** (`HEAD_REF` / `IS_CROSS` from the step 1 **get-pr**) rather than from `origin/$BASE_BRANCH`, and no new task branch is cut — you continue committing on the PR's own branch.
