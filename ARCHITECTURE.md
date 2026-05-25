# Architecture

`lazygh` is a Go CLI with a lazygit-style TUI for GitHub pull request work. The codebase is split into a thin startup layer, provider-neutral domain models, a `gh` adapter, a persistent cache, and a TUI that derives screens from state.

## System map

```text
cmd/lazygh
  -> internal/config        load and resolve config
  -> internal/theme         apply palette
  -> internal/githubcli     build `gh`-backed adapters
  -> internal/tui           run the TUI

internal/tui
  -> internal/github        provider-neutral GitHub models
  -> internal/cache         persist lists, detail, and diff data
  -> internal/story         generate story reviews through an external agent
  -> internal/theme         read resolved colors
  -> internal/clipboard     talk to the system clipboard
```

## Package roles

| Package | Role |
| --- | --- |
| `cmd/lazygh` | Composition root. Parses startup commands, loads config, applies theme, wires dependencies, and starts the runner. |
| `internal/app` | Minimal app wrapper that runs the configured runner. |
| `internal/config` | Loads `config.toml`, resolves defaults, normalizes keymaps, links, cache paths, story-review settings, and pull request searches. |
| `internal/theme` | Owns palette presets and the resolved runtime color variables. |
| `internal/github` | Defines provider-neutral GitHub models, enums, errors, and URL parsing. |
| `internal/githubcli` | Adapts the `gh` CLI into focused services and domain-facing capability adapters. This is the only layer that should know transport details. |
| `internal/cache` | Persists provider-neutral pull request and notification data in SQLite. |
| `internal/story` | Builds prompts, shells out to an external agent, and normalizes story-review output. |
| `internal/tui` | Owns the terminal UI, view state, render pipeline, async orchestration, and user interactions. |
| `internal/clipboard` | Wraps clipboard access behind a small interface. |

## Startup flow

`cmd/lazygh/main.go` is the only real composition root. It parses the startup mode, loads config, resolves theme and cache settings, builds `internal/githubcli` adapters, and injects them into `tui.AppDeps`.

`internal/app` stays intentionally small. It only calls `Run()` on the configured runner.

The runner is `tui.Program`. `Program.Run()` creates the `gocui` GUI, configures global terminal settings, starts the loading spinner, registers keybindings, and enters the main loop.

## Domain and adapter split

The most important boundary in the repo is the split between `internal/github` and `internal/githubcli`.

`internal/github` defines the app's GitHub language. It holds pull request, review, notification, build, reaction, and session models without any `gh` command knowledge.

`internal/githubcli` is the `gh` adapter. It owns command formatting, REST and GraphQL transport helpers, parsing, assembly, and capability adapters that convert raw `gh` output into `internal/github` values.

That split keeps the TUI from depending on transport quirks. The TUI talks to interfaces from `internal/tui/deps.go`, not to shell commands.

## Persistent cache

`internal/cache` stores provider-neutral data in SQLite. It caches pull request lists, notifications, pull request detail, and pull request diff data.

The cache is optimistic but simple. The TUI shows cached data early, then refreshes live data in the background. Detail and diff entries refresh when the live summary moves forward or when a local mutation invalidates the cached entry.

Within `internal/tui`, the persistent cache is wrapped by store types such as `persistentCacheStore`, `pullRequestListStore`, `detailStore`, and `reviewStore`. Those stores combine persisted data with in-memory caches and in-flight tracking.

## TUI architecture

The TUI aims for a functional-core, imperative-shell split. It is closer than it used to be. It is not yet strict TEA.

### Core state

`internal/tui/model.go` defines `Model`, the durable UI state for focus, tabs, list rows, selected indexes, search state, pane layout, and actions-popup state.

`Model` is the closest thing to the functional core. It owns many state transitions directly, and the newer reducer layer routes more behavior through explicit messages.

### Runtime shell

`internal/tui/program.go` defines `Program`, the shell object. It still holds a lot: the model, query and mutation ports, caches, stores, async helpers, widget runtime state, image loading, status state, and the live `gocui.Gui`.

That makes `Program` the main coordinator. It is useful, but it is also the largest remaining architectural compromise.

### Messages, update, and commands

The TUI now has explicit `Msg`, `Update`, and `Cmd` types.

- `Msg` types describe events such as focus changes, search edits, popup requests, cache hydration, and async results.
- `Update` applies those messages to the live program state.
- `Cmd` values describe shell work such as loading pull requests, hydrating cached lists, loading detail and diff data, rendering markdown images, and running popup-triggered mutations.

`dispatch()` runs `Update`, executes returned commands, and then hands control to `afterStateChange()` for shell sync and redraw.

