# Diff-level auto-detections

The severity-tagged pattern tables the `om-auto-review-pr` body scans the PR diff
against in step 7, before running the full `om-code-review` skill. When a pattern
applies to this repository's stack and conventions, it is a mandatory finding,
not an optional heuristic; skip rows that have no equivalent in this codebase
(for example, the i18n row in a repo without i18n).

## Blocker auto-detections

| Pattern in diff | Finding |
|-----------------|---------|
| Removed or renamed a published event name, message topic, or webhook type | Blocker: published event names are a frozen contract surface |
| Removed a field from an API response schema or serialized response type | Blocker: response fields are additive-only |
| Renamed or removed a database column or table in a migration without a migration path | Blocker: destructive schema changes need an explicit migration/deprecation plan |
| Removed a public export or import path without a re-export bridge or deprecation note | Blocker: public entry points require a deprecation window |
| A query missing the data-scoping filter (account/workspace/organization ID) that sibling queries in the same area apply | Blocker: data-scoping breach |
| A shared data-access or security wrapper (encryption, sanitization, guarded client) replaced with a raw lower-level call | Blocker: downgrading an established security wrapper is a security regression |

## Major auto-detections

| Pattern in diff | Finding |
|-----------------|---------|
| New route, handler, subscriber, or worker file missing the registration or metadata exports the codebase's conventions require | Major: required exports for discovery/registration |
| Direct low-level HTTP or data call in UI or page code, outside tests, where the repo provides a shared client helper | Major: must use the shared client helper |
| Behavior change with no corresponding test file in the diff | Major: behavior changes must include tests |
| Entity or schema changed but no migration file or no-op rationale in the diff | Major: schema changes must ship with a scoped migration |
| Hand-written migration SQL that bypasses the repo's migration tooling without a scoped rationale | Major: prefer generated/tooled migrations; manual SQL must be scoped and keep the tooling's state files in sync |
| Missing explicit data scoping in sub-entity queries | Major: defense in depth |

## Minor auto-detections

| Pattern in diff | Finding |
|-----------------|---------|
| Hardcoded user-facing string in API errors or UI labels, in a repo that uses an i18n system | Minor: must route through i18n |
| New `any` type annotation (or the language's equivalent unchecked cast) outside tests | Minor: use typed schemas and runtime narrowing |
| Ad-hoc `alert(` or custom toast instead of the repo's standard notification helper | Minor: use the standard helper |

## Nit auto-detections

| Pattern in diff | Finding |
|-----------------|---------|
| One-letter variable name outside loop counters `i`, `j`, `k` | Nit: use descriptive names |
| Inline comment on self-explanatory code | Nit: remove comment |
| Added docstring or comment on unchanged function | Nit: do not annotate unchanged code |
