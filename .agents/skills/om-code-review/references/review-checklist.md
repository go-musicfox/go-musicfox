# Code Review Checklist — Full Reference

A stack-agnostic staff-engineer review checklist. Apply every applicable section based on which files changed; skip sections that don't apply to the diff. Each check is phrased as a question to ask of the change, with the severity a violation typically earns in parentheses. Severities and the verdict rule are defined at the end.

When the pipeline config (`.ai/agentic.config.json`) sets `reviewChecklist`, apply that repo-local file in addition to this one — it extends these rules, never replaces them.

## 1. Correctness & Edge Cases

### Inputs and boundaries

- [ ] Does the change do what the PR description or linked issue says — and nothing it doesn't say? (mismatch: major)
- [ ] What happens on empty input: empty string, empty list, zero rows, missing file, absent header? (major)
- [ ] What happens on null/absent values at every new property access or dereference? (major)
- [ ] Are numeric boundaries handled: zero, negative, maximum, overflow, off-by-one in ranges and pagination (first page, last page, exactly page-size items)? (major)
- [ ] Does string handling survive unicode, whitespace-only input, mixed line endings, and very long inputs? (minor; major when it feeds storage or security decisions)
- [ ] Are encoding boundaries respected — bytes vs characters, declared charset vs actual content? (major)

### Branching and state

- [ ] Are all branches of a switch/match handled, including the default? If an enum or variant was added, did every consumer that branches on it get updated? (major)
- [ ] Are error, null, and sentinel return values actually checked by the callers this diff introduces? (major)
- [ ] Does the new code preserve invariants the surrounding code assumes — sorted order, non-emptiness, uniqueness, referential integrity? (major)
- [ ] For state machines: are invalid transitions rejected rather than silently applied? (major)
- [ ] Do early returns skip required cleanup, counters, or audit writes? (major)
- [ ] When the code mutates a shared or aliased structure, is a defensive copy needed to avoid surprising a caller? (major)

### Time, numbers, and identity

- [ ] Is time handled correctly: timezones, DST transitions, comparisons using a consistent clock, expiry checked against the same clock that issued the timestamp? (major)
- [ ] Is money computed with exact types (integers in minor units or decimals), never binary floating point? (blocker for billing paths)
- [ ] Are rounding and truncation deliberate and consistent with what the rest of the system does? (major for money; minor otherwise)
- [ ] Are comparisons using the right notion of equality — identity vs value, case sensitivity, locale-aware collation? (major)
- [ ] Is the operation idempotent where its trigger can fire twice (retries, double-clicks, redelivery)? (major)

## 2. Security

### Authentication & authorization

- [ ] Does every new endpoint, handler, or command enforce authentication server-side? "The UI hides the button" is not enforcement. (blocker)
- [ ] Does authorization check the specific record, not just the caller's role — can this caller act on THIS object (no insecure direct object references)? (blocker)
- [ ] Are permission checks applied on every path to the operation, including bulk endpoints, exports, background triggers, and streaming or long-poll channels? (blocker)
- [ ] Do privilege changes (role edits, permission grants) require a higher privilege than the one being granted? (blocker)
- [ ] Are session and token lifetimes, revocation, and rotation unaffected — or deliberately changed with review? (major)
- [ ] Do new auth-adjacent endpoints (login, reset, invite, verify) have rate limiting or another brute-force control? (major)
- [ ] Are secret comparisons (tokens, signatures, MACs) done with constant-time comparison? (major)

### Injection & unsafe input

- [ ] Is untrusted input parameterized in every query — never concatenated into SQL, query DSLs, or ORM raw fragments? (blocker)
- [ ] Is untrusted input kept out of shell commands, or passed as argument arrays with no shell interpolation? (blocker)
- [ ] Are file paths built from user input canonicalized and checked against a base directory (no path traversal)? (blocker)
- [ ] Are URLs fetched server-side validated against an allowlist (no server-side request forgery to internal addresses)? (blocker)
- [ ] Are redirect targets taken from user input validated (no open redirects)? (major)
- [ ] Is user content escaped or sanitized before rendering into HTML, and never fed to `eval`-like sinks? (blocker)
- [ ] Is deserialization of untrusted data restricted to safe formats and explicit schemas? (blocker)
- [ ] Do write endpoints bind an explicit allowlist of fields rather than accepting the whole request body into the model (no mass assignment)? (blocker)
- [ ] Are file uploads validated for type and size, and stored where they cannot be executed or served as code? (blocker/major)

