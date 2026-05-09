package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectFiles_GivenMiseToml_WhenReadingTheTasks_ThenItDefinesInstallAndBinaryRunTasks(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "mise.toml"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"[tasks.install]",
		"go install ./cmd/lazygh",
		"[tasks.lazygh]",
		"./bin/lazygh",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected mise.toml to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheRepo_WhenInspectingThemeSources_ThenItDoesNotShipAThemesDirectory(t *testing.T) {
	_, actualErr := os.Stat(filepath.Join("..", "..", "themes"))
	if !os.IsNotExist(actualErr) {
		t.Fatalf("expected the themes directory to be absent, actual error %v", actualErr)
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheMiseSection_ThenItDocumentsGlobalAndRepoLocalUsage(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"mise use -g go:codeberg.org/l-lin/lazygh/cmd/lazygh@latest",
		"mise exec go:codeberg.org/l-lin/lazygh/cmd/lazygh@latest -- lazygh",
		"mise run install",
		"mise run lazygh",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheThemeSection_ThenItDocumentsTheBackgroundOverride(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"background",
		"fills the full TUI background",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheThemeSection_ThenItDocumentsTheThemePresetSelection(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"preset = \"system\"",
		"Available presets include `system`, `light`, and `dark`",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheConfigSection_ThenItDocumentsXDGConfigHomeBeforeTheHomeFallback(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"`$XDG_CONFIG_HOME/lazygh/config.toml`",
		"falls back to `~/.config/lazygh/config.toml`",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheThemeSection_ThenItDocumentsTheActionsPopupThemePicker(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"Change theme",
		"updates the resolved config file immediately",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheThemeSection_ThenItDocumentsTheMarkdownHeadingBackgroundOverride(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"markdown_heading_background",
		"controls the full-line heading fill",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheThemeSection_ThenItDocumentsThePullRequestReferenceOverride(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"pull_request_reference",
		"colors the `owner/repo#123` prefix in pull-request lists",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheThemeSection_ThenItDocumentsThePullRequestTitleOverride(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"pull_request_title",
		"colors the pull-request title text in pull-request lists",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheThemeSection_ThenItDocumentsThePullRequestStatusIconPaletteReuse(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"pull_request_status_*_background",
		"also colors the `` status icon in pull-request lists",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheThemeSection_ThenItDocumentsMergeCheckRowHighlights(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"`success_background` and `failure_background` also fill pull-request rows in view 2",
		"when the Merge Checks summary is fully passing or failing",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheKeymapSection_ThenItDocumentsFullPageNavigationAndTheTextInputException(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"full_page_down",
		"full_page_up",
		"`ctrl-d`/`ctrl-u`",
		"`ctrl-f`/`ctrl-b`",
		"Text inputs keep `ctrl-b` and `ctrl-f` for cursor movement",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheKeymapSection_ThenItDocumentsZTZZAndZB(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"`zt`, `zz`, and `zb` place the selected row",
		"`za` for inline conversations",
		"`zM` and `zR` close or open every fold in the current detail context",
		"`zt`/`zz`/`zb` place the cursor at the top/center/bottom",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheKeymapSection_ThenItDocumentsWordAndWORDMotions(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"`w`, `e`, and `b` follow vim word motions",
		"`W`, `E`, and `B` use whitespace-delimited `WORD` motions",
		"move_cursor_to_next_big_word = \"W\"",
		"move_cursor_to_big_word_end = \"E\"",
		"move_cursor_to_previous_big_word = \"B\"",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheKeymapSection_ThenItDocumentsBehaviorFirstSharedScopesTwoKeySequencesAndFixedViewShortcuts(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"single key like `\"q\"` or a two-key sequence like `\"za\"`",
		"shared behavior-first scopes",
		"`keymaps.global` covers actions that work across multiple panes",
		"`keymaps.global.previous_tab` and `keymaps.global.next_tab` cover tab switches",
		"`keymaps.global.next_side_view` and `keymaps.global.previous_side_view` share both the global and side-pane aliases",
		"first binding stays global and later bindings stay side-pane-only",
		"`keymaps.modal_editor.cancel` covers the modal editor",
		"help, detail, and actions popup search reuse shared scopes",
		"`0`, `1`, `2`, and `3` stay fixed",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheKeymapSection_ThenItDocumentsBehaviorFirstScopesAndLeftRightArrowsForCursorMotion(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"[keymaps.selection]",
		"[keymaps.cursor]",
		"move_cursor_left = [\"h\", \"left\"]",
		"move_cursor_right = [\"l\", \"right\"]",
		"close = [\"esc\", \"q\"]",
		"[keymaps.modal_editor]",
		"next_side_view = [\"tab\", \"l\"]",
		"previous_side_view = [\"shift+tab\", \"h\"]",
		"previous_tab = \"[\"",
		"next_tab = \"]\"",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheActionsSection_ThenItDocumentsTheAssigneePickerAndGitHubConstraints(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"`Assign PR` opens a searchable assignee picker",
		"Press `enter` to toggle an assignee, then press `alt+enter` to save.",
		"GitHub only allows up to 10 assignees per pull request",
		"permission to assign users in that repository",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheLinksSection_ThenItDocumentsTheOpenCommandAndGXShortcut(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"[links]",
		"open_command",
		"gx",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}

func TestProjectFiles_GivenTheReadme_WhenReadingTheStoryReviewSection_ThenItDocumentsClaudeCodeCodexAndOpencodeExamples(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"story mode",
		"claude-code",
		"codex",
		"opencode",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}
