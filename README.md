# lazygh

`lazygh` is a Go CLI that aims to make GitHub pull request work less annoying with a lazygit-like TUI.

## Prerequisites

- `mise`
- Go `1.25.9` through `mise`
- `gh` for the connected user view and the later GitHub-backed milestones

## Installation
### with `mise`

Install `lazygh` globally with `mise`'s Go backend:

```sh
mise use -g go:codeberg.org/l-lin/lazygh/cmd/lazygh@latest
lazygh
```

Run it once without a global install:

```sh
mise exec go:codeberg.org/l-lin/lazygh/cmd/lazygh@latest -- lazygh view https://github.com/acme/widgets/pull/42
mise exec go:codeberg.org/l-lin/lazygh/cmd/lazygh@latest -- lazygh review https://github.com/acme/widgets/pull/42
```

### Run from source

```sh
git clone https://codeberg.org/l-lin/lazygh/cmd/lazygh
mise run run
```

### Install and use from this checkout

```sh
mise run install
mise run lazygh
mise run lazygh view https://github.com/acme/widgets/pull/42
mise run lazygh review https://github.com/acme/widgets/pull/42
```

### Tasks

```sh
mise run run
mise run install
mise run lazygh
mise run test
mise run fmt
mise run tidy
```

Open a pull request directly in browser mode with `lazygh view <pr-url>`. Start review mode immediately with `lazygh review <pr-url>`.

## Config

`lazygh` looks for `~/.config/lazygh/config.toml`.

If the file is missing, `lazygh` starts with the built-in defaults. If the TOML is malformed, startup fails. Unknown scopes, unknown actions, invalid key strings, invalid keymap value types, invalid theme colors, invalid story-review settings, invalid cache settings, and invalid pull-request search entries are ignored, because apparently survival is preferable to drama.

### Cache

`lazygh` uses SQLite as cache storage to improve user experience.
Use `[cache]` to control the persistent SQLite cache.

- By default, `lazygh` stores the cache at `$XDG_DATA_HOME/lazygh/cache.sqlite3`.
- If `XDG_DATA_HOME` is unset, it falls back to `~/.local/share/lazygh/cache.sqlite3`.
- `lazygh` shows cached pull-request lists immediately, then refreshes the active list in the background.
- Cached PR detail and review diff entries refresh only when the live list reports a newer `updatedAt`, or when `lazygh` mutates that PR and invalidates the cached entry.

This example overrides the cache path:

```toml
[cache]
path = "/tmp/lazygh/cache.sqlite3"
```

### Themes

By default, it will use the `system` preset if not set:

```toml
[theme]
preset = "system"
```

List of presets can be found in [`preset.go`](internal/theme/preset.go). Available presets include `system`, `light`, and `dark`.

Use the `Change theme` action from the actions popup to switch presets at runtime. It updates `~/.config/lazygh/config.toml` immediately.

You can find the list of palette variables in [palette.go](internal/theme/palette.go). `background` fills the full TUI background.

This example starts from `kanagawa-dark` and overrides a few colors.

```toml
[theme]
preset = "kanagawa-dark"
# Values must use the `#RRGGBB` format.
background = "#1F1F28"
active_border = "#7E9CD8"
inactive_border = "#54546D"
selected_line_background = "#363646"
pull_request_reference = "#656D76"
pull_request_title = "#DCD7BA"
success = "#98BB6C"
success_background = "#2B3328"
failure = "#E46876"
failure_background = "#43242B"
pending = "#C8C093"
pending_background = "#363646"
muted = "#727169"
comment_author_badge = "#7E9CD8"
comment_author_badge_background = "#223249"
markdown_heading = "#7E9CD8"
markdown_heading_background = "#223249"
syntax_keyword = "#957FB8"
syntax_string = "#98BB6C"
pull_request_status_merged = "#957FB8"
pull_request_status_merged_background = "#252535"
diff_addition_highlight_background = "#35513B"
diff_deletion_highlight_background = "#5A2E35"
```

`pull_request_reference` colors the `owner/repo#123` prefix in pull-request lists.
`pull_request_title` colors the pull-request title text in pull-request lists.
`pull_request_status_*_background` also colors the `` status icon in pull-request lists.
`markdown_heading_background` controls the full-line heading fill in rendered markdown.
`success_background` and `failure_background` also fill pull-request rows in view 2 when the Merge Checks summary is fully passing or failing.

### Links

You can open a link either from the actions popup (default keymap `a`), or by pressing `gx`.

- By default, `lazygh` uses `open` on macOS.
- By default, `lazygh` uses `xdg-open` on Linux.

```toml
[links]
# Example opens links with Firefox on macOS.
# Can be a string or an array of strings. `lazygh` appends the resolved URL as the last argument.
open_command = ["open", "-a", "Firefox"]
```

### Actions popup

Press `a` to open the actions popup in the current context.

`Assign PR` opens a searchable assignee picker. Press `enter` to toggle an assignee, then press `alt+enter` to save. GitHub only allows up to 10 assignees per pull request, and the authenticated user still needs permission to assign users in that repository.

### Story review

You can start a PR review as a story, which groups changes by logical "chapters", to help you review.

It uses an external AI coding tool, like claude-code, pi, codex, etc...

To have this feature, configure the `story_review.agent_command`:
Configure the AI command under `[story_review]`:

```toml
[story_review]
agent_command = ["pi", "--models", "anthropic/claude-sonnet-4-6", "--no-session", "-p", "@{{prompt_file}}"]
```

By default, it will use the prompt in [prompt.go](internal/story/prompt.go).

You can override the prompt to use:

```toml
[story_review]
agent_command = ["pi", "--models", "anthropic/claude-sonnet-4-6", "--no-session", "-p", "@{{prompt_file}}"]
prompt = """
Group the changes into a logical, reviewer-friendly story. Use a professional tone. Prefer chapters that reflect one cohesive behavior change, refactor step, or debugging thread. Explain what each chapter is doing, why it exists, and what a reviewer should mentally connect across the listed files. Keep the narrative concise, concrete, and useful for code review.
"""
```

You can find some prompt examples in [`prompts/story-review/`](./prompts/story-review/).

### Pull request searches

You can customize your own pull request searches under `[[pull_requests.searches]]`.

```toml
[[pull_requests.searches]]
label = "My PRs"
command = ["search", "prs", "--author", "@me", "--state", "open", "--sort", "updated", "--order", "desc"]

[[pull_requests.searches]]
label = "My reviews"
command = ["search", "prs", "--reviewed-by", "@me", "--limit", "100", "--state", "open", "--sort", "updated", "--order", "desc"]

[[pull_requests.searches]]
label = "Requested"
command = ["search", "prs", "--review-requested", "@me", "--limit", "100", "--state", "open", "--sort", "updated", "--order", "desc"]

# New search
[[pull_requests.searches]]
label = "Escalated"
command = ["search", "prs", "--search", "label:escalated state:open", "--sort", "updated", "--order", "desc"]
```

### Keymap overrides

Use scoped tables under `[keymaps]`.

Across views, `ctrl-d`/`ctrl-u` move by half a page and `ctrl-f`/`ctrl-b` move by a full page. Text inputs keep `ctrl-b` and `ctrl-f` for cursor movement.

In browser mode, `zt`, `zz`, and `zb` place the selected row at the top, center, or bottom of the side pane. Use `za` for inline conversations. In view 0, `zt`/`zz`/`zb` place the cursor at the top/center/bottom. `zM` and `zR` close or open every fold in the current detail context.

```toml
[keymaps.global]
quit = "ctrl+c"
next_side_view = "tab"
previous_side_view = "shift+tab"

[keymaps.main]
toggle_help = "?"
focus_user_view = "1"
focus_pull_requests_view = "2"
focus_notifications_view = "3"
open_search = "/"
move_selection_down = ["j", "down"]
move_selection_up = ["k", "up"]
page_down = "ctrl+d"
page_up = "ctrl+u"
full_page_down = ["ctrl+f", "pagedown"]
full_page_up = ["ctrl+b", "pageup"]
grow_focused_pane = "+"
shrink_focused_pane = "-"

[keymaps.side]
next_side_view = "l"
previous_side_view = "h"
focus_detail_view = "0"
move_selection_to_top = "g"
move_selection_to_bottom = "G"
# `zt`/`zz`/`zb` place the selection at the top/center/bottom in side panes.
recenter_selection = "z"
exit_review_mode = ["esc", "ctrl+[", "q"]

[keymaps.user]
open_detail = "enter"
copy_pull_request_url = "y"

[keymaps.pull_requests]
# In review mode, `[[`/`]]` move between files and `[c`/`]c` move between comments.
previous_tab = "["
next_tab = "]"
open_detail = "enter"
copy_pull_request_url = "y"
comment_on_pull_request = "c"
open_actions_popup = "a"

[keymaps.notifications]
open_detail = "enter"
mark_notification_read = "r"
mark_notification_done = "d"
open_actions_popup = "a"

[keymaps.detail]
move_cursor_left = "h"
move_cursor_right = "l"
move_cursor_to_row_start = "0"
move_cursor_to_row_end = "$"
move_cursor_to_top = "g"
open_link_under_cursor = "x"
move_cursor_to_bottom = "G"
move_cursor_to_next_word = "w"
move_cursor_to_word_end = "e"
move_cursor_to_previous_word = "b"
next_search_match = "n"
previous_search_match = "N"
enter_visual_mode = "v"
enter_line_visual_mode = "V"
# In browser mode, `[` and `]` cycle `Description`, `Comments`, `Commits`, and `Changes`.
# In review mode, `[[`/`]]` move between files and `[c`/`]c` move between comments.
# On diff lines in review mode view `0`, `c` and the actions popup both add inline comments.
previous_tab = "["
next_tab = "]"
copy_pull_request_url = "y"
comment_on_pull_request = "c"
open_actions_popup = "a"
# `gx` opens the link under the cursor, `za` toggles inline conversations in review mode, `zM` and `zR` close or open every fold in the current detail context, and `zt`/`zz`/`zb` place the cursor at the top/center/bottom.
toggle_inline_conversation_prefix = "z"
close_all_folds = "M"
open_all_folds = "R"
close = ["esc", "ctrl+[", "q"]

[keymaps.search]
submit = ["enter", "ctrl+j"]
cancel = ["esc", "ctrl+["]

[keymaps.actions_popup]
focus_search = "/"
move_selection_down = ["j", "down"]
move_selection_up = ["k", "up"]
page_down = "ctrl+d"
page_up = "ctrl+u"
full_page_down = ["ctrl+f", "pagedown"]
full_page_up = ["ctrl+b", "pageup"]
move_selection_to_top = "g"
move_selection_to_bottom = "G"
# `zt`/`zz`/`zb` place the selection at the top/center/bottom in the popup.
recenter_selection = "z"
execute_selected_action = "enter"
close = ["esc", "ctrl+[", "q"]

[keymaps.actions_popup_search]
focus_list = ["enter", "tab"]
close = ["esc", "ctrl+["]

[keymaps.modal_editor]
submit = "alt+enter"
close = ["esc", "ctrl+["]

[keymaps.help]
full_page_down = ["ctrl+f", "pagedown"]
full_page_up = ["ctrl+b", "pageup"]
close = ["esc", "ctrl+[", "q"]
```