### Secrets

- [ ] Are there no credentials, API keys, tokens, or private keys committed in code, config, fixtures, or test data? (blocker)
- [ ] Are secrets kept out of logs, error messages, stack traces, URLs, and client-visible responses? (blocker)
- [ ] Are secrets read from the environment or a secret store, not baked into build artifacts shipped to clients? (blocker)
- [ ] When auth cookies or headers are touched: are the protective flags and policies (secure, http-only, same-site or the platform equivalents) preserved? (major)

### Input validation & data scoping

- [ ] Is every input validated at the trust boundary with an explicit schema — allowlist of fields, types, ranges, sizes — not just used and hoped for? (major; blocker when it guards a write)
- [ ] Are size limits enforced on request bodies, uploads, and collection parameters? (major)
- [ ] Does every query on scoped data filter by the owning scope (user, account, team, workspace)? Do list endpoints, search, exports, and aggregates preserve that scoping? (blocker)
- [ ] Can an identifier from one scope be replayed in another — a cross-scope reference smuggled through a foreign key or lookup? (blocker)
- [ ] Are passwords hashed with a slow, salted algorithm, and do auth errors avoid revealing whether an account exists? (blocker / major)
- [ ] Is any home-rolled cryptography introduced where a standard library primitive exists? (blocker)

## 3. Breaking Changes & Contract Stability

A contract surface is anything an external consumer may depend on. Removing or renaming one without a deprecation path is a blocker; the deprecation path is: mark deprecated → keep a working bridge for a documented window → remove later, with migration notes. When the project documents its own compatibility policy, apply it on top of these checks.

### Code-level contracts

- [ ] Exported/public APIs: was any symbol removed or renamed, a required parameter added, parameters reordered, a return type changed, a type field removed or narrowed? (blocker)
- [ ] Are new parameters added as optional with sensible defaults, so existing callers compile and behave unchanged? (blocker when not)
- [ ] Do documented public import or module paths still resolve, with moved code re-exported from the old location during the bridge window? (blocker)
- [ ] Did the minimum supported runtime or platform version get raised silently? (major)

### Wire-level contracts

- [ ] HTTP routes: was any URL removed or renamed, a method changed, a response field removed or retyped, a status code or error shape that clients branch on changed? Adding optional fields is fine. (blocker)
- [ ] Events and messages: was any event name renamed or removed, or a payload field consumers rely on removed or retyped? Dual-emit old and new during a bridge window. (blocker)
- [ ] Webhooks and callbacks delivered to external consumers: are payload shape and signature scheme unchanged, or versioned? (blocker)
- [ ] CLI: was any command or flag renamed or removed, a default changed, or output that scripts parse reformatted? (major; blocker for published tools)

### Data-level contracts

- [ ] Database schema: was any table or column renamed or dropped, a type narrowed, a default removed, or a constraint tightened against data that already exists in production? (blocker)
- [ ] Serialized data: do cached values, stored blobs, cookies, or client-persisted state written by the old code still deserialize under the new code? (major)
- [ ] Identifiers stored in data (permission names, status values, type discriminators): were any renamed without a data migration for existing rows? (blocker)
- [ ] Config formats: were keys renamed, defaults changed silently, or old config files made unparseable? (major)
- [ ] Feature flags: did a default flip that changes behavior for existing installations without an announcement? (major)
- [ ] Are deprecations announced where the project announces them (changelog, release notes), with a stated removal target? (minor)

## 4. Tests

- [ ] Is every behavior change covered by a test that fails without the change? Run the test mentally against the pre-change code. (major)
- [ ] Does a bug fix ship a regression test that reproduces the original bug? (major)
- [ ] Were any assertions weakened to make the suite pass — equality loosened to "contains", exact counts removed, tolerances raised, assertions deleted, tests skipped or marked pending? (major)
- [ ] Were snapshot or golden files updated with review of the actual differences, not regenerated blindly? (major)
- [ ] Do tests exercise the real code path rather than mocking the unit under test into a tautology? (major)
- [ ] Are failure and edge paths tested, not only the happy path? (major)
- [ ] Are the tests deterministic: no sleeps as synchronization, no dependence on wall-clock time, test order, network, or shared global state? (minor; major when already flaky)
- [ ] Do risk-heavy paths — permissions, data scoping, money, migrations, concurrency, external contracts — get integration-level coverage, not just unit mocks? (major)
- [ ] Does a concurrency-sensitive fix have a test, or a documented reason why one is impractical plus the manual verification performed? (major)
- [ ] Do test names describe the behavior under test, and do failures produce a message a stranger can act on? (nit)
- [ ] Is test data free of real personal data and real credentials? (blocker)
- [ ] Were fixtures and factories updated rather than duplicated? (minor)
- [ ] If a test was deleted, is the behavior it protected either gone or covered elsewhere? (major)

