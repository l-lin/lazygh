---
name: pattern-hybrid-tea-modal-adapter
description: Use when refactoring an interactive modal component whose semantic state can be reduced but whose host UI, inputs, timers, and callbacks remain imperative.
disable-model-invocation: false
---

# Hybrid TEA modal adapter

Keep the host integration imperative and make the modal's semantic transitions explicit.

## Structure

- Store cursor, mode, pending commands, annotations, searches, viewport state, and async-operation bookkeeping in one tagged model.
- Let `update(model, message, context)` perform state transitions and return ordered effects.
- Keep `Component`, `Input`, clipboard, timers, callbacks, and render requests in a thin adapter.
- Put display-row geometry and width-dependent reflow in separate pure helpers.
- Let the view consume model state and return rendered lines plus derived viewport state without calling TUI methods.

## Invariants

- Use discriminated unions for mutually exclusive interactions such as normal, visual, comment, and search input.
- Keep pending command variants distinct when their discriminants overlap, for example by wrapping motion and yank pending states.
- Preserve async race protection with both an operation ID and a session ID. Ignore completions from disposed or refreshed sessions.
- Keep effects ordered when host callbacks have ordering semantics, such as saving text before closing an overlay.
- Snapshot cursor, selections, annotations, search state, and yank highlights by visible-display ordinal before width-dependent reflow, then restore them against the new document.

## Pi TUI focus

A container implementing `Focusable` must propagate its focus setter to whichever `Input` is currently open. Set the child input's focus when opening it and again whenever the container's focus changes; this keeps IME cursor placement correct.

## Verification

Run the focused tests after each extraction, then the full test command and typecheck. Add regression coverage for async completion ordering, width reflow with active decorations, and focus changes after opening embedded inputs.
