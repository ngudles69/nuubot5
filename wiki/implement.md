# Implementation Workflow

Status: Active project workflow.
Purpose: Preserve user intent through design, audit, implementation, proof, and
closeout.

## Intent

Far-reaching work must not jump directly from conversation into code.

The user's intent must remain visible and independently reviewable.

Design review and implementation review answer different questions.

Design blockers must be cleared before implementation.

Implementation blockers must be cleared before closeout.

The final summary must report the work, proof, audits, fixes, and remaining
limits.

This workflow may later become a Codex skill.

It remains a project wiki contract until that decision.

## When To Use

Use this workflow for:

- new architecture;
- cross-package behavior;
- lifecycle changes;
- persistence changes;
- new durable data contracts;
- trading behavior;
- timing-sensitive behavior; and
- proof-heavy implementation.

Trivial mechanical changes do not require this complete workflow.

## Step 1 — Record User Intent

Write the user's intent before proposed design.

Preserve:

- desired outcome;
- ownership expectations;
- removal expectations;
- lifecycle;
- persistence;
- downstream consumers;
- future direction;
- explicit constraints; and
- known deferred work.

Do not rewrite intent into implementation assumptions.

Separate confirmed intent from proposed design.

## Step 2 — Write Proposed Design

Write the smallest coherent design satisfying the recorded intent.

Define:

- canonical owner;
- object boundaries;
- data flow;
- lifecycle ordering;
- mutation boundaries;
- persistence;
- failure behavior;
- removal path;
- initial scope;
- deferred scope; and
- exact proof.

Durable design belongs in the owning `wiki/**` page.

The first implementation may remain intentionally small.

The design must allow additive growth without speculative frameworks.

## Step 3 — Adversarial Design Audit

Spawn one independent read-only adversarial reviewer.

The reviewer receives:

- original user intent;
- proposed design;
- owning project contracts;
- relevant current implementation;
- known constraints; and
- planned proof.

The reviewer checks:

- intent adherence;
- internal coherence;
- ownership;
- hidden coupling;
- loopholes;
- missing states;
- lifecycle ordering;
- persistence boundaries;
- failure paths;
- downstream effects;
- sequencing;
- removal claims;
- unnecessary abstraction; and
- proof completeness.

The reviewer does not edit files.

## Step 4 — Clear Design Blockers

Triage every material design finding.

Accept or reject each finding with evidence.

Fix accepted blockers in the owning design.

Formatting and prose-only issues receive direct fixes.

Do not send trivial fixes through another audit.

Re-audit only when fixes materially change ownership, behavior, contracts, or
unresolved judgment.

Implementation begins only when no material design blocker remains.

## Step 5 — Implement

Implement the audited initial scope.

Trace every changed line to user intent or accepted design.

Keep the change detachable where required.

Do not add deferred features.

Update owning design pages when implementation proves a new fact.

Maintain current work and proof in `HANDOFF.md`.

## Step 6 — Prove

Run proof proportional to risk.

Proof may include:

- focused tests;
- full tests;
- static analysis;
- deterministic replay;
- database integrity;
- state and lifecycle assertions;
- timing and memory;
- stability runs;
- real data;
- artifact queries; and
- removal or buildability checks.

Record exact commands, outcomes, and evidence paths.

Do not claim success from logs without checking their measured boundaries.

## Step 7 — Adversarial Implementation Audit

Spawn one independent read-only adversarial reviewer.

The reviewer compares:

- user intent;
- audited design;
- exact implementation diff;
- tests;
- runtime proof;
- persisted evidence; and
- known limitations.

The reviewer checks correctness, reachability, lifecycle, persistence, timing,
downstream behavior, missing proof, duplication, and bloat.

The reviewer does not edit files.

## Step 8 — Clear Implementation Blockers

Triage every material implementation finding.

Fix accepted blockers at the owning invariant.

Rerun focused and affected full proof.

Do not re-audit trivial fixes.

Re-audit material behavior, ownership, or contract changes when project rules
require it.

Do not close with known material blockers.

## Step 9 — Final Summary

Report:

- implemented outcome;
- affected owners;
- design audit result;
- accepted and rejected findings;
- implementation audit result;
- fixes;
- proof;
- timing and memory;
- artifacts;
- known limits;
- deferred work; and
- required user approval.

Separate confirmed facts from inference.

Do not require the user to reconstruct task history.

## Required State

During active work, maintain:

- `DONE`;
- `TODO`; and
- `PENDING USER APPROVAL`.

The root owner continues authorized work until every TODO is complete, blocked,
or awaiting user approval.

## Current Example

Telemetry and RunReport use this workflow:

```text
record user intent
write telemetry and RunReport designs
adversarially audit design
clear design blockers
implement audited initial scope
run proof
adversarially audit implementation
clear implementation blockers
rerun proof
summarize
```

