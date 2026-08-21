# Humane gate — manipulation check

Apply this during every walk. A screen can pass every usability check and
still fail people: deceptive patterns are designs that work FOR the business
BY working against the user. AI-generated UI reproduces them readily,
because they are common in training data and they often improve the metrics
a model is praised for. This gate exists so "converts better" never
outranks "treats people honestly".

Source taxonomy: deceptive.design (Brignull), aligned with the EU Digital
Services Act and the California Privacy Rights Act, both of which prohibit
several of these outright.

## The check

For every screen walked, ask one question about each persuasive element:
**who benefits from this design choice?** When the honest answer is "the
business, at the user's expense", it is a finding — evidence tier
`[STANDARD]` (cite the pattern name below; cite DSA/CPRA where consent or
subscriptions are involved), severity by user harm, not by legal risk.

## Patterns to recognize

**Money and commitment**

- **Hidden costs** — fees appear only after the user invested effort.
- **Hidden subscription** — recurring payment without clear, explicit consent.
- **Hard to cancel** — one click to subscribe, an obstacle course to leave.
- **Currency confusion** — virtual currencies obscuring real spend.
- **Comparison prevention** — bundling or scattering information so offers
  cannot be compared.

**Pressure and emotion**

- **Fake urgency / fake scarcity** — invented timers and stock counters.
- **Confirmshaming** — the decline option worded to humiliate ("No thanks,
  I hate saving money").
- **Nagging** — repeating an interruption until the user gives in.
- **Fake social proof** — invented reviews, activity feeds, popularity.

**Deception in the interface itself**

- **Trick wording** — double negatives, consent questions engineered to
  misread.
- **Visual interference** — the honest option styled to be missed; the
  paid option styled as the only button.
- **Preselection** — defaults chosen for the business, not the user.
- **Disguised ads** — ads dressed as content or controls.
- **Sneaking** — items, costs or commitments added without explicit action.
- **Forced action** — an unrelated obligation gated onto the user's task.
- **Obstruction** — deliberate friction on unprofitable paths.
- **Addictive design** — mechanics engineered for compulsion rather than
  value (streak guilt, infinite variable rewards without stopping cues).

## Reporting rules

- Name the pattern in the finding; the taxonomy name is the citation.
- The pattern line of the finding must show the honest alternative (e.g.
  symmetric Accept/Decline buttons; cancel path as short as signup).
- An improvement in a conversion metric is NOT a trade-off argument that
  can dismiss a finding from this gate — say so explicitly when relevant.
- When the team knowingly chooses to keep such a pattern, record it as an
  explicit decision in the conventions manual section, never as a silent
  outcome of review.
