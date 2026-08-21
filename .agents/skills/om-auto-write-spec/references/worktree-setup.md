# Worktree setup — isolated worktree and spec branch

Detailed procedure for step 2 (create) and step 8 (cleanup) of `om-auto-write-spec`. Never run in the user's primary worktree.

## Create the worktree and task branch (step 2)

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON_DIR=$(git rev-parse --git-common-dir)
WORKTREE_PARENT="$REPO_ROOT/.ai/tmp/om-auto-write-spec"
CREATED_WORKTREE=0
BRANCH="spec/${SLUG}"

if [ "$GIT_DIR" != "$GIT_COMMON_DIR" ]; then
  WORKTREE_DIR="$PWD"
else
  WORKTREE_DIR="$WORKTREE_PARENT/${SLUG}-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$WORKTREE_PARENT"
  git fetch origin "$BASE_BRANCH"
  git worktree add --detach "$WORKTREE_DIR" "origin/$BASE_BRANCH"
  CREATED_WORKTREE=1
fi

cd "$WORKTREE_DIR"
git checkout -B "$BRANCH" "origin/$BASE_BRANCH"
```

Then install dependencies with whatever the repository's lockfile implies (npm, pnpm, bun, cargo, etc.); skip when the project needs no install step.

Rules:

- Reuse the current linked worktree when already inside one. Never nest worktrees.
- The main worktree must stay untouched.
- Always clean up the temporary worktree at the end, but only if you created it this run.

## Cleanup sequence (steps 2 and 8)

Run in a `trap`/finally so crashes also clean up:

```bash
cd "$REPO_ROOT"
if [ "$CREATED_WORKTREE" = "1" ]; then
  git worktree remove --force "$WORKTREE_DIR"
fi
git worktree prune
```

## om-auto-write-spec specifics

- The task branch is always `spec/${SLUG}`, detached from `origin/$BASE_BRANCH` — never `feat/` or `fix/`; the implementation skill that follows owns those.
- A spec run is docs-only: the dependency-install step is usually unnecessary. Install only when step 5's mockup/screenshot capture needs the app booted, and even then prefer the shared test environment the `om-prepare-test-env` descriptor provides over provisioning inside this worktree.
