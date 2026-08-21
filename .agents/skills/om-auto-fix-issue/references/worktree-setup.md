# Worktree setup — isolated worktree and fix branch

Detailed procedure for step 5 (create) and step 12 (cleanup) of `om-auto-fix-issue`. Never run in the user's primary worktree.

## Create the worktree and fix branch (step 5)

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON_DIR=$(git rev-parse --git-common-dir)
WORKTREE_PARENT="$REPO_ROOT/.ai/tmp/om-auto-fix-issue"
CREATED_WORKTREE=0

if [ "$GIT_DIR" != "$GIT_COMMON_DIR" ]; then
  WORKTREE_DIR="$PWD"
else
  WORKTREE_DIR="$WORKTREE_PARENT/issue-{issueId}-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$WORKTREE_PARENT"
  git fetch origin "$BASE_BRANCH"
  git worktree add --detach "$WORKTREE_DIR" "origin/$BASE_BRANCH"
  CREATED_WORKTREE=1
fi

cd "$WORKTREE_DIR"
BRANCH_PREFIX="fix"
# Switch to feat only when the issue is clearly an enhancement or new capability,
# not a corrective change to existing behavior.
git checkout -B "${BRANCH_PREFIX}/issue-{issueId}-{slug}" "origin/$BASE_BRANCH"
```

Then install dependencies with whatever the repository's lockfile implies (npm, pnpm, bun, cargo, etc.); skip when the project needs no install step.

Rules:

- Reuse the current linked worktree when already inside one. Never nest worktrees.
- The main worktree must stay untouched.
- Always clean up the temporary worktree at the end, but only if you created it this run.

## Cleanup sequence (steps 5 and 11)

Run in a `trap`/finally so crashes also clean up:

```bash
cd "$REPO_ROOT"
if [ "$CREATED_WORKTREE" = "1" ]; then
  git worktree remove --force "$WORKTREE_DIR"
fi
git worktree prune
```

## om-auto-fix-issue specifics

- Sanitize every interpolated value before substituting it into the commands above: `{issueId}` must be purely numeric, and `{slug}` is one you generate yourself from the issue title — lowercase it, replace everything outside `[a-z0-9]` with `-`, and cap it at ~40 characters. Never substitute raw tracker-provided text into a shell command, branch name, or path.
- Branches use `fix/issue-{issueId}-{slug}` for corrective work or `feat/issue-{issueId}-{slug}` for enhancements.
- This procedure applies to the bug route only; on the feature route the delegated skills (`om-auto-write-spec`, `om-auto-implement-spec`) own their own worktrees.
