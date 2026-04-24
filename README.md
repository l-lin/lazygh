# lazygh

`lazygh` is a Go CLI that aims to make GitHub pull request work less annoying with a lazygit-like TUI.

## Status
The repo now boots into a three-view TUI:
- view `0`: detail pane, including rich PR metadata, markdown body rendering, and comments from `gh pr view`
- view `1`: connected user from `gh api user`
- view `2`: pull requests from ordered, configurable `gh` searches, with tabs named from the config

The next milestones can focus on layout polish and extra PR actions.

## Prerequisites
- `mise`
- Go `1.25.9` through `mise`
- `gh` for the connected user view and the later GitHub-backed milestones

## Run
```sh
mise run run
```

## Tasks
```sh
mise run run
mise run test
mise run fmt
mise run tidy
```

## Config
`lazygh` looks for `~/.config/lazygh/config.toml`.

If the file is missing, `lazygh` starts with the built-in defaults. If the TOML is malformed, startup fails. Unknown scopes, unknown actions, invalid key strings, invalid keymap value types, and invalid pull-request search entries are ignored, because apparently survival is preferable to drama.

### Pull request searches
Configure ordered tabs under `[[pull_requests.searches]]`.

- The configured list fully defines the pull-request tabs.
- You can rename or remove the built-in `My PRs` and `Requested` tabs by replacing them in the list.
- You can add extra searches by appending more entries.
- If the list is missing or ends up empty after validation, `lazygh` falls back to the built-in `My PRs` and `Requested` tabs.

`lazygh` runs `gh` itself, so configure only the arguments after `gh`. A string value is split on whitespace. If one argument needs spaces, use an array instead. Each command must return a JSON array with the fields `title,number,repository,url,body,state,isDraft,updatedAt`.

This example renames the default tabs and adds a third one.

```toml
[[pull_requests.searches]]
label = "My PRs"
command = ["search", "prs", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"]

[[pull_requests.searches]]
label = "Requested"
command = ["search", "prs", "--review-requested", "@me", "--limit", "100", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"]

# New search
[[pull_requests.searches]]
label = "Escalated"
command = ["search", "prs", "--search", "label:escalated state:open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"]
```

### Keymap overrides
Use scoped tables under `[keymaps]`.

For multi-key motions, configure the prefix key. `move_selection_to_top = "g"` and `move_cursor_to_top = "g"` mean `gg` goes to the top.

This example mirrors the built-in defaults.

```toml
[keymaps.global]
quit = "ctrl+c"
next_side_view = "tab"
previous_side_view = "shift+tab"

[keymaps.main]
toggle_help = "?"
focus_user_view = "1"
focus_pull_requests_view = "2"
open_search = "/"
move_selection_down = ["j", "down"]
move_selection_up = ["k", "up"]
page_down = "ctrl+d"
page_up = "ctrl+u"
grow_focused_pane = "+"
shrink_focused_pane = "-"

[keymaps.side]
next_side_view = "l"
previous_side_view = "h"
focus_detail_view = "0"
move_selection_to_top = "g"
move_selection_to_bottom = "G"
exit_review_mode = ["esc", "ctrl+[", "q"]

[keymaps.user]
open_detail = "enter"
copy_pull_request_url = "y"

[keymaps.pull_requests]
previous_tab = "["
next_tab = "]"
open_detail = "enter"
copy_pull_request_url = "y"
comment_on_pull_request = "c"
open_actions_popup = "a"

[keymaps.detail]
move_cursor_left = "h"
move_cursor_right = "l"
move_cursor_to_row_start = "0"
move_cursor_to_row_end = "$"
move_cursor_to_top = "g"
move_cursor_to_bottom = "G"
move_cursor_to_next_word = "w"
move_cursor_to_word_end = "e"
move_cursor_to_previous_word = "b"
next_search_match = "n"
previous_search_match = "N"
enter_visual_mode = "v"
enter_line_visual_mode = "V"
previous_tab = "["
next_tab = "]"
copy_pull_request_url = "y"
comment_on_pull_request = "c"
open_actions_popup = "a"
close = ["esc", "ctrl+[", "q"]

[keymaps.search]
submit = ["enter", "ctrl+j"]
cancel = ["esc", "ctrl+["]

[keymaps.actions_popup]
focus_search = "/"
move_selection_down = ["j", "down"]
move_selection_up = ["k", "up"]
move_selection_to_top = "g"
move_selection_to_bottom = "G"
execute_selected_action = "enter"
close = ["esc", "ctrl+[", "q"]

[keymaps.actions_popup_search]
focus_list = ["enter", "tab"]
close = ["esc", "ctrl+["]

[keymaps.modal_editor]
submit = "alt+enter"
close = ["esc", "ctrl+["]

[keymaps.help]
close = ["esc", "ctrl+[", "q"]
```

