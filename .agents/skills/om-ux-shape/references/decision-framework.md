# Decision Framework

Use this framework to move from a requested feature to a defensible product decision.

## 1. Frame the problem

Complete only the fields that affect the decision:

- **Actor:** Who is trying to make progress?
- **Situation:** When and where does the need occur?
- **Job:** What progress are they trying to make?
- **Current behavior:** How do they do it now?
- **Friction:** What blocks, slows, confuses, or increases risk?
- **Desired behavior:** What should become possible or easier?
- **Business effect:** Why does that behavior matter to the organization?
- **Constraints:** Time, technology, policy, accessibility, data, cost, or channel limits.

Compress the result into:

> For [actor] in [situation], the main obstacle to [job] is [friction]. We believe enabling [behavior] can create [user outcome] and contribute to [business effect].

Do not force this sentence when the evidence contradicts the requested feature.

## 2. Maintain an evidence ledger

Use four evidence classes:

| Class | Meaning | Treatment |
|---|---|---|
| Known | Directly supplied or observed | Use as a constraint or input |
| Inferred | Interpretation supported by known evidence | State the reasoning |
| Assumed | Necessary but unverified belief | Attach a validation path |
| Unknown | Missing information that may change the decision | Rank by risk |

Prefer behavioral evidence over stated preferences. Never present synthetic personas, imagined quotes, or generic market claims as findings.

## 3. Define outcomes

Define outcomes in this order:

1. **User outcome:** meaningful progress or reduced harm;
2. **Behavioral signal:** observable change that indicates progress;
3. **Business effect:** retention, revenue, cost, risk, adoption, quality, or strategic learning;
4. **Guardrail:** a metric or condition that must not degrade.

Avoid feature adoption as the only success measure. Usage can increase while the underlying job remains unsolved.

## 4. Assess risks

Rate each material risk as low, medium, or high and state why:

- **Value:** Will people choose this or benefit from it?
- **Usability:** Can they understand and complete the task?
- **Feasibility:** Can the team deliver acceptable quality, latency, and reliability?
- **Viability:** Does it work with the business model, policy, operations, brand, and cost?

For AI features, add model quality, harmful error, privacy, trust, and loss-of-control risks.

Tackle the highest combination of uncertainty and consequence first.

## 5. Compare directions

Generate two or three meaningfully different mechanisms, not cosmetic variants. Compare:

- fit with the diagnosis;
- user and business outcomes;
- strength of evidence;
- cognitive and operational complexity;
- risk and reversibility;
- time and cost to learn;
- fit with existing patterns and capabilities.

Recommend one direction. Use the smallest reversible commitment when evidence is weak.

## 6. Cut scope coherently

Keep a capability only when removing it prevents the primary job, destroys trust, or blocks learning.

Classify scope as:

- **Now:** required to complete and evaluate the primary job;
- **Later:** plausible value but not required for the first decision;
- **Not doing:** conflicts with focus, evidence, safety, or economics.

Do not omit error recovery, permissions, accessibility, or measurement merely because they are not visible in the happy path.

## 7. Match tests to assumptions

| Riskiest belief | Smallest useful test |
|---|---|
| The problem is real or frequent | Behavioral interview, observation, support/log review |
| The proposed value changes choice | Concept comparison, fake door, concierge test |
| People can understand the flow | Task-based prototype usability test |
| The technology can meet the bar | Technical spike or representative evaluation set |
| AI can add unique value | Wizard-of-Oz comparison against rules/manual baseline |
| The economics or policy work | Cost model, pricing test, legal/operations review |
| The released feature changes behavior | Instrumented pilot with guardrails |

Define before testing: expected signal, failure threshold, and the decision each result triggers.

