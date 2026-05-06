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

## Install with `mise`
Install `lazygh` globally with `mise`'s Go backend:

```sh
mise use -g go:codeberg.org/l-lin/lazygh/cmd/lazygh@latest
lazygh
```

Run it once without a global install:

```sh
mise exec go:codeberg.org/l-lin/lazygh/cmd/lazygh@latest -- lazygh review https://github.com/acme/widgets/pull/42
```

## Run from this checkout
```sh
mise run run
```

## Install and use from this checkout
```sh
mise run install
mise run lazygh
mise run lazygh review https://github.com/acme/widgets/pull/42
```

## Tasks
```sh
mise run run
mise run install
mise run lazygh
mise run test
mise run fmt
mise run tidy
```

## Config
`lazygh` looks for `~/.config/lazygh/config.toml`.

If the file is missing, `lazygh` starts with the built-in defaults. If the TOML is malformed, startup fails. Unknown scopes, unknown actions, invalid key strings, invalid keymap value types, invalid theme colors, invalid story-review settings, invalid cache settings, and invalid pull-request search entries are ignored, because apparently survival is preferable to drama.

### Cache
Use `[cache]` to control the persistent SQLite cache.

- By default, `lazygh` stores the cache at `$XDG_DATA_HOME/lazygh/cache.sqlite3`.
- If `XDG_DATA_HOME` is unset, it falls back to `~/.local/share/lazygh/cache.sqlite3`.
- `lazygh` shows cached pull-request lists immediately, then refreshes the active list in the background.
- Cached PR detail and review diff entries refresh only when the live list reports a newer `updatedAt`, or when `lazygh` mutates that PR and invalidates the cached entry.

This example overrides the cache path.

```toml
[cache]
path = "/tmp/lazygh/cache.sqlite3"
```

### Themes
Configure palette overrides under `[theme]`.

- Every color key is optional.
- Values must use the `#RRGGBB` format.
- `markdown_heading_background` controls the full-line heading fill.
- `pull_request_reference` colors the `owner/repo#123` prefix in pull-request lists.
- `pull_request_title` colors the pull-request title text in pull-request lists.
- Missing or invalid colors fall back to the built-in palette.
- Bundled theme examples live in `themes/`. Copy one into `~/.config/lazygh/config.toml` and tweak it if your eyes demand it.

Bundled examples:
- `themes/catppuccin-latte.toml`
- `themes/catppuccin-frappe.toml`
- `themes/catppuccin-macchiato.toml`
- `themes/catppuccin-mocha.toml`
- `themes/kanagawa-dark.toml`
- `themes/kanagawa-light.toml`
- `themes/tokyonight-dark.toml`
- `themes/tokyonight-light.toml`
- `themes/gruvbox-dark.toml`
- `themes/gruvbox-light.toml`
- `themes/nord.toml`

This example overrides only a few colors.

```toml
[theme]
active_border = "#7E9CD8"
inactive_border = "#54546D"
selected_line_background = "#363646"
pull_request_reference = "#656D76"
pull_request_title = "#DCD7BA"
markdown_heading = "#7E9CD8"
markdown_heading_background = "#223249"
syntax_keyword = "#957FB8"
syntax_string = "#98BB6C"
diff_addition_background = "#2B3328"
diff_deletion_background = "#43242B"
```

### Links
Use the actions popup in pull-request detail, or press `gx` in any detail view, to open the link under the cursor.

- By default, `lazygh` uses `open` on macOS.
- By default, `lazygh` uses `xdg-open` on Linux.
- Override the opener under `[links]` when your desktop environment demands a different ritual.
- `open_command` can be a string or an array of strings. `lazygh` appends the resolved URL as the last argument.

This example opens links with Firefox on macOS.

```toml
[links]
open_command = ["open", "-a", "Firefox"]
```

### Story review
Use the actions popup on a pull request and pick `Review PR as story`.

- `lazygh` asks an external AI command to group changed files into review chapters.
- If `[story_review].agent_command` is missing, the action fails and tells you to configure it.
- If `[story_review].prompt` is missing, `lazygh` uses the built-in professional prompt.
- In story review mode, view `2` shows chapters with nested files. Selecting a chapter shows its narrative in view `0`. Selecting a file shows the diff again, because chaos has limits.

Configure the AI command under `[story_review]`.

- `agent_command` can be a string or an array of strings.
- Use `{{prompt_file}}` anywhere in the command to inject the generated prompt file path.
- If the command does not contain `{{prompt_file}}`, `lazygh` appends the prompt file path as the last argument.
- `prompt` is optional. It controls the tone and chaptering guidance that `lazygh` wraps around the PR metadata and diff.

This example uses `pi` and the built-in default prompt.

```toml
[story_review]
agent_command = ["pi", "--models", "anthropic/claude-sonnet-4-6", "--no-session", "-p", "@{{prompt_file}}"]
```

This example overrides the prompt.

```toml
[story_review]
agent_command = ["pi", "--models", "anthropic/claude-sonnet-4-6", "--no-session", "-p", "@{{prompt_file}}"]
prompt = """
Group the changes into a logical, reviewer-friendly story. Use a professional tone. Prefer chapters that reflect one cohesive behavior change, refactor step, or debugging thread. Explain what each chapter is doing, why it exists, and what a reviewer should mentally connect across the listed files. Keep the narrative concise, concrete, and useful for code review.
"""
```

Prompt examples live in `prompts/story-review/`:
- `prompts/story-review/default.md`
- `prompts/story-review/sanderson.md`
- `prompts/story-review/caveman.md`
- `prompts/story-review/emoji.md`

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

For multi-key motions, configure the prefix key once. `move_selection_to_top = "g"` and `move_cursor_to_top = "g"` make `gg` go to the top. `recenter_selection = "z"` makes `zz` recenter the selected row in side panes and the actions popup. In the detail pane, `toggle_inline_conversation_prefix = "z"` keeps `za` for inline conversations and also makes `zz` recenter the cursor. `page_down` and `page_up` move half a page and recenter on every supported view. In review mode on views `0` and `2`, `previous_tab` and `next_tab` become prefix keys. With the defaults, `[[` and `]]` move between files, and `[c` and `]c` move between comments.

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
# In review mode, `[[`/`]]` move between files and `[c`/`]c` move between comments.
previous_tab = "["
next_tab = "]"
copy_pull_request_url = "y"
comment_on_pull_request = "c"
open_actions_popup = "a"
# `gx` opens the link under the cursor, `za` toggles inline conversations in review mode, `zz` recenters the cursor.
toggle_inline_conversation_prefix = "z"
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
move_selection_to_top = "g"
move_selection_to_bottom = "G"
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
close = ["esc", "ctrl+[", "q"]
```

