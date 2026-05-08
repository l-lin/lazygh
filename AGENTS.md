# AGENTS.md

## Project
- Module path: `codeberg.org/l-lin/lazygh`
- Language: Go
- Go version: `1.25.x`
- Task runner: `mise`
- Binary entry point: `cmd/lazygh`

## Repo rules
- Run repeated commands through `mise run ...`.
- When a task references `.sandbox/todo-*.md` or `.sandbox/plans/*`, read that file first and treat it as the acceptance spec.
- Implement one todo or one plan phase at a time unless the user explicitly asks for batching.
- Use TDD. Add or update a failing test before implementation.
- Keep test names in BDD form: `TestX_GivenY_WhenZ_ThenW`.
- After each completed todo or phase, run `mise run fmt`, `mise run test`, and `mise run build`, unless the task is planning-only or the user says otherwise.
- After each completed todo or phase, update the referenced todo file, update `.sandbox/TODO.md`, and make a conventional commit, unless the task is planning-only or the user says otherwise.
- Before promising `.sandbox` updates are part of a commit, run `git check-ignore -v` on the touched files. If they are ignored, tell the user immediately that those updates are local-only unless explicitly force-added or moved.
- Scope work to the active todo, plan, and latest user clarification. Do not apply older MVP assumptions when newer project plans supersede them.
- Treat view `0` as a read-only detail pane.
- Keep GitHub shelling in `internal/githubcli`.
- Keep rendering logic in `internal/tui`.
- Keep palette values in `internal/theme`.
- Keep user-facing icons in `internal/tui/icons.go`.
- Prefer the standard library and small dependencies.
- Default TUI choice is `github.com/jesseduffield/gocui` because it stays close to lazygit.

## TUI verification
- For rendering, theme, popup, cursor, markdown, image, search-highlight, status-line, or layout changes, headless tests are required but not sufficient.
- Add deterministic tests first, then do one live TTY verification when feasible.
- Use `tmux` when verification needs a persistent terminal, real keypresses, or other TTY-only behavior.
- Do not claim pixel-graphics verification from `tmux capture-pane`; it only proves text fallbacks.

## Repo skills
- Project skills live in `.pi/skills/`.
- Use `lazygh-todo-executor` for tasks that reference `.sandbox/todo-*.md`, `.sandbox/plans/*`, `start todo`, or implementation plans.
- Use `tui-visual-verifier` for rendering, theme, popup, cursor, markdown, image, or other TUI-acceptance work.
- Keep durable repo-wide rules in `AGENTS.md`, and keep repeatable task workflows in skills.

## AI guidance
- Update this file before adding repo-specific agent skills.
- Add a dedicated skill only if the rule is repeated often enough that `AGENTS.md` stops being precise.
