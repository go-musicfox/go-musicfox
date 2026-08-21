# Shared rules

Canonical rules shared by every skill in this collection. They always apply,
in addition to the skill-specific rules in the skill body. On conflict, the
stricter rule wins.

- **Interactive run — a user is in the loop.** This skill acts once, may ask
  the few questions that could change the direction, reports, and hands
  control back. It is not an `om-auto-*` skill: it chains no further skills
  and starts no implementation unless asked.
- **Secrets hygiene.** Never paste secrets, tokens, `.env` content, or raw
  credentials into specs, handoffs, or reports.
- **Emoji glossary** in user-facing output: 🎯 goal · 📋 plan · 📝 spec · 🏷️ labels · 📸 evidence · 🔍 review · 🧪 tests · 💥 breaking · ✅ pass · ❌ fail · ⚠️ needs-human · ⛔ blocked · 🔁 resume · 🚀 merge/release. Emojis decorate; parsers key on text markers only.
- **Reporting style.** User-facing output is a deliverable, not a log: write
  complete sentences, explain the why behind every verdict and recommendation,
  and structure sections with the glossary emojis. Never compress reporting to
  save tokens. Fill the shapes in `references/report-templates.md` exactly.
- **Reader's language over method vocabulary.** The framework's vocabulary is
  for the author, never the reader: section names like evidence ledger, value
  gaps, complexity hotspots, behavioral signal, or guardrail must not appear
  in the delivered text. Render each as a plain statement about screens,
  behavior, and what to change.
- **The result obeys the rules it enforces.** Before delivering, check the
  text against the repository's own copy conventions (the manual section of
  `.uxproof/conventions.md` and any team rules the user has stated). A
  recommendation that flags a copy violation while committing it is invalid.

## om-ux-shape specifics

- **Never invent evidence.** No fabricated research, user quotes, metrics, or
  constraints. An unverified belief is labeled as an assumption and given a
  validation path; a provisional persona is not evidence.
- **Recommend, do not enumerate.** One direction with the decisive trade-off
  beats an unranked menu. Rejected options stay internal unless they help the
  reader understand a consequential choice.
- **Depth follows risk.** A small, reversible decision gets a short answer.
  Do not make a low-risk task carry the full process.
- This skill writes no code and touches no tracker item. Its Handoff output is
  written to be consumed by the collection's implementing skills.
