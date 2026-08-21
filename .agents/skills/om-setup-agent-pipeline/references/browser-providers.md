# Browser providers

Browser-capable skills use the same committed-descriptor pattern as trackers.
They name provider operations — **ensure-installed**, **doctor**, **open**,
**snapshot**, **interact**, **assert**, **screenshot**, and **close** — and read
`.ai/browsers/<provider>.md`, selected by `browser.provider`. The repo's copy is
authoritative and may extend the shipped operations without editing installed
skills.

This collection ships `agent-browser.md` and `playwright.md`, plus
`references/browsers/TEMPLATE.md` for custom providers. `agent-browser` is the
fresh-setup default and self-provisions its native CLI, Chrome for Testing, and
available OS libraries on macOS, Linux, WSL2, Git Bash, and native Windows. It
uses local browser processes only — no cloud-browser account or API key.

Backward compatibility is deliberate: a config without `browser.provider` is
read as `playwright`, and browser consumers accept the legacy `playwright`
object in `test-env.json`. Re-run this setup skill or `om-apply-upgrade-notes`
to make the choice explicit and install a browser descriptor.
