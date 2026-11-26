# Conversation guide

How `om-brainstorm` runs the exploration (workflow steps 1–2 and the convergence in step 4). The hard rules live in the skill body — this file is technique.

## Sources before questions

The ladder, in order: the repository (code, agent instruction files) → docs and existing specs → the tracker (step 3, read-only, when available) → the user. The user is asked directly only what has no other source: motivation, priorities, appetite, constraints, taste. A question whose answer sits in the repo is homework, not conversation.

## Question discipline — one at a time

- Ask one open question, listen, follow the answer. The next question comes from what was said, not from a script.
- Batch only trivially closed questions (binary or multiple-choice with an obvious option set), and only when they are genuinely independent of each other.
- Never present a numbered wall of questions — that is a form, not a conversation.
- When the user answers with a solution, ask about the problem it solves before adopting it.

## Divergence moves

- Restate the problem without the user's proposed solution in it. If the problem disappears, the proposal was the problem.
- Generate two or three genuinely different alternatives, plus an explicit "do nothing / build nothing" option priced with its real consequences.
- Name the riskiest assumption in the current favorite, and the cheapest way to test it.
- Kill vague framing — a goal that cannot be stated specifically is not ready to route. The bar: "admin manages team" is vague; "admin invites by email and assigns a role; cannot delete users" is ready.
- Run the bundling test early: does the idea contain more than one independently deployable capability (would each function without the other)? A bundle routes as its first slice; the rest become their own briefs or parked issues — never one giant brief.
- Ask what breaks or gets worse if this succeeds (second-order effects: load, support, workflow collisions).

## Convergence checklist

Stop exploring when all three hold:

- The conclusion type is identifiable — one exit-ramp row fits better than the others.
- The unknowns that block routing are resolved; they become the brief's Resolved-unknowns table.
- The user signals enough — depth of exploration is their call, not the skill's.

Then run the challenger gate (workflow step 4) before presenting the conclusion as final.
