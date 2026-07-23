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
nuubot_setup(log, sweep_id, bot_id) -> ctx

init
  config = Config(repository/config.toml)
  bot    = SweepStore(config.sweep_database).load(sweep_id, bot_id)
  verify bot.ticks_path is inside config.shared_data
  return ctx(config, bot)
```

## Notes

- Setup performs admission only. It owns no running child.
