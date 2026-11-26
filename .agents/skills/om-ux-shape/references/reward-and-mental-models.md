# Reward function and mental models — two decisions teams skip

Distilled from Google's People + AI Guidebook (PAIR), chapters "User Needs +
Defining Success" and "Mental Models". Both are decisions, not polish: skip
them and the model or the marketing makes them for you, badly.

## 1. The reward function: decide which mistake hurts less

Every AI feature makes two kinds of mistakes: a **false alarm** (acting when
it should not have) and a **miss** (staying silent when it should have
acted). Optimizing one always costs the other. Decide explicitly, from the
USER'S seat, which is cheaper:

- A spam filter must prefer letting spam through over eating a real
  message — a missed invoice costs more than one more delete.
- Fraud detection prefers false alarms — an annoyed customer costs less
  than a drained account.
- A mockup generator prefers a visible placeholder over a confidently
  wrong component.

Write the decision into the AI contract as one sentence: "when unsure,
this feature errs toward ___ because for the user ___". Then check that
the UI matches it: the preferred error must be cheap to notice and cheap
to recover from (this connects to the failure-design section of
ai-interaction.md). A team that cannot write this sentence has not decided
what it is optimizing for.

## 2. Mental models: shape the first five minutes deliberately

Users form a theory of what the system is on first contact and then read
everything through it. Three patterns:

- **Lead with the benefit, never the technology.** "Find photos by
  describing them" beats anything containing the word model, AI or
  neural. The technology framing sets expectations of magic that the
  product must then fail.
- **Onboard progressively.** Introduce capabilities as they become
  relevant to the task at hand instead of a feature tour at signup;
  a capability shown before its context is forgotten by the time it
  matters.
- **Set expectations for change.** When the system adapts or improves
  over time, say so upfront ("suggestions get better as you use this")
  and announce noticeable shifts — an unexplained behavior change reads
  as breakage and resets trust to zero.

In Shape mode these feed the interaction contract's capability-framing
row; in Review mode, an AI feature introduced with technology-first copy
or a frozen expectation of a changing system is a finding at tier
`[RESEARCH]` (cite the guidebook chapter).