## 5. Error Handling & Failure Modes

- [ ] Is every caught exception handled, logged with context, rethrown, or explicitly documented as safe to ignore — no silent empty catch blocks? (major)
- [ ] Does error wrapping preserve the original cause and stack so the root failure stays diagnosable? (minor)
- [ ] Are errors at external boundaries (network, disk, subprocess, parse) anticipated and handled, not allowed to surface as opaque crashes? (major)
- [ ] For batch operations, is partial failure defined — abort all, skip and report, or retry — and is the choice deliberate? (major)
- [ ] Does failure leave the system consistent: multi-step writes in a transaction or with compensating cleanup; temp files, locks, and connections released in finally paths? (blocker for data integrity)
- [ ] Is fallback behavior deliberate — and fail-closed wherever the decision affects security or money? A permission check that fails open on error is a hole. (blocker)
- [ ] Is retry logic bounded, with backoff, and only around idempotent operations? (major)
- [ ] Do all outbound calls have timeouts so a hung dependency cannot hang the caller? (major)
- [ ] Do failure paths avoid unbounded growth — queues, buffers, or retry backlogs that accumulate while a dependency is down? (major)
- [ ] Are user-facing error messages actionable without leaking internals (stack traces, query text, file paths)? (minor; major when they leak)
- [ ] What happens if the process dies mid-operation — does startup recovery or the next run tolerate the half-done state? (major)
- [ ] Are new error types or codes consistent with how the codebase already models errors? (minor)

## 6. Performance

### Data access

- [ ] Is there a query, network call, or file read inside a loop that should be a batch operation (the N+1 pattern)? (major)
- [ ] Is every list query bounded — pagination or an explicit limit on anything that grows with data volume? (major)
- [ ] Do new query patterns on large tables have supporting indexes, and do new indexes not duplicate existing ones? (major)
- [ ] Is anything fetched and then filtered in memory when the datastore could filter it? (major)
- [ ] Is a whole file or table loaded into memory where streaming or chunking would do? (major)
- [ ] Are connections, clients, and pools reused rather than constructed per request? (major)

### Computation and allocation

- [ ] Are there accidental O(n²) shapes: membership tests or lookups inside a loop over the same collection — where a set or map would do? (major on unbounded n; minor otherwise)
- [ ] On hot paths: repeated compilation of regexes, repeated parsing of the same config, synchronous I/O, string concatenation in tight loops, per-item allocations that could be hoisted? (minor; major on measured hot paths)
- [ ] Does the change add meaningful startup or cold-start cost (eager loading, upfront network calls)? (minor)
- [ ] Could new locking serialize a hot path (coarse lock around work that could run concurrently)? (major)

### Payloads and caching

- [ ] Are payloads proportionate: no over-fetching of columns or relations, no unbounded response bodies, no large assets added to a critical path? (minor/major)
- [ ] If a cache was added: is invalidation wired to every write path, are keys scoped correctly, and is the stale-read window acceptable? (major; blocker if stale reads cross data scopes)
- [ ] Are performance claims in the PR backed by a measurement, not an assertion? (minor)

## 7. Concurrency & Race Conditions

- [ ] Any check-then-act sequence (read, verify, write) that two concurrent callers can interleave — is it protected by an atomic operation, a lock, or a unique constraint? (blocker for money, inventory, or uniqueness invariants; otherwise major)
- [ ] Is there a database unique constraint as the backstop for any application-level uniqueness check? (major)
- [ ] Is shared mutable state (globals, singletons, module-level caches) safe under concurrent requests, threads, or interleaved async callbacks? (blocker/major)
- [ ] Are queue consumers and webhook handlers idempotent under at-least-once delivery — duplicate execution MUST NOT corrupt data? (major)
- [ ] Do consumers avoid depending on cross-queue or cross-partition ordering that the transport does not guarantee? (major)
- [ ] Can two code paths acquire the same locks in different orders (deadlock), or is a slow call awaited while a lock is held? (major)
- [ ] Do distributed locks and leases expire safely relative to the duration of the work they guard? (major)
- [ ] Are transaction isolation assumptions correct for the datastore's actual default level? (major)
- [ ] Filesystem check-then-use (exists, then open/write) — is it tolerant of the file changing in between? (major when attacker-reachable, else minor)
- [ ] Concurrent edits to the same record: is there an explicit strategy (optimistic version check, lock, merge), or is last-write-wins being accepted silently? (major)
- [ ] Are all promises/futures awaited or explicitly detached with their errors observed — no fire-and-forget that swallows failures? (major)
- [ ] On shutdown or deploy, is in-flight work drained or safely resumable? (minor/major)

