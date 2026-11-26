# Shared test environment: reuse and fast bootstrap

How this skill attaches to the shared test environment written by `om-prepare-test-env`, and when it may skip the slow bootstrap chain. `om-prepare-test-env` owns the full protocol; this file is the consumer-side contract.

## Reuse the shared descriptor

Before discovering how to run the app yourself, check for a shared environment descriptor written by `om-prepare-test-env` at `<paths.qa>/test-env.json` (default `.ai/qa/test-env.json`). When it reports `"status":"running"` and passes the validation in the fast-bootstrap contract below (PID alive, readiness probe answers, still fresh), **attach to that instance** — read `baseUrl`, `credentials`, the provider-neutral `browser` object, and `testRunner` from it. Resolve older descriptors through the legacy `playwright` object. The `credentials` entries are references: `passwordEnv` names a variable in the gitignored `credentialsFile` env file, which the shell loads (`set -a; . "$CREDENTIALS_FILE"; set +a`) so commands can use `"$TEST_ADMIN_PASSWORD"` literally — the shell expands the value, the agent never reads the file and never learns it. Never hardcode a password value into a test file or repeat one in reports, comments, or logs; a legacy descriptor with inline values is passed through the runner's environment the same way. This is how QA (`om-auto-qa-pr`) and integration tests share one environment instead of each booting their own, so runs are faster and identical.

When no descriptor exists (or it is stale), invoke `om-prepare-test-env` to discover or provision one — it establishes the run command, provisions backing services, installs the configured browser provider, and writes the descriptor — then attach. Fall back to manual discovery (skill body, step 3) only when `om-prepare-test-env` is unavailable or the user asked to run against an already-running instance.

## Fast bootstrap: reuse the environment, cache the build

Bootstrapping a test environment — install, codegen, build, provision a database, seed, start, wait for readiness — is usually the slowest part of an integration run. Prepare and reuse the environment through the `om-prepare-test-env` skill, which owns the full protocol; the contract in short:

- **Reuse first.** Read the environment descriptor (`.ai/qa/test-env.json`, or the repo tooling's own state file) before starting anything, and reuse the recorded environment only after validating it: the owning PID is alive (`kill -0`), a real readiness probe answers (shell page → API → one authenticated round trip), and the env is fresh (within TTL and no tracked source file modified since `startedAt`). Anything stale gets cleared and rebuilt — never test against stale code.
- **Cache the build.** Skip the preparation/build chain only when the build-cache descriptor's fingerprints match (source `path:size:mtime` hash, build-shaping env vars) and every recorded artifact exists; when in doubt, rebuild.
- **Prepare fresh workspaces.** In a fresh checkout or worktree, run the repo's preparation chain (install → codegen → build, in the order the repo's scripts and CI imply) before launching any discovered test-env command — such commands assume a built workspace.
- **Lock the bootstrap.** One PID-checked bootstrap at a time per checkout; a concurrent caller waits, then reuses what the other produced.
- **Honor repo tooling.** When the repo's own test-env tooling implements reuse/caching (state files, reuse flags, cache TTLs), use its mechanism and flags instead of duplicating it.
- **Record lessons.** A bootstrap failure that taught you a prerequisite goes into the repo-local skill, the generated scripts, and the descriptor notes before you finish — next time must not repeat it.
