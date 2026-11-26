# AI Interaction Framework

Apply this reference whenever AI is requested, implied, or already part of the product.

## Contents

- [Run the AI necessity gate](#1-run-the-ai-necessity-gate)
- [Choose automation or augmentation](#2-choose-automation-or-augmentation)
- [Define the AI contract](#3-define-the-ai-contract)
- [Design trust calibration](#4-design-trust-calibration)
- [Design failure before polish](#5-design-failure-before-polish)
- [Cover relevant states](#6-cover-relevant-states)
- [Measure user value](#7-measure-user-value-not-model-theater)
- [Avoid default AI patterns](#avoid-default-ai-patterns)

## 1. Run the AI necessity gate

Ask:

1. What user problem exists independently of AI?
2. What unique capability does AI add: prediction, classification, generation, personalization, recognition, or natural-language interpretation?
3. Would a deterministic rule, search, template, or conventional UI solve it more clearly and cheaply?
4. What quality level is required for the feature to remain useful?
5. What happens when the AI is confidently wrong, slow, unavailable, or too expensive?
6. Can the user recognize a poor result and recover?
7. Are the required data available, permitted, representative, and understandable?

Choose explicitly:

- **No AI:** a conventional solution is clearer or safer;
- **Rules first:** deterministic behavior covers the core need;
- **AI-assisted:** AI proposes; the user reviews or edits;
- **AI-collaborative:** user and AI iteratively shape the result;
- **AI-automated:** AI acts within bounded, observable, reversible authority.

Do not use AI merely to make a familiar interaction feel novel.

## 2. Choose automation or augmentation

Prefer automation when the task is repetitive, unpleasant, low-stakes, objectively checkable, and easy to reverse.

Prefer augmentation when the task is subjective, creative, identity-bearing, high-stakes, socially consequential, hard to specify, or one for which the user remains responsible.

Increase human oversight as consequence, uncertainty, or irreversibility increases. For consequential actions, require preview, explicit approval, an audit trail, and a fallback.

## 3. Define the AI contract

Specify:

- **Input:** user-provided and contextual data;
- **Output:** what the AI returns and in what structure;
- **Quality bar:** acceptable, unacceptable, and ambiguous output;
- **Latency:** immediate, progressive, background, or asynchronous;
- **Authority:** suggest, draft, rank, decide, or act;
- **Explanation:** what the user needs to understand and when;
- **Control:** edit, regenerate, constrain, compare, approve, undo, disable, or escalate;
- **Learning:** what feedback or behavior changes future results;
- **Fallback:** what remains possible without the AI.

Expose capability and limitations at the moment they matter. Do not front-load a technical lecture.

## 4. Design trust calibration

Aim for appropriate trust, not maximum trust.

- State what the system can do and how reliably it can do it.
- Make important data sources and missing context visible.
- Explain a result when the explanation can change a decision.
- Communicate uncertainty in user language, not decorative confidence scores.
- Provide sources or evidence when factual verification matters.
- Let the user correct the system efficiently.
- Notify users when behavior, personalization, or data use changes.
- Avoid anthropomorphism that implies human understanding, intention, or accountability.

Never use confident copy to conceal probabilistic behavior.

## 5. Design failure before polish

Cover relevant failure types:

- **Input failure:** insufficient, invalid, unsafe, or ambiguous input;
- **Model failure:** low-quality, unsupported, inconsistent, or unavailable output;
- **Context failure:** technically plausible output based on a wrong assumption about the user or situation;
- **Action failure:** the system cannot complete or reverse the intended action;
- **Silent failure:** neither the user nor system immediately recognizes a harmful error.

For each likely failure, define:

1. how it is detected;
2. what the user sees;
3. what remains under user control;
4. the fastest path forward;
5. what the team can learn from it.

Do not use “try again” as the only recovery when the user can provide better input, narrow scope, switch methods, or continue manually.

## 6. Cover relevant states

Consider:

- AI unavailable or not configured;
- insufficient context or permissions;
- ready for input;
- generating, searching, or acting;
- partial or streamed output;
- successful output awaiting review;
- low-confidence or conflicting output;
- user correction or regeneration;
- refusal, safety boundary, or unsupported request;
- service, quota, latency, or cost failure;
- manual fallback and recovery.

Include only states that can plausibly occur, but do not design only the ideal response.

## 7. Measure user value, not model theater

Combine:

- task completion and quality;
- time or effort saved;
- correction, override, and undo rate;
- abandonment after AI output;
- successful recovery from failure;
- harmful or high-cost error rate;
- repeat use for the underlying job;
- latency and cost per successful outcome.

Acceptance rate alone is weak: people may accept a suggestion because checking it is harder than trusting it.

## Avoid default AI patterns

- Do not default every feature to an empty chat box.
- Do not require users to become prompt engineers for routine tasks.
- Do not hide deterministic actions behind conversational ambiguity.
- Do not generate broad output when a constrained choice is safer.
- Do not request feedback the product cannot interpret or use.
- Do not personalize invisibly or make irreversible changes silently.
- Do not remove the non-AI path unless evidence supports doing so.
