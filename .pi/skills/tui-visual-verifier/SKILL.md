---
name: tui-visual-verifier
description: Use when a change in this repo affects rendering, layout, themes, popups, cursor behavior, markdown, images, or other terminal-visible behavior that needs more than headless tests.
---

# TUI Visual Verifier

## Overview
Headless tests catch document and state logic. They do not prove the terminal looks right. Use this skill when the acceptance criteria are visual or TTY-specific.

## Use It For
- rendering or layout changes
- theme or palette updates
- popup, status-line, or prompt changes
- cursor, selection, search highlight, or fold behavior
- markdown, diff, or image rendering
- terminal protocol work such as kitty graphics

## Verification order
1. Add deterministic tests first:
   - render or document helpers
   - model or program tests
   - cache invalidation tests when theme or document caching changes
2. Run the relevant tests, then `mise run fmt`, `mise run test`, and `mise run build`, unless the task is strictly visual triage.
3. Do one live TTY check when feasible. **REQUIRED SUB-SKILL:** Use `tmux` when the flow needs a persistent terminal, real keypresses, or concurrent inspection.
4. Verify the exact user-visible surface that changed:
   - selected row and active frame
   - search highlight versus selection highlight
   - popup cursor, scrolling, and footer or status hints
   - theme switch and cache refresh
   - fold or conversation toggle state
   - markdown or diff layout
5. For image or graphics protocol work:
   - verify text fallback separately from pixel rendering
   - do not claim pixel verification from `tmux capture-pane`
   - prefer cropped or downscaled screenshots when attachments are too large
6. If the bug only reproduces live, write a handoff note in `.sandbox/handoffs/` with the failing example, the current hypothesis, and what was actually verified.

## Stop signs
- `Tests passed` is not enough for a visual acceptance change.
- `I checked capture-pane` is not enough for kitty or other pixel graphics.
- `The headless render looks right` is not enough for cursor, focus, or overlay bugs.
