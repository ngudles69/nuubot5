# TMUX Workspace

This file owns the Nuubot5 TMUX and PSMux startup workspace.

PSMux exports `TMUX`. The same startup gate applies to both tools.

## Order

Run this process after the six normal `AGENTS.md` startup reads and before
reporting `READY.`.

Normal startup remains unchanged when `TMUX` is not set.

## Viewer startup

The first request may assign one exact viewer role:

- planner;
- plan auditor;
- coder; or
- implementation auditor.

That request starts a viewer, not the control agent.

The viewer MUST NOT inspect, close, split, launch, or focus panes.

The viewer MUST enter standby. It MUST NOT plan, audit, code, delegate, edit, or
run task commands before a separate task.

The viewer MUST reply:

```text
READY.

DONE

- Startup contracts loaded.
- Role active: <role>.

TODO

- None.

PENDING USER APPROVAL

- First <role> task.

Next: waiting for your task.
```

Only `<role>` changes.

## Control gate

Use the native TMUX or PSMux command for every operation.

1. Read the current pane identity and list every pane in the current window.
2. Treat the first pane position as control. Never require literal pane ID `%1`.
3. Continue only when the current pane is first or is the only pane.
4. Otherwise report the pane list, fail startup, and stop.
5. When siblings exist, close every sibling by explicit pane ID.
6. Verify the control pane is the only remaining pane.
7. If any command or check fails, report the failed step and stop.

Closing siblings is approved startup reset work. Do not request confirmation.

## Layout

Record every pane ID from live command output. Never predict or hard-code IDs.

Build this layout:

```text
+----------------------+-------------+-------------+
|                      | planner     | plan auditor|
| control              +-------------+-------------+
|                      | coder       | impl auditor|
+----------------------+-------------+-------------+
```

1. Split control left/right at 50 percent.
2. Split the right pane top/bottom at 50 percent.
3. Split both right panes left/right at 50 percent.
4. Verify one full-height control pane and four viewer panes.
5. Fail and stop when the layout or pane count differs.

Viewer order is:

1. top-left: planner;
2. top-right: plan auditor;
3. bottom-left: coder; and
4. bottom-right: implementation auditor.

## Launch

For each viewer in order:

1. send `codex`;
2. send `Enter`;
3. return focus to control;
4. wait until the Codex composer is ready;
5. send:

```text
For this session, you are acting as an interactive <role> agent.
This role assignment is standby only.
Do not manage the TMUX or PSMux workspace.
Do not plan, audit, code, delegate, edit, or run task commands before a separate task.
Reply exactly:

READY.

DONE

- Startup contracts loaded.
- Role active: <role>.

TODO

- None.

PENDING USER APPROVAL

- First <role> task.

Next: waiting for your task.
```

6. send `Enter`;
7. return focus to control.

Wrapped prompt text is valid.

## Proof

Before control reports `READY.`:

- all five panes exist;
- control is first, full-height, and active;
- the right side is a two-by-two grid;
- all four Codex viewers returned the required response; and
- every command failure is reported without fallback or repair.
