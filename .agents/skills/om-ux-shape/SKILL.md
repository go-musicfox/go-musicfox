---
name: om-ux-shape
description: Turn a vague product, UI/UX, or AI feature idea into a decided direction. Use when shaping a feature, simplifying an overcomplicated flow, deciding whether and how to use AI, defining screen states, planning validation, or preparing a design handoff for engineering.
---

# Shape Useful Features

Turn ambiguity into a clear product decision before turning it into screens.
Connect user value, business value, interaction quality, AI behavior, delivery
constraints, and evidence in one lightweight process.

**Input** — a feature idea, an existing concept or product area, or a decided
direction that needs implementation detail.
**Output** — one filled shape from `references/report-templates.md`.

## Choose the mode

- **Shape** for a vague opportunity, request, or feature idea. The default.
- **Review** for an existing concept, flow, design, prototype, or product area
  (including a whole module handed over by `om-ux-review-pr`).
- **Handoff** when the direction is decided and implementation-ready behavior
  is what is missing.

Combine modes only when the request genuinely spans them, and never make a
small task carry the full process.

## Who reads the result

Establish this before writing, because it decides how concrete the output must
be. Ask when it is unclear, otherwise **default to the least-context reader**:
someone who will build or draw this, does not know the design system, and was
not in the conversation. Write for them. A senior designer can skim a concrete
answer; nobody can build an abstract one.

## Required reading

These are not optional background. When the condition is met, load the
reference before finishing the step: skipping it produces the failure this
skill exists to prevent, an answer that sounds reasonable and decides nothing.

| Condition | Load |
|---|---|
| Shape or Review mode | `references/decision-framework.md` |
| AI is proposed, implied, or already present | `references/ai-interaction.md`, then `references/hai-guidelines.md` and `references/reward-and-mental-models.md` for whatever passed the gate |
| Outcomes or validation metrics are being defined | `references/human-value-metrics.md` |
| Before writing any result | `references/report-templates.md` |
| Before delivering any result | `references/quality-rubric.md` |

`references/foundations.md` explains the rationale behind the process; read it
only when adapting the process or evolving this skill.

## Operating principles

1. Start from the consequential problem, not the requested interface.
2. Treat requirements as claims until evidence supports them.
3. Label facts, inferences, assumptions, and open questions. Never invent
   research, user quotes, metrics, or constraints.
4. Tie the user outcome to a business effect without treating business value
   as a substitute for user value.
5. Prefer the smallest coherent end-to-end solution over a collection of
   features.
6. Recommend a direction. Do not hide behind an unranked menu of options.
7. Make every UI element earn its place by enabling an action, decision,
   status, explanation, or recovery.
8. Treat AI as a design material with uncertainty, latency, cost, and failure
   modes, not as a default interface.
9. Preserve meaningful human control, especially for consequential or
   hard-to-reverse actions.
10. Match the depth of the process and output to the decision's risk.

## Workflow

0. **Agentic setup** — follow `references/agentic-setup.md`: repo-local
   override contract, the design contract as constraints when present, and the
   untrusted-content boundary. Shared communication and reporting rules live
   in `references/rules.md`.

1. **Establish the decision.** State the decision being made, the primary
   actor and situation, the intended user and business outcomes, and the mode
   and depth. Ask only questions whose answers could materially change the
   direction; otherwise proceed with clearly marked assumptions.

2. **Build the evidence ledger.** Separate what is known, inferred, assumed,
   and unknown, following `references/decision-framework.md` (§2). Prioritize
   unknowns by decision risk, not curiosity.

3. **Diagnose.** Write a one-sentence diagnosis naming the main obstacle to
   progress, distinguishing the underlying job from the requested feature.
   Then define one primary behavioral outcome, its plausible business effect,
   and a guardrail against harmful optimization. Framing and outcome
   discipline: `references/decision-framework.md` (§1, §3); choosing signals
   that mean people are better off: `references/human-value-metrics.md`.

4. **Test the proposed mechanism.** When AI is involved, run the necessity
   gate in `references/ai-interaction.md` and explicitly consider a rules-based
   alternative; a design that passes the gate is then checked against
   `references/hai-guidelines.md`, and its preferred-mistake decision and
   first-contact framing against `references/reward-and-mental-models.md`.
   For any feature, rate the four product risks (value, usability,
   feasibility, viability) per `references/decision-framework.md` (§4), adding
   trust, safety, privacy, and model-quality risks for AI.

5. **Choose a direction.** Generate two or three meaningfully different
   mechanisms, compare them per `references/decision-framework.md` (§5), and
   select one, explaining the decisive trade-off. Tag the claims that carry
   the argument with their honest tier from `references/evidence-tiers.md`. In
   Review mode, rank findings by impact × frequency × reach, never by ease of
   fix.

6. **Shape the smallest coherent feature.** One primary job and happy path,
   the minimum states and recovery paths trust requires, and an explicit now,
   later, and not-doing split (`references/decision-framework.md` §6). A thin
   but broken slice is not an MVP: the smallest coherent feature completes a
   real job end to end and survives its likely failures.

7. **Specify the interaction contract, concretely.** Entry point and trigger,
   information required, system response, primary decisions and actions, the
   relevant empty, loading, partial, success, error, and permission states,
   and the edit, undo, dismiss, retry, fallback, or escalation paths, plus
   accessibility and content requirements. For AI, also specify capability
   framing, uncertainty, explanations, data use, feedback, control, and
   behavior when the model cannot help.

   Concrete means: name the screens, name the components (from the contract
   registry when one exists), and write the actual labels, headings, empty-state
   sentences, and error messages rather than describing them. "Add a helpful
   empty state" is unfinished work; the finished version says what the screen
   shows, in the words the user will read. **If a reader could not build or
   draw it from your output, the step is not done.**

8. **De-risk and deliver.** Name the riskiest unverified belief, choose the
   smallest test that could change the decision, and state what each result
   triggers (`references/decision-framework.md` §7). Then fill the matching
   shape in `references/report-templates.md`, and apply
   `references/quality-rubric.md` before delivering: a zero in diagnosis, user
   outcome, coherent scope, AI necessity, or AI control and recovery means the
   result is not ready. Close with the one-line **Applied** note the template
   defines, so the reader can see which checks ran and which did not apply
   instead of guessing what was considered.

## Response behavior

- Lead with the recommendation or verdict.
- Use plain language and concrete product behavior; keep process narration
  shorter than the decision it supports.
- Scale detail down for low-risk work and up for consequential, novel, or
  implementation-ready work.
- If the evidence does not support a confident recommendation, say what is
  provisional and propose the smallest learning step.
- If the user asks to build the feature, use this workflow to decide, then
  continue into implementation. The Handoff shape is written to feed the
  collection's implementing skills; `om-ux-review-pr` closes the loop on the
  resulting PR.
