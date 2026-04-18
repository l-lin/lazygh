# lazygh

`lazygh` is a Go CLI that aims to make GitHub pull request work less annoying with a lazygit-like TUI.

## Status
The repo now boots into a three-view dummy TUI:
- view `0`: detail pane
- view `1`: connected user
- view `2`: pull requests with `My PRs` and `Requested` tabs

The next milestone replaces dummy data with `gh` output.

## Prerequisites
- `mise`
- Go `1.25.9` through `mise`
- `gh` for the later GitHub-backed milestones

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
