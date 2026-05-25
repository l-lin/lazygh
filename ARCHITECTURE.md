# Architecture

`lazygh` is a Go CLI with a lazygit-style TUI for GitHub pull request work. The repo has five main parts: startup wiring, provider-neutral domain models, a `gh` adapter, a persistent cache, and the TUI shell.

## System map

```text
cmd/lazygh
  -> internal/config      load config and startup mode
  -> internal/theme       resolve and apply palette
  -> internal/githubcli   build gh-backed adapters
  -> internal/tui         run the TUI

internal/tui
  -> internal/github      provider-neutral GitHub models
  -> internal/cache       persistent SQLite cache
  -> internal/story       story-review generation
  -> internal/theme       runtime colors
  -> internal/clipboard   clipboard access
```

## Package roles

- `cmd/lazygh`: composition root. It parses startup input, loads config, builds adapters, and starts the app.
- `internal/app`: tiny wrapper around the configured runner.
- `internal/config`: config loading, defaults, keymaps, cache paths, links, story-review settings, and pull-request searches.
- `internal/theme`: palette presets and exported runtime color variables.
- `internal/github`: provider-neutral GitHub models, enums, errors, and URL parsing.
- `internal/githubcli`: the only transport-aware GitHub adapter layer.
- `internal/cache`: SQLite persistence for pull requests, notifications, detail, and diff data.
- `internal/story`: prompt building, agent execution, and story-review parsing.
- `internal/clipboard`: clipboard interface and system implementation.
- `internal/tui`: state, rendering, async orchestration, and interaction handling.

## TUI shape

### Durable state

`internal/tui/model.go` holds the durable browser state: focus, tabs, list rows, search state, pane layout, and popup state.

Two child models now own the hot detail and review transitions:

- `detailStateModel` owns wrap width, cursor placement, viewport sync, and search sync.
- `reviewSessionState` owns file-tree selection and collapsed file or thread state.

### Shell

`internal/tui/program.go` defines `Program`, the shell object. It is grouped into four bundles instead of one flat bag:

- `programDeps`: injected ports and services
- `programStores`: caches, stores, and in-flight trackers
- `programViewRuntime`: promoted UI runtime state
- `programShellRuntime`: GUI, timers, async runner, and shell-only services

This keeps the composition root in one place, but it is still broader than a strict Elm shell.

### Loop

The TUI now has explicit `Msg`, `Update`, and `Cmd` types.

1. A keybinding, popup, editor, or async result emits a `Msg`.
2. `Update` mutates state and returns typed `Cmd` values.
3. `dispatch()` executes those commands.
4. `afterStateChange()` runs workflow planning, shell sync, and redraw.

`dispatchAsync()` is still the one production bridge that hops worker results back onto the UI thread.

## Read side and render side

The read side is much cleaner than it used to be.

- `ScreenState` describes logical browser, review, and story-review state.
- `ScreenLayout` turns that state plus terminal size into frames.
- `screenComposition` binds frames to renderers.
- `applyScreenComposition()` applies the result to `gocui`.

The render layer is mostly read-only now. Expensive document building and cache mutation live outside the render entrypoints.

The main read-only projection seams are:

- `footerPresenter`
- `helpPresenter`
- `actionsPopupPresenter`
- `searchViewPresenter`
- `reviewSessionReadModel`

Those snapshots keep footer, help, popup, title, and review-mode derivation off the full `*Program` bag.

## Command surfaces

Shell work now lives behind explicit command files.

- `workflow_commands.go`: detail, diff, cache hydration, notifications, and image loading
- `cmd_actions_popup_async_requests.go`: actions-popup async transport
- `cmd_modal_editor_submit_requests.go`: modal submit transport
- `cmd_popup_feature_request_requests.go`: popup feature transport
- `cmd_interaction.go`: popup reporting, clipboard work, link opening, detail-search follow-up, review-comment focus, page navigation, GUI reconfigure, and manual refresh registration
- `assignee_picker_search_cmd.go`: assignee search transport

These command files still live in `internal/tui`, but they now build focused runtime bundles at the `Cmd.execute(...)` boundary instead of passing the full shell bag deep into helpers.

## Workflow planning

`plannedWorkflow()` is the main post-update planner.

- `workflow_plan_selectors.go` reads live state and cache state.
- `workflow_plans.go` derives pure workflow plans.
- `workflow_stores.go` tracks in-flight and invalidation state.

The planner no longer flips store flags inline while deciding commands. Load starts and cache hydration now land through typed messages.

## Boundaries that matter

- GitHub shelling belongs in `internal/githubcli`.
- Provider-neutral models belong in `internal/github`.
- Palette values belong in `internal/theme`.
- Rendering belongs in `internal/tui`.
- Detail view `0` is read-only.
- Direct GitHub port calls in `internal/tui` are confined to `workflow_commands.go`.

## Current posture

The repo is TEA-shaped, not strict TEA.

What is already in good shape:

- explicit `Msg`, `Update`, and `Cmd`
- pure workflow planners
- child reducers for detail and review state
- read-only presenters and read models
- update files no longer show direct GUI or live-view coupling in the audit

What is still shell-heavy:

- `program_navigation.go` and `program_navigation_support.go`
- detail-view helper files such as `detail_bulk_fold.go`, `program_character_motion.go`, `yank_motion.go`, `review_inline_conversation.go`, `open_link.go`, `pull_request_build.go`, and `pull_request_reviewer.go`
- runtime entrypoints such as `view_url.go`, `editor_dispatch.go`, `actions_popup_async_success.go`, and `refresh_active_view.go`

## Audit snapshot, 2026-05-25

The post-refactor audit still shows these main gaps:

- `program_navigation.go`: 51 `func (program *Program)` methods, still the largest shell-heavy pocket
- direct shell hotspots by file: `program_navigation.go` (24), `detail_bulk_fold.go` (16), `shell_refresh.go` (11), `yank_motion.go` (10), `program_character_motion.go` (7), `review_inline_conversation.go` (6), `pull_request_build.go` (5), `open_link.go` (4), `pull_request_reviewer.go` (3)
- `update*.go` files no longer appear in the GUI or live-view coupling report
- direct GitHub ports in `internal/tui` remain confined to `workflow_commands.go`

That means the next migration work should stay focused on the remaining shell helpers, not on the reducer layer that was just cleaned up.

## Start here

If you need to rebuild the mental model fast, read these files in order:

- `cmd/lazygh/main.go`
- `internal/tui/program.go`
- `internal/tui/model.go`
- `internal/tui/update.go`
- `internal/tui/render_pipeline.go`
- `internal/tui/screen_state.go`
- `internal/tui/deps.go`
- `internal/tui/workflow_plans.go`
- `internal/tui/cmd_interaction.go`
- `internal/github/doc.go`
- `internal/githubcli/doc.go`
