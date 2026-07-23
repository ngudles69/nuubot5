# Nuubot5 Project Instructions

## Startup

At the start of every root session:

1. Read `HANDOFF.md`.
2. Read `wiki/PROJECT.md`.
3. Read `wiki/USER.md`. This contract is prescriptive, not a suggestion.
4. Read `wiki/SOUL.md`. This contract is prescriptive, not a suggestion.
5. Read `wiki/ARCHITECTURE.md`.
6. Read `wiki/DESIGN.md`.
7. If `TMUX` is set, read `TMUX.md` and follow its startup contract.

Read-only startup commands do not require confirmation. Do not invent work when
the user has not requested any.

The `TMUX.md` workspace bootstrap is approved startup work. It does not require
separate pre-action confirmation.

Report current state and next action in a `READY.` response.

## Before Action

Before editing, running task commands, installing software, or taking external
action:

1. Restate the user's intent.
2. State assumptions and unresolved choices.
3. Name the canonical owner and affected files.
4. State the expected outcome and exact proof.
5. Wait for explicit user confirmation.

Discussion and read-only startup are not actions. Investigate before asking a
question that local evidence can answer.

## Before Coding

Before coding:

1. Read `wiki/coding/STYLE.md`.
2. Read `wiki/coding/RULES.md`.

## Execution

- Read the owning wiki page and trace the real flow before editing.
- Make one coherent scoped change.
- Follow `wiki/coding/STYLE.md` and `wiki/coding/RULES.md`.
- Keep durable design in `wiki/**` and current restart state in `HANDOFF.md`.
- Update the owning `wiki/design/**` page when implementation proves a new design fact.
- Never commit or push without explicit user authority.
- Report confirmed facts separately from inference.

## Orchestration Continuity

- The root MUST NEVER stop while authorized `TODO` items remain.
- The root agent owns continuity, sequencing, delegation, verification, and final reporting. It MUST keep work moving without waiting for the user to manage it.
- The root MUST maintain visible `DONE`, `TODO`, and `PENDING USER APPROVAL` lists during active work.
- The root MUST update all three lists when work changes.
- Empty lists MUST show `None`.
- The root MUST never require the user to reconstruct or manage task state.
- The root MUST continuously optimize task order, dependencies, and safe parallelism.
- After each completion, failure, or user message, the root MUST continue or delegate the next authorized `TODO` immediately.
- Stopping is permitted only when every `TODO` is `DONE`, genuinely blocked, or moved to `PENDING USER APPROVAL`.
- The root MUST NOT serialize independent work without a stated reason.
- `HANDOFF.md` MUST record the active task immediately.
- `HANDOFF.md` MUST update proof when work completes. It is not only a closeout report.
- Verification MUST be proportional to change risk.
- Deterministic mechanical fixes MUST close with direct proof, without another reviewer.
- Re-audit MUST occur only when behavior, ownership, contracts, or unresolved judgment changed.

## Prose Contract

This contract applies to chat, plans, reports, wiki, handoff, comments, and
agent prompts.

- Use caveman prose.
- Lead with the result.
- Use compact prose with a maximum of 25 words per point.
- Use bullets only for a genuine list.
- Cut filler, repetition, and repeated context.
- Do not write walls of text.
- Expand only when the user asks.

Code, commands, paths, logs, and exact quotations are exempt.

## Key Decisions

These decisions apply project-wide. Local convenience MUST NOT override them.

### ALWAYS USE

- Use `-tags noasm`; the optimized decoder produced one corrupt timestamp at
  run 183.
- When the user requests an `Nx` or `Nx test`, run
  `./rtest.sh N 6 9`: N runs, Sweep ID 6, Bot ID 9. Report pass/fail,
  attempted runs, total and average duration, replay timing, and the result
  log path.
- Use the simplest idiomatic Go that satisfies the current requirement.
- Use the Go standard library before external or custom code.
- Use an approved pure-Go library before writing custom mechanical code.
- Standard Go build tags, including canonical `noasm`, remain allowed.
- Keep the explicitly approved `nuubot-server`, `nuubot-cli`, and
  `nuubot-runner` command shells as naming placeholders that print
  `Under Construction.` until their real implementation is authorized.
- Keep explicitly approved package reservation files limited to one package
  comment and declaration until implementation is authorized.

### DO NOT USE

- Do not use CGO or native C/C++ bindings without explicit prior user approval.
- Do not use `unsafe`, handwritten assembly, or `//go:linkname` without explicit
  prior user approval.
- Do not use runtime internals, compiler internals, Go plugins, or dynamically
  loaded native libraries without explicit prior user approval.
- Do not use non-standard compilers or toolchains without explicit prior user
  approval.
- Do not treat pre-contract source drift as precedent.
