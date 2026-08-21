# Worktree setup — isolated worktree and PR checkout

Detailed procedure for step 2 (create) and step 6 (cleanup) of `om-auto-fix-pr`. Never run in the user's primary worktree.

## Create the worktree (step 2)

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
GIT_DIR=$(git rev-parse --git-dir)
GIT_COMMON_DIR=$(git rev-parse --git-common-dir)
WORKTREE_PARENT="$REPO_ROOT/.ai/tmp/om-auto-fix-pr"
CREATED_WORKTREE=0

if [ "$GIT_DIR" != "$GIT_COMMON_DIR" ]; then
  WORKTREE_DIR="$PWD"
else
  WORKTREE_DIR="$WORKTREE_PARENT/pr-${PR_NUMBER}-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$WORKTREE_PARENT"
  git fetch origin "$BASE_BRANCH"
  git worktree add --detach "$WORKTREE_DIR" "origin/$BASE_BRANCH"
  CREATED_WORKTREE=1
fi

cd "$WORKTREE_DIR"
```

Then install dependencies with whatever the repository's lockfile implies (npm, pnpm, bun, cargo, etc.); skip when the project needs no install step.

Rules:

- Reuse the current linked worktree when already inside one. Never nest worktrees.
- The main worktree must stay untouched.
- Always clean up the temporary worktree at the end, but only if you created it this run.

## Cleanup sequence (steps 2 and 6)

Run in a `trap`/finally so crashes also clean up:

```bash
cd "$REPO_ROOT"
if [ "$CREATED_WORKTREE" = "1" ]; then
  git worktree remove --force "$WORKTREE_DIR"
fi
git worktree prune
```

## om-auto-fix-pr specifics

- **Check out the PR head, not a fresh branch.** Instead of creating a task branch off the base, check out the PR head inside the worktree via the tracker operation **checkout-pr** for `{prNumber}`; the whole run continues on the PR's real head branch. (In `--ci-only --branch <name>` mode there is no PR: `git fetch origin "<name>" && git checkout -B "<name>" "origin/<name>"` instead.)
- **One worktree for the whole loop.** The same worktree serves every outer-loop cycle; sub-skills invoked from inside it reuse it per the never-nest rule. Clean up only what this run created.
- If the loop switches to a fork carry-forward replacement PR (see `references/base-merge.md`), the carry branch created by `om-auto-review-pr` becomes the checked-out branch for the remaining cycles.
