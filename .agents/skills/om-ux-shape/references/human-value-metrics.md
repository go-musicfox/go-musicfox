# Human-value metrics — choosing signals that mean people are better off

Use when defining outcomes (workflow step 3) and validation. Based on
Google's HEART framework (Rodden, Hutchinson, Fu), adapted to this
collection's rule: measure value delivered TO people, and treat every
machine-friendly proxy with suspicion.

## The five dimensions

- **Happiness** — attitude: satisfaction, perceived usefulness, ease.
  Measured by asking (surveys, ratings), so it lags and gets gamed by
  prompt timing; never the only dimension.
- **Engagement** — depth and frequency of voluntary use.
- **Adoption** — new users or teams actually starting to use the thing.
- **Retention** — do they come back; churn reduced.
- **Task success** — completion rate, time, error rate on the primary job.

## Goals → signals → metrics

For each chosen dimension, write the chain explicitly:

1. **Goal** — what better looks like for the person ("operators add a
   person without help").
2. **Signal** — observable behavior that would indicate it ("creation
   completed without abandoning the form").
3. **Metric** — the number ("form abandonment rate under 10%").

A metric without its goal and signal is a number waiting to be gamed.

## Rules of use

- Pick 1-2 dimensions per feature, not five. Tools lean on Task success;
  habitual products add Retention; measure Engagement only where MORE use
  genuinely means MORE value delivered.
- **Engagement is the most abusable dimension**: session length grows when
  users are lost, addicted, or blocked just as well as when they are
  served. Pair any engagement metric with a task-success guardrail, and
  cross-check against the humane gate — a metric fed by nagging, streak
  guilt or fake urgency measures extraction, not value.
- Feature adoption alone is a vanity outcome: usage can rise while the
  underlying job stays unsolved (this echoes the outcomes rule in the
  decision framework).
- Always define at least one **guardrail metric** that must not degrade
  (support tickets, correction rate, time on the OLD path for users who
  avoided the new one).
- For AI features, wire this to the value measures in ai-interaction.md
  §7: correction/override rate and successful recovery from failure are
  task-success signals, not noise.
