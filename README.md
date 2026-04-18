# lazygh

`lazygh` is a Go CLI that aims to make GitHub pull request work less annoying with a lazygit-like TUI.

## Status
The repo now boots into a three-view TUI:
- view `0`: detail pane
- view `1`: connected user from `gh api user`
- view `2`: dummy pull requests with `My PRs` and `Requested` tabs

The next milestone replaces the pull request dummy data with `gh` output.

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
