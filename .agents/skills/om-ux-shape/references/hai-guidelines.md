# Human-AI interaction checklist — the 18 guidelines, by moment of use

Companion to ai-interaction.md: the necessity gate decides WHETHER and at
what level AI belongs; this checklist verifies the DESIGN of whatever
passed the gate. Source: Microsoft's Guidelines for Human-AI Interaction
(Amershi et al., CHI 2019; HAX Toolkit) — a synthesis of 20+ years of
research, here condensed for review use. Check the phase that matches the
feature; not every guideline applies to every product.

## At the start (what the user sees before trusting it)

1. **Make clear what the system can do.**
2. **Make clear how well it does it** — set expectations about quality,
   not just capability.

## During interaction

3. **Time services based on context** — act when the user needs it, not
   when the system finds it convenient.
4. **Show contextually relevant information** — relevant to the current
   task, not everything the model knows.
5. **Match relevant social norms** — tone and formality fit the context.
6. **Mitigate social biases** — outputs do not reinforce stereotypes.

## When the system is wrong (design failure before polish)

7. **Support efficient invocation** — easy to ask for the AI when wanted.
8. **Support efficient dismissal** — easy to ignore or hide when unwanted.
9. **Support efficient correction** — fixing a wrong result costs less
   than living with it.
10. **Scope services when in doubt** — degrade gracefully or ask, instead
    of guessing confidently.
11. **Make clear why the system did what it did** — explanation where the
    explanation can change the user's decision.

## Over time

12. **Remember recent interactions** — context carries forward.
13. **Learn from user behavior** — personalize from what users do.
14. **Update and adapt cautiously** — no rug-pulls of learned behavior.
15. **Encourage granular feedback** — feedback the product can act on.
16. **Convey the consequences of user actions** — what feedback and
    settings will change.
17. **Provide global controls** — users can steer what is monitored and
    how the system behaves.
18. **Notify users about changes** — capability changes are announced.

## How to use in this skill

- In Shape mode: guidelines 1-2 and 7-11 shape the interaction contract's
  capability framing, control and failure rows.
- In Review mode: walk the four phases in order; a violated guideline is
  a finding at tier `[RESEARCH]` (cite the guideline number and name).
- Remember the writing rule: the guideline names are for the author —
  the delivered report describes what the user sees and what to change.
