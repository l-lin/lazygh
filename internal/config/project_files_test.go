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
