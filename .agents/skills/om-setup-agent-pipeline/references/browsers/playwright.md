# Browser provider: playwright

This is the compatibility provider for repositories that use Playwright for
agent-driven browser exploration and evidence capture.

## Operations

### ensure-installed

If the repository already has `@playwright/test` or a `playwright.config.*`, use
its package runner and config. Otherwise install Playwright as a development
dependency with the package manager selected by the repository's lockfile, then
run its browser installer for Chromium with OS dependencies. On Linux, retry the
dependency install only when root or passwordless elevation is already
available; do not ask the operator to perform the install.

Write a minimal shared config at `<paths.qa>/playwright.config.ts` when the repo
has none. It reads `process.env.BASE_URL`; timeouts and retries live in that
shared config.

Output:

```text
BROWSER_PROVIDER=playwright
BROWSER_INSTALLED=0|1
BROWSER_COMMAND=<repository package-runner command>
BROWSER_VERSION=<version or unknown>
BROWSER_NOTES=<empty or concrete blocker>
```

### doctor

Run the repository package runner's Playwright version command, then launch a
one-page Chromium smoke test against a local `data:` URL. Return non-zero when
the browser cannot launch.

### open

Use Playwright's configured Chromium project and a throwaway spec outside the
repository's discovered test directories. Navigate to the validated `BASE_URL`.

### snapshot

Use the live page's accessibility tree or role/label queries. Record exact
roles, labels, text, and element states before interacting.

### interact

Use only role/label/text locators observed by **snapshot**. Do not guess CSS
paths. Re-observe after navigation or material DOM changes.

### assert

Use Playwright web-first assertions in the throwaway spec. Return non-zero on a
failed condition and preserve the error context artifact.

### screenshot

Call `page.screenshot({ path: SCREENSHOT_PATH, fullPage: true })`, then verify
the PNG exists and is non-empty.

### close

Close the throwaway page, context, and browser in `finally`. Safe to repeat.

## Rules

- Use the repository's package manager and existing runner configuration; never
  add a second Playwright installation.
- Keep throwaway exploration specs under the QA artifacts directory, never in a
  discovered test directory.
- Repository-native committed tests remain authoritative and run with the
  repository's existing command.
