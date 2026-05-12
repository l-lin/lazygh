# TUI panel, view, and overlay contracts

Frozen by TODO 46. Later refactors may replace internals, but they must keep these user-facing rules until another tested TODO changes them on purpose.

## Numbered views

### Browser mode
- `view 0` is the main panel.
- `view 1` is the connected-user side view.
- `view 2` is the pull-requests side view.
- `view 3` is the notifications side view.

### Review and story review
- Side-panel content may change.
- `view 0` stays the main panel.
- The main panel still shows the content driven by the active side view and its selected row.

## Navigation
- `h` and `Shift+Tab` move left across side views.
- `l` and `Tab` move right across side views.
- `0`, `1`, `2`, and `3` jump to the matching numbered view.
- `[` and `]` switch tabs only inside the active view that exposes tabs.

## Cursor and selection
- Only `view 0` shows the main-pane cursor.
- Side views show rows and selected rows, not the main cursor.
- Overlays may show their own input cursor when they own focus.

## Overlays and hints
- Actions open in framed popup chrome.
- Text editing opens in the framed input-box popup.
- Contextual key hints stay in the bottom-right status area.

## Safety-net tests
- Model and program contracts live in `internal/tui/panel_view_contract_model_test.go`.
- Render, overlay, review, and story-review contracts live in `internal/tui/panel_view_contract_render_test.go` and `internal/tui/panel_view_contract_review_test.go`.
- The live TTY smoke harness lives in `internal/tui/panel_view_contract_manualvisual_test.go`.

## Live TTY smoke check
Run the manual harness in a tmux pane, wait for the ready token, drive the UI with `0`, `1`, `2`, `3`, `Tab`, `Shift+Tab`, and `a`, then release the done token.

```sh
LAZYGH_TMUX_READY_TOKEN=lazygh-ready \
LAZYGH_TMUX_DONE_TOKEN=lazygh-done \
  go test -tags manualvisual ./internal/tui -run TestManualVisual_PanelViewContracts -count=1
```
