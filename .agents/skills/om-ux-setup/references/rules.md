# Shared rules

Canonical rules shared by every skill in this collection. They always apply,
in addition to the skill-specific rules in the skill body. On conflict, the
stricter rule wins.

- **Interactive run — a user is in the loop.** This skill acts once, asks the
  questions only the team can answer, reports, and hands control back. It is
  not an `om-auto-*` skill: it chains no further skills and continues past the
  report only when asked.
- **Secrets hygiene.** Never paste secrets, tokens, `.env` content, or raw
  credentials into the contract, reports, or logs. The contract describes
  design, not configuration.
- **Emoji glossary** in user-facing output: 🎯 goal · 📋 plan · 📝 spec · 🏷️ labels · 📸 evidence · 🔍 review · 🧪 tests · 💥 breaking · ✅ pass · ❌ fail · ⚠️ needs-human · ⛔ blocked · 🔁 resume · 🚀 merge/release. Emojis decorate; parsers key on text markers only.
- **Reporting style.** User-facing output is a deliverable, not a log: write
  complete sentences and explain the why behind every number you report.
  Never compress reporting to save tokens. Fill the shape in
  `references/report-templates.md` exactly and expand with detail.
- **Reader's language over method vocabulary.** Report what was found in the
  repository and what it means for the team, not the names of the extractor's
  internal steps.

## om-ux-setup specifics

- **Extract, do not judge.** This skill writes the contract and hands it over.
  Producing findings, verdicts, or a review of code, screens, or mockups is
  out of scope even when the user's wider goal is an evaluation: name the
  skill that owns that work and let the user invoke it.

- **Never regenerate silently.** An existing contract is team knowledge under
  review. When one is present, say what a refresh would change and let the
  user decide.
- **The manual section is sacred.** Everything between the manual markers in
  `conventions.md` survives every regeneration. If a tool or a step would
  drop it, stop and report ⛔ rather than losing the team's own rules.
- **Proposed is not enforced.** In a repository with no declared tokens, the
  derived palette is documentation of what the code already does. It never
  arms the audit and is never cited as a `[PRODUCT]` rule until the team
  declares real tokens.
- The contract is committed like code. Treat a contract change as a code
  change: reviewable, diffable, and explained in the commit.
