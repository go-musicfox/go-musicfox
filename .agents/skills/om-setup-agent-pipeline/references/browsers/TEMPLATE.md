# Browser provider: <name>

This descriptor implements the browser-automation operations used by
`om-prepare-test-env`, `om-auto-qa-pr`, and `om-integration-tests`. The
config's `browser.provider` value selects the committed copy at
`.ai/browsers/<name>.md`.

## Contract

A provider descriptor MUST define every operation below. Commands may differ by
platform, but the observable behavior and output stay the same.

### ensure-installed

Inputs: `QA_DIR`, current platform, and architecture. Install the provider and
its local browser engine without asking the operator to preinstall a runtime,
package manager, browser, or OS library. Reuse a healthy installation. On Linux,
attempt required system-library installation non-interactively when the agent
already has root or passwordless elevation. Never silently fall back to a cloud
browser or request third-party credentials.

Output these machine-readable lines:

```text
BROWSER_PROVIDER=<name>
BROWSER_INSTALLED=0|1
BROWSER_COMMAND=<absolute command path or repository runner command>
BROWSER_VERSION=<version or unknown>
BROWSER_NOTES=<empty or concrete blocker>
```

Consumers parse everything after the first `=` as the value; they never source
these lines as shell code. This preserves Windows command paths containing
spaces.

Unsupported OS/CPU combinations and exhausted privilege failures are explicit
blockers. Never report installed until the live-launch check passes.

### doctor

Run the provider's non-destructive health check, including a real headless
browser launch. Return non-zero when navigation and screenshots would fail.

### open

Input: `BASE_URL` or a validated URL. Open it in an isolated session scoped to
the current run. Output enough state for later operations to address the same
session.

### snapshot

Return an accessibility/interactive snapshot with stable element references.
The executing agent uses only references observed here; it never guesses
selectors.

### interact

Input: one observed element reference plus an action (`click`, `fill`, keyboard
input, select, or equivalent). Execute the action and return non-zero on failure.
Re-snapshot after navigation or material DOM changes.

### assert

Input: an observed condition such as visible text, URL, element state, or value.
Return zero only when the condition is observed. Include the observed value in
output so reports are evidence-backed.

### screenshot

Input: an output PNG path, optionally full-page. Write the screenshot exactly at
that path and return non-zero when capture fails.

### close

Close only the session created by **open**. Safe to call more than once.

## Rules

- Keep installation agent-owned and local. Never require the operator to run a
  prerequisite command or subscribe to a remote browser service.
- Keep sessions isolated per run and close them in `trap`/`finally` cleanup.
- Never put real credentials in commands, snapshots, screenshots, or logs.
- A repository-native committed E2E runner remains authoritative. This provider
  owns agent-driven exploration, assertions, and evidence capture; it does not
  replace an existing suite merely because it was selected.
