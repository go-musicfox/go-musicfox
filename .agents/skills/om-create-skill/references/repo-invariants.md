# Repo invariants a generated skill must satisfy

The hard, repo-specific constraints `om-create-skill` bakes into every skill it
authors or splits. These are what make a skill "OM-shaped" rather than generic.

## The meta-constraint: never reproduce the forbidden literals

`scripts/lint.sh` greps the whole `skills/` tree (recursively, including
`references/`) for forbidden tokens. A skill that *documents* those tokens would
match its own rule and fail the lint. Therefore this skill — and every skill it
generates — describes the constraints **abstractly** and points to
`scripts/lint.sh` as the source of truth, instead of spelling the literals. This
also keeps generated skills product-agnostic, which is the point of the rule.

## The lint gate (authoritative: `scripts/lint.sh`)

Read the script for the exact patterns; the categories are:

- **Frontmatter**: every `skills/<name>/SKILL.md` starts with a `---` frontmatter
  block declaring `name` (equal to the directory name) and a non-empty
  `description`.
- **Product-agnostic content** (behavior, not the `om-` name prefix, which is
  allowed): no upstream product/monorepo name token, no scoped upstream package
  name, no hard-coded base-branch name, no specific alternative package-manager
  keyword, no app-specific decryption-helper name. The base branch always comes
  from config (`baseBranch`), never hard-coded.
- **Tracker abstraction**: no direct tracker-CLI command inside a skill. All
  tracker interaction is expressed as **named operations** resolved through the
  tracker descriptor. The one place raw CLI commands belong is the shipped
  descriptors under `references/trackers/` — never in a skill body.

When authoring, prefer wording that never needs the forbidden literals. When
splitting, the moved text already passed lint on the source, so preserving it 1:1
keeps it clean.

## Tracker-operation vocabulary

Skills name operations; the descriptor at `.ai/trackers/<tracker>.md` maps them
to concrete commands. Common operations seen across the pipeline:
`current-user`, `get-pr`, `get-pr-diff`, `get-pr-checks`, `get-required-checks`,
`create-pr`, `review-pr`, `comment-pr`, `assign-pr`, `unassign-pr`, `label-pr`,
`unlabel-pr`, `search-prs`, `list-prs`, `default-branch`, `repo-info`,
`get-issue`, `comment-issue`, `assign-issue`, `unassign-issue`, `close-issue`,
`checkout-pr`, `attach-image-evidence`, and the label guards `label_exists` /
`apply_label` / `apply_issue_label` / `remove_issue_label` /
`set_pipeline_label`. A new skill uses these names; it does not invent CLI calls.

## Shared pipeline protocols to reuse (don't reinvent)

A generated skill that touches PRs/issues should reuse these, not parallel copies:

- **Config load**: the standard `.ai/agentic.config.json` snippet from the
  `om-setup-agent-pipeline` skill, resolving `BASE_BRANCH`, `LABELS_ENABLED`,
  `QA_GATE`, `TRACKER`/`TRACKER_FILE`, and `validation.commands`.
- **Repo-local extension check**: after loading config, check for
  `.ai/skills/<name>/SKILL.md` and treat it as configuration that can add repo
  specifics but cannot relax safety.
- **Three-signal claim/lock**: assignee + `in-progress` label + claim comment;
  release in a `trap`/finally even on failure; stale-lock recovery; `--force` to
  override with an explicit comment.
- **Isolated worktree**: detect an existing linked worktree, otherwise create a
  temporary one under `.ai/tmp/<skill>/`, and clean up only what this run
  created — never touch the primary worktree.
- **Label taxonomy**: pipeline labels (`review`, `changes-requested`, `qa`,
  `qa-failed`, `merge-queue`, `blocked`, `do-not-merge`), meta labels
  (`needs-qa`, `skip-qa`, `qa-approved`, `qa-self-verified`, `in-progress`, `ci-monitoring`), one
  `priority-*` and one `risk-*`; every mutation goes through the descriptor's
  label guards.
- **Chain handoff contract**: emit the chaining reference lines on success —
  `PR: #<number> (link: <url>)`, plus `Issue:` / `Spec:` lines where defined
  (parse legacy `PR_URL=` / `PR_NUMBER=` / `SPEC_PATH=` input, never emit it);
  read prior-step output from a `— PREVIOUS STEP (<skill>) said —` block; honor
  `NO_ACTION_NEEDED` / `LOW_CONFIDENCE` tokens; keep each chain skill
  independently runnable.
- **Validation gate**: run `validation.commands` in order; every failure is a
  finding.

The paste-ready text for the config-load, repo-local-extension, and
untrusted-content-boundary blocks lives in `references/shared-boilerplate.md`.
