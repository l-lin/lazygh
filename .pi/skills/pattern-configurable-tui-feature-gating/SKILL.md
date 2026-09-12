---
name: pattern-configurable-tui-feature-gating
description: Use when configuration controls whether an interactive TUI feature or pane is visible, navigable, and active.
disable-model-invocation: false
---

# Configurable TUI feature gating

Keep the feature pure, gate the shell.

1. Resolve one explicit configuration flag with a safe default, and ignore invalid values consistently with the host application.
2. Project the flag into screen state and layout so disabled views cannot remain active or leave broken numbering and focus.
3. Gate workflow planning, cache hydration, refresh commands, and background loading, not only rendering.
4. Register feature-specific keybindings, actions, and footer or help hints only when enabled.
5. Add deterministic tests for disabled, enabled, invalid, and stale-focus states, plus tests proving disabled mode performs no work.
6. Verify the user-visible surface in a live TTY with `tmux`, checking both the default-disabled and explicit-enabled configurations.

Keep the feature pure, gate the shell. Run the project’s formatter, test suite, and build before the live check; record any unavailable project command and its documented equivalent.