## 8. Readability & Naming

- [ ] Do names say what things are and do — and were names updated when behavior changed underneath them? A misleading name is worse than a vague one. (minor; major when actively misleading)
- [ ] Is one concept called by one name throughout the diff, matching what the codebase already calls it? (minor)
- [ ] Does each function do one thing at one level of abstraction; is deep nesting flattened with early returns? (minor)
- [ ] Are error paths readable — handled early and locally — rather than woven through the happy path? (minor)
- [ ] Are boolean names positive and unambiguous (no double negatives to parse at the call site)? (nit)
- [ ] Are bare boolean arguments readable at the call site, or should they be named options or separate functions? (nit)
- [ ] Are magic numbers and strings named constants with the reason for the value discoverable? (minor)
- [ ] Is dead code, commented-out code, and leftover debug output removed rather than shipped? (minor)
- [ ] Can a competent stranger follow the control flow top to bottom without simulating the runtime in their head? Clever one-liners rewritten for the next reader? (minor)
- [ ] Does the change match the project's documented style and design conventions — formatting, structure, design-system tokens or variables where the project defines them? (minor)
- [ ] No one-letter variable names outside conventional tight scopes (loop indices). (nit)

## 9. Scope Discipline

- [ ] Does every changed line trace to the stated purpose of the change? Drive-by reformatting, renames, and refactors mixed into a fix inflate review risk. (minor; major when the churn obscures the behavioral change)
- [ ] Is mechanical churn (formatting-only, generated-file regeneration) separated into its own commits or PRs where the project allows? (minor)
- [ ] Could the diff be split into independently reviewable pieces? Reviewability is a feature; an unreviewable PR hides bugs. (minor)
- [ ] Is anything speculative: unused options, configuration for imagined futures, abstractions with a single implementation and no second caller in sight? (minor)
- [ ] When an internal API changed, were ALL callers migrated in the same change — no half-migrated state where old and new coexist without a plan? (major)
- [ ] Did unrelated lockfile, dependency, or tool-config drift sneak into the diff by accident? (major — it silently changes build behavior)
- [ ] Are generated files consistent with their sources, regenerated by the tool rather than hand-edited? (major)
- [ ] Were files deleted or moved that the change didn't need to touch? (minor)
- [ ] If the change adds a temporary toggle or workaround, is its removal tracked (issue link or dated TODO)? (minor)

## 10. Dependency Hygiene

- [ ] Is the new dependency necessary — could the standard library or twenty lines of local code cover it? Every dependency is a permanent liability. (minor; major for heavy or invasive ones)
- [ ] Does it duplicate functionality of a dependency the project already has? (minor)
- [ ] Is the package healthy: maintained, widely used, a compatible license, a reasonable transitive tree, no known vulnerabilities? (major; blocker for known-vulnerable versions)
- [ ] Is the exact package name correct — no typosquat of a popular name? (blocker)
- [ ] Is the version pinned or locked consistently with the repo's practice, and is the lockfile (with its integrity hashes) updated in the same change? (major)
- [ ] Is the dependency classified correctly — build/test-only dependencies not shipped in the runtime set? (minor; major when it bloats production)
- [ ] For upgrades: was the changelog reviewed for breaking changes, and are major-version bumps kept out of unrelated feature work? (major)
- [ ] Is vendored or copy-pasted third-party code attributed with its license? (major)
- [ ] Do install scripts or postinstall hooks of the new dependency do anything surprising? (blocker when they do)

## 11. Observability

- [ ] Are new failure modes logged with enough context to debug in production — operation, identifiers, cause — so the on-call engineer isn't guessing? (minor; major for new critical paths with no signal)
- [ ] Do log messages exclude secrets, tokens, passwords, and personal data? (blocker for secrets; major for personal data)
- [ ] Are log levels honest: expected conditions not logged as errors, real errors not demoted to debug? (minor)
- [ ] Is there no per-item logging inside hot loops flooding the log budget? (minor)
- [ ] Are correlation or request identifiers propagated through new code paths where the project uses them? (minor)
- [ ] When the project uses metrics or tracing, are new operations instrumented consistently — and are renamed metrics or log lines that dashboards and alerts parse treated as a breaking change? (minor; major for renames)
- [ ] Can a user-visible failure be distinguished from a bug from the logs alone? (minor)
- [ ] Are security-relevant decisions (denied access, failed auth) auditable where the project keeps an audit trail? (minor; major for regulated flows)

