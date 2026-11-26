# Evidence tiers

Every claim in a recommendation carries a tag naming the strongest tier it
honestly supports. The tag is part of the contract with the reader: it tells
them how hard to push back.

1. `[PRODUCT]` — this repository's own design contract (`.uxproof/`), its
   analytics, or a documented team decision. Cite the rule or the file.
2. `[STANDARD]` — WCAG, platform guidelines, or a regulation. Name which one
   (for example WCAG 2.4.7, or the DSA for consent patterns).
3. `[PLATFORM]` — default framework or operating-system behavior users
   already expect.
4. `[RESEARCH]` — published usability research. Name the source.
5. `[HEURISTIC]` — a recognized heuristic. Name which one.
6. `[ASSUMPTION]` — reviewer judgment. Allowed, but labeled and falsifiable.

Rules:

- Never dress an `[ASSUMPTION]` as a `[STANDARD]`. Inflating a tier to win an
  argument destroys the value of every other tag in the report.
- A review whose findings are mostly assumptions must say so in its summary,
  so the author knows how much of it is taste.
- Without a design contract, tier 1 is unavailable. Say that once, on the
  Contract line, instead of stretching lower tiers to sound authoritative.
