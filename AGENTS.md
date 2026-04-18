# AGENTS.md

## Project
- Module path: `codeberg.org/l-lin/lazygh`
- Language: Go
- Task runner: `mise`
- Binary entry point: `cmd/lazygh`

## Repo rules
- Run repeated commands through `mise run ...`.
- Use TDD. Add or update a failing test before implementation.
- Keep test names in BDD form: `TestX_GivenY_WhenZ_ThenW`.
- Keep the MVP scoped to three views only.
- Treat view `0` as a read-only detail pane.
- Cycle focus only between side views `1` and `2`.
- Keep GitHub shelling in `internal/githubcli`.
- Keep rendering logic in `internal/tui`.
- Keep palette values in `internal/theme`.
- Prefer the standard library and small dependencies.
- Default TUI choice is `gocui` because it stays close to lazygit.

## AI guidance
- Update this file before adding repo-specific agent skills.
- Add a dedicated skill only if the rule is repeated often enough that `AGENTS.md` stops being precise.
