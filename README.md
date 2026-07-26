# lazygh

`lazygh` is a Go CLI that aims to make GitHub pull request work less annoying with a lazygit-like TUI with vim-motion.

![PR list|500](./doc/pr_list.png)
![PR changes|500](./doc/pr_changes.png)
![PR comments|500](./doc/comments.png)
![PR review|500](./doc/review.png)
![PR story review|500](./doc/story_review.png)
![Change theme|500](./doc/change_theme.png)

Need the code map too? Read [ARCHITECTURE.md](./ARCHITECTURE.md).

## Prerequisites

- [`gh`](https://cli.github.com/) to connect to GitHub. Private image loading uses your `gh` auth session.
- A terminal with kitty graphics protocol support, such as Kitty or Ghostty, if you want inline images.
- If you run `lazygh` inside `tmux`, enable passthrough with `set -g allow-passthrough on`.
- [NerdFont](https://www.nerdfonts.com/) to get the icons.

## Getting started
### Installation with [`mise`](https://mise.jdx.dev/cli/)

Install `lazygh` globally with `mise`'s Go backend:

```sh
mise use -g go:github.com/l-lin/lazygh/cmd/lazygh@latest
lazygh
```

Or run it once without a global install:

```sh
mise exec go:github.com/l-lin/lazygh/cmd/lazygh@latest -- lazygh
```

### Directly from release page

You can download the binary directly from the [release page](https://github.com/l-lin/lazygh/releases) and use it.

### Run from source

```sh
git clone https://github.com/l-lin/lazygh/cmd/lazygh
mise run run
```

### Install and use from this checkout

```sh
mise run install
mise run lazygh
mise run lazygh view https://github.com/acme/widgets/pull/42
mise run lazygh review https://github.com/acme/widgets/pull/42
```

### Development tasks

```sh
mise run test
mise run coverage
mise run build
```

`mise run coverage` writes `.sandbox/coverage/coverage.out` and prints the repo-wide coverage summary.

### Releases

Tagged pushes that match `v*` publish release archives and `checksums.txt`.

Use `mise run release-check` to validate `.goreleaser.yaml`.
Use `mise run release-snapshot` to build the release artifacts locally without publishing them.

## Usage

```bash
# Show the installed version.
lazygh --version
# Show the CLI help.
lazygh --help
# Open TUI.
lazygh
# Directly view the PR details.
lazygh view https://github.com/acme/awesome/pull/123
# Show command-specific help.
lazygh review --help
lazygh story-review --help
# Directly start reviewing a PR.
lazygh review https://github.com/acme/awesome/pull/123
# Directly start a story review a PR.
# (Must configure `story_review.agent_command` beforehand)
lazygh story-review https://github.com/acme/awesome/pull/123
```

### Images

`lazygh` renders markdown images and supported HTML `<img>` tags in detail views, only if your terminal support [kitty image protocol](https://sw.kovidgoyal.net/kitty/graphics-protocol/).

It keeps small images at their natural size. It scales larger images down to fit the detail pane.

If inline graphics are unavailable, or an image download fails, `lazygh` still shows a visible `[Image: …]` caption and the resolved URL.

For private repositories, `lazygh` asks GitHub to render the markdown with repository context, then downloads the resolved image URL with your authenticated `gh` session when needed.

## Config

`lazygh` looks for `$XDG_CONFIG_HOME/lazygh/config.toml`. If `XDG_CONFIG_HOME` is unset, it falls back to `~/.config/lazygh/config.toml`.

If the file is missing, `lazygh` starts with the built-in defaults. If the TOML is malformed, startup fails. Unknown scopes, unknown actions, invalid key strings, invalid keymap value types, invalid theme colors, invalid display settings, invalid story-review settings, invalid cache settings, and invalid pull-request search entries are ignored, because apparently survival is preferable to drama.

### Display

Use `[display]` to control how repository names appear in browser list rows.

- By default, `lazygh` uses `owner_name`, which shows full references such as `acme/widgets#42`.
- Set `repository_style = "name"` to shorten list rows to `widgets#42`.
- This only affects browser list rows, including the pasted PR tab.
- Detail headers, detail metadata, URLs, cache keys, and GitHub commands keep full repository identities.

```toml
[display]
repository_style = "name"
```

Accepted values:

- `owner_name`, the default
- `name`

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

Available presets include `system`, `light`, `dark`, and `grey`, plus the additional presets listed in [`preset.go`](internal/theme/preset.go).

`grey` is a light preset inspired by `nvim-grey`. It keeps the background unset, flattens most syntax colors, and bolds keyword-like structure by default.

You can switch presets from inside the app too. `Change theme` in the actions popup updates the resolved config file immediately.

You can find the list of palette variables in [palette.go](internal/theme/palette.go).

`syntax_*_style` keys accept a list of `"bold"`, `"italic"`, and `"underline"`. Use `[]` to clear a preset's default syntax style for that bucket.

A few palette entries do more than their names suggest:

- `background` fills the full TUI background.
- `markdown_heading_background` controls the full-line heading fill.

- `pull_request_reference` colors the repository reference prefix in pull-request and notification lists, whether it is `owner/repo#123` or the shorter `repo#123` display form.
- `pull_request_title` colors the pull-request title text in pull-request lists.
- `pull_request_status_*_background` also colors the `` status icon in pull-request lists.
- `success_background` and `failure_background` also fill pull-request rows in view 2 when the Merge Checks summary is fully passing or failing.

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
warning = "#FFA066"
comment_author_badge = "#7E9CD8"
comment_author_badge_background = "#223249"
markdown_heading = "#7E9CD8"
markdown_heading_background = "#223249"
syntax_keyword = "#957FB8"
syntax_keyword_style = ["bold"]
syntax_string = "#98BB6C"
pull_request_status_merged = "#957FB8"
pull_request_status_merged_background = "#252535"
diff_addition_highlight_background = "#35513B"
diff_deletion_highlight_background = "#5A2E35"
```

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

### Story review

Story review powers `lazygh` story mode. It shells out to an external coding agent and expects the final answer to be JSON only.

Configure `story_review.agent_command` under `[story_review]`. `lazygh` writes the generated prompt to a temporary file and replaces `{{prompt_file}}` in the configured command. If your command does not include `{{prompt_file}}`, `lazygh` appends the prompt file path as the last argument.

Examples:

```toml
[story_review]
agent_command = ["pi", "--models", "anthropic/claude-sonnet-4-6", "--no-session", "-p", "@{{prompt_file}}"]
```

```toml
# Claude Code
[story_review]
agent_command = ["claude", "-p", "@{{prompt_file}}"]
```

By default, `lazygh` uses the prompt in [prompt.go](internal/story/prompt.go).

You can override the prompt too:

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
`lazygh` prepends `gh search prs` for you, so each entry only needs the flags.

```toml
[[pull_requests.searches]]
label = "My PRs"
flags = ["--author", "@me", "--state", "open", "--sort", "updated", "--order", "desc"]

[[pull_requests.searches]]
label = "My reviews"
flags = ["--reviewed-by", "@me", "--limit", "100", "--state", "open", "--sort", "updated", "--order", "desc"]

[[pull_requests.searches]]
label = "Requested"
flags = ["--review-requested", "@me", "--limit", "100", "--state", "open", "--sort", "updated", "--order", "desc"]

[[pull_requests.searches]]
label = "Escalated"
flags = ["--search", "label:escalated state:open", "--sort", "updated", "--order", "desc"]
```

In view `2`, press `:` or choose `Custom search` from the actions popup. Submitting it creates or updates the `Custom` tab.

Press `Ctrl+V` to read a GitHub pull request URL from the clipboard and open it directly in fullscreen view `0`.
Choose `Open PR from URL` from the actions popup when you want to type or paste the URL manually.

### Keymap overrides

Use scoped tables under `[keymaps]`.

A keymap value can be a single key like `"q"`, a modified key like `"alt+y"`, or a two-key sequence like `"za"`. Arrays still let you keep multiple alternatives.

You can find the default keymaps in [default_keymaps.toml](internal/config/default_keymaps.toml)