## 12. Docs & Comments

- [ ] Do comments explain WHY — invariants, non-obvious constraints, links to the decision — rather than narrating what the code plainly does? (nit)
- [ ] Were docstrings or comments added to code the change didn't touch? They don't belong in this diff. (nit)
- [ ] Were existing comments and docs updated where behavior changed — a stale comment is worse than none? (minor; major when it now says something false)
- [ ] Are README, setup instructions, config references, and help text updated when flags, environment variables, or steps changed? (minor)
- [ ] Do published API docs match the new signatures and response shapes? (major for published APIs)
- [ ] Do code examples in the docs still run against the changed code? (minor)
- [ ] Is a significant design decision recorded where the repo keeps decision records? (minor)
- [ ] Does every TODO carry an owner or an issue link, not just a wish? (nit)

## Anti-Pattern Quick Table

Flag any of these on sight:

| Anti-pattern | Severity | Fix |
|---|---|---|
| Public contract surface removed/renamed without a deprecation bridge | blocker | Keep the old surface working, add the new one, deprecate with a removal target |
| Missing owning-scope filter on a query over scoped data | blocker | Add the scope filter; add a test proving isolation |
| Permission check only in the UI | blocker | Enforce server-side; the UI merely reflects it |
| Untrusted input concatenated into a query, command, or path | blocker | Parameterize; canonicalize paths against a base directory |
| Secret committed, logged, or echoed to the client | blocker | Remove, rotate the secret, load from a secret store |
| Check-then-act race on money, inventory, or uniqueness | blocker | Atomic operation, lock, or unique constraint |
| Fail-open error handling around a security decision | blocker | Fail closed; deny on error |
| Migration with destructive or unrelated schema churn | blocker | Regenerate scoped to the intended entities only |
| Bug fix without a regression test | major | Add a test that reproduces the original bug |
| Weakened or deleted assertions to get the suite green | major | Restore the assertion; fix the code instead |
| Query or network call inside a loop | major | Batch, prefetch, or join |
| Unbounded list query on data that grows | major | Paginate or cap with an explicit limit |
| Outbound call without a timeout | major | Set an explicit timeout and handle its expiry |
| Empty catch block | major | Handle, log with context, rethrow, or document the ignore |
| Non-idempotent consumer under at-least-once delivery | major | Make it idempotent (dedup key, upsert, version check) |
| Lockfile or dependency drift unrelated to the change | major | Revert the drift or split it into its own change |
| Hand-edited generated file | major | Regenerate from the source of truth |
| Hard-coded user-facing string in a localized codebase | minor | Route through the localization mechanism |
| Hard-coded style values where the design system defines tokens | minor | Use the documented tokens or variables |
| Dead or commented-out code shipped | minor | Delete it; version control remembers |
| Stale comment contradicting the code | minor | Update or remove the comment |
| One-letter variable name outside a loop index | nit | Use a descriptive name |
| Comment narrating self-explanatory code | nit | Remove the comment |
| Docstring added to unchanged code | nit | Remove it from this diff |

## Severity Scale

| Severity | Meaning | Examples |
|----------|---------|----------|
| **blocker** | Merging this causes or invites real damage: security holes, data loss or corruption, cross-scope data leaks, silently broken public contracts, failing validation gates | Missing permission check; SQL built by concatenation; dropped column consumers still read; secret in a log line |
| **major** | A defect or debt that must be resolved before merge unless the maintainer explicitly accepts and documents the risk | Correctness bug on a realistic path; bug fix without a regression test; unbounded query; check-then-act race; weakened assertions |
| **minor** | Should be fixed, but does not block the merge on its own | Convention violation; readability problem; stale comment; missing log context |
| **nit** | Optional polish; the author decides | Naming taste; comment phrasing; boolean-argument readability |

## Verdict Rule

- Any **blocker** → **request changes**. No exceptions, no matter how green the other checks are.
- Any **major** without an explicit, documented waiver from the maintainer → **request changes**.
- Only minors and nits → **approve**, and list them in the review so the author can pick them up.