`dispatchAsync()` is the only production path that still talks to `uiUpdater.Apply(...)` directly. Worker goroutines use it to hop back onto the UI thread with typed result messages.

Popup and modal-editor feature files now stop at typed request messages. Update-owned command builders own the GitHub-facing work.

### Screen derivation

The TUI separates screen derivation from terminal application in a few layers.

- `ScreenState` describes the logical browser, review, or story-review state.
- `ScreenLayout` turns that logical state plus terminal size into frames.
- `screenComposition` binds frames to renderers.
- `applyScreenComposition()` applies the derived composition to `gocui`.

This split is one of the cleanest parts of the TUI. It makes panel routing, overlays, tabs, key-hint context, and layout decisions easier to test without a live terminal.

### Render pipeline

The render pipeline is mostly read-only now.

Renderers configure views and write content, but selector caching, list viewport cleanup, terminal image sync, and detail-view shell sync have been pushed out of `render.go` and into shell or selector helpers.

The detail pane uses selector-style document builders for:

- pull request detail documents
- conversation documents
- rendered diff rows
- review diff documents

Those selectors memoize expensive derived data outside the render entrypoints.

### Async workflow planning

The TUI plans background work after state changes. `plannedCommands()` inspects the current state and returns commands for cache hydration, connected-user loading, pull request lists, notifications, detail, diff data, and inline images.

The stores under `workflow_stores.go` track in-flight state, cache state, and invalidation state. They are shell-oriented coordinators, not pure reducers, but their model writes now land through typed cache-hydration messages instead of mutating `program.model` directly.

### Overlays and editors

Search, the actions popup, transient errors, the build popup, and the modal editor all live in `internal/tui`. The editor callbacks dispatch typed edit messages instead of redrawing their own views directly.

Modal editors now support typed submit-request messages. Popup and modal-editor feature files no longer call GitHub query or mutation ports directly. Commenting, review submissions, pull request edits, inline comment mutations, notification actions, and story-review startup now dispatch typed request messages and let update-owned commands talk to the ports.

That keeps one redraw path in charge of rendering. Popup-local error presentation is still centralized, but it now routes through reducer-owned messages instead of a legacy result struct. The legacy submit callback path remains only for local, non-GitHub editor flows such as custom search. The live editor objects still sit inside `Program` rather than inside a pure child model.

## Story review pipeline

Story review is a separate vertical slice.

`internal/story` builds a prompt from pull request metadata and diff items, writes it to a temporary file, shells out to a configured agent command, parses JSON output, and normalizes the resulting chapter list.

The TUI then turns that normalized story into `reviewStoryData`, which drives story mode in the left tree and detail pane.

## A typical read path

A normal read flow looks like this:

1. The user changes selection or focus.
2. The TUI dispatches a `Msg`.
3. `Update` mutates the model.
4. `afterStateChange()` asks `plannedCommands()` what data is now missing or stale.
5. The shell runs typed load commands through the injected ports.
6. Async results come back as typed result messages through `dispatchAsync()`.
7. The reducer updates caches and visible rows.
8. The render pipeline derives screen state, layout, and view content, then applies them to `gocui`.

## A typical mutation path

A normal mutation flow looks like this:

1. The user triggers an action from a keybinding, popup, or editor.
2. The feature surface dispatches a typed request message.
3. `Update` builds a command through the injected capability.
4. Long-running work runs in the async shell when needed.
5. Completion comes back as a typed result message, or a reducer-owned success hook runs on the UI thread.
6. `Update` applies optimistic success, rollback, invalidation, feedback, or popup state changes.
7. The shell redraws the derived screen.

That flow is much cleaner than it was. It still is not universal.

## Current architectural posture

The repo has clear boundaries, and they matter.

- GitHub shelling belongs in `internal/githubcli`.
- Provider-neutral models belong in `internal/github`.
- Palette values belong in `internal/theme`.
- Rendering belongs in `internal/tui`.
- Detail view `0` is a read-only detail pane.

The TUI is now TEA-inspired with real `Msg`, `Update`, and `Cmd` pieces. It is still not strict TEA because `Program` remains large, some explicit shell helpers and loader files still call GitHub ports directly, and a few shell-oriented coordinators still sit beside the reducer rather than inside it.

If you want to understand the project quickly, start with these files:

- `cmd/lazygh/main.go`
- `internal/tui/program.go`
- `internal/tui/model.go`
- `internal/tui/update.go`
- `internal/tui/screen_state.go`
- `internal/tui/render_pipeline.go`
- `internal/tui/deps.go`
- `internal/github/doc.go`
- `internal/githubcli/doc.go`
- `internal/cache/sqlite.go`
