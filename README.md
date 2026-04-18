# lazygh

`lazygh` is a Go CLI that aims to make GitHub pull request work less annoying with a lazygit-like TUI.

## Status
The repo is bootstrapped. The next milestone is a three-view MVP:
- view `0`: detail pane
- view `1`: connected user
- view `2`: pull requests with `My PRs` and `Requested` tabs

## Prerequisites
- `mise`
- Go `1.24.4` through `mise`
- `gh` for the later GitHub-backed milestones

## Run
```sh
mise run run
```

## Tasks
```sh
mise run test
mise run fmt
mise run tidy
```
