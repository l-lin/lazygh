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

1. A keybinding, popup, editor-intent callback, or async result emits a `Msg`.
2. `Update` mutates state and returns typed `Cmd` values.
   Pull-request list hydrate/load messages also normalize opened-summary insertion and durable pinning there instead of hiding it in loader helpers.
3. `dispatch()` executes those commands.
4. `afterStateChange()` runs workflow planning, shell sync, and redraw.

`dispatchAsyncMessage()` is the shell bridge that hops worker results back onto the UI thread.

## Read side and render side

The read side is much cleaner than it used to be.

- `ScreenState` describes logical browser, review, and story-review state.
- `ScreenLayout` turns that state plus terminal size into frames.
- `screenComposition` binds frames to renderers.
- `applyScreenComposition()` applies the result to `gocui`.

The render layer is mostly read-only now. Expensive document building and cache mutation live outside the render entrypoints.

`refreshViews()` now runs with a short-lived read cache so footer and popup presenters, popup action lists, keybinding label resolution, and review-session read models are computed once per redraw instead of several times.

The main read-only projection seams are:

- `footerPresenter`
- `helpPresenter`
- `actionsPopupPresenter`
- `searchViewPresenter`
- `reviewSessionReadModel`
- `review_session_selectors.go`
- `detail_cursor_selectors.go`

Those snapshots keep footer, help, popup, title, and hot review/story selection derivation off the full `*Program` bag.

## Command surfaces

Shell work now lives behind explicit command files.

- `workflow_session_commands.go`: connected-user load
- `workflow_pull_request_list_commands.go`: pull-request list load, reload, and cache hydration
- `workflow_pull_request_detail_commands.go`: pull-request detail and diff load, cache hydration, and diff team-owner enrichment
- `workflow_notification_commands.go`: notifications plus issue and release detail loads
- `workflow_detail_image_commands.go`: markdown HTML and detail-image loads
- `cmd_actions_popup_async_requests.go`: actions-popup async transport
- `cmd_modal_editor_submit_requests.go`: modal submit transport
- `cmd_popup_feature_request_requests.go`: popup feature transport
- `cmd_interaction_*.go`: split interaction command surfaces by domain — browser/clipboard I/O, navigation/viewport work, detail-search follow-up, link and clipboard preparation, modal-editor execution, build-run loading, and forced refresh execution
- `cmd_detail_fold.go`: detail fold and inline-thread live-view sync
- `cmd_detail_motion.go`: detail/build-popup motion and pending-yank live-view sync
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
- Direct GitHub port calls in `internal/tui` are confined to explicit command or loading files such as `workflow_pull_request_detail_commands.go`, `workflow_notification_commands.go`, `program_loading.go`, `notification_loading.go`, and `notification_detail_loader.go`.

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
- `internal/tui/cmd_interaction_navigation.go`
- `internal/tui/cmd_interaction_link_clipboard.go`
- `internal/github/doc.go`
- `internal/githubcli/doc.go`
