---
name: lazygh-todo-executor
description: Use when work in this repo references `.sandbox/todo-*.md`, `.sandbox/plans/*`, `start todo`, or an implementation plan that should be executed one phase at a time.
---

# Lazygh Todo Executor

## Overview
Default workflow for repo todo files and phase-based plans. The goal is to remove repeated boilerplate, not invent more of it.

## Use It For
- `.sandbox/todo-*.md`
- `.sandbox/plans/*`
- prompts like `start todo N`, `implement plan`, or `proceed with phase N`

## Workflow
1. Read the referenced todo or plan fully before changing code. Treat it as the acceptance spec.
2. Work on one todo or one plan phase at a time unless the user explicitly asks for batching.
3. Ask questions only for blockers or real ambiguities. Use `ask-user-question` with concrete options.
4. Start red: add or update the failing test first. Keep test names in BDD form.
5. Go green with the smallest change that satisfies the active step. Follow the package boundaries from `AGENTS.md`.
6. Verify in this order:
   - targeted test(s)
   - `mise run fmt`
   - `mise run test`
   - `mise run build`
7. After finishing the todo or phase:
   - update the referenced todo file
   - update `.sandbox/TODO.md`
   - make a conventional commit
8. Before claiming a todo update is committed, run `git check-ignore -v` on the touched `.sandbox` files. If they are ignored, tell the user immediately that the updates are local-only unless explicitly force-added or moved.
9. If the task touched rendering, layout, theme, popups, markdown, images, or cursor behavior, use `tui-visual-verifier` before closing the task.
10. If the task blocks on terminal or runtime behavior, leave a concise handoff in `.sandbox/handoffs/`.

## Planning-only work
If the task is planning-only, skip code commands, reread the changed files, and say that no code tests were run.

## Quick checks
- Did the test fail before the implementation?
- Did I update both the detailed todo and `.sandbox/TODO.md`?
- Did I verify whether `.sandbox` changes can actually be committed?
- Did I commit once the active todo or phase was done?
