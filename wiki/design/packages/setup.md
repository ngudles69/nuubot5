# Setup Package

Status: Implemented.
Covers: `internal/setup/setup.go`
Purpose: Return one fully admitted context before BtRunner composition.

## Canonical Source

- `D:/rust/nuubot4/src/setup.rs`

## Scope & Responsibilities

`nuubot_setup` loads root configuration, loads the identified Bot, resolves
owned paths, and enforces the shared-data boundary.

## Program Flow

```text
init
  resolve root
  load config
  load bot
  validate ticks path
  return setup
```

## Notes

- Setup performs admission only. It owns no running child.
- Setup returns one value, so Create, Start, Run, Loop, and Stop do not apply.
