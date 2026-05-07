package config

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestLoad_GivenBundledThemeExamples_WhenLoading_ThenEachExampleParsesIntoAResolvedPalette(t *testing.T) {
	expectedExampleFileNames := []string{
		"catppuccin-frappe.toml",
		"catppuccin-latte.toml",
		"catppuccin-macchiato.toml",
		"catppuccin-mocha.toml",
		"gruvbox-dark.toml",
		"gruvbox-light.toml",
		"kanagawa-dark.toml",
		"kanagawa-light.toml",
		"nord.toml",
		"tokyonight-dark.toml",
		"tokyonight-light.toml",
	}
	exampleDirectory := filepath.Join("..", "..", "themes")

	for _, fileName := range expectedExampleFileNames {
		configPath := filepath.Join(exampleDirectory, fileName)

		actual, actualErr := Load(configPath)

		then_noError(t, actualErr)
		if reflect.DeepEqual(actual.Theme, theme.Palette{}) {
			t.Fatalf("expected %q to define at least one theme override", fileName)
		}
		expectedPresetName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		expectedOverrides, ok := theme.PresetOverrides(expectedPresetName)
		if !ok {
			t.Fatalf("expected %q to be available as a theme preset", expectedPresetName)
		}
		if !reflect.DeepEqual(actual.Theme, expectedOverrides) {
			t.Fatalf("expected %q overrides %+v, actual %+v", fileName, expectedOverrides, actual.Theme)
		}
		if actual.Theme.BackgroundHex == "" {
			t.Fatalf("expected %q to define %q", fileName, "background")
		}
		if actual.Theme.MarkdownHeadingBackgroundHex == "" {
			t.Fatalf("expected %q to define %q", fileName, "markdown_heading_background")
		}
		if actual.Theme.PullRequestReferenceHex == "" {
			t.Fatalf("expected %q to define %q", fileName, "pull_request_reference")
		}
		if actual.Theme.PullRequestTitleHex == "" {
			t.Fatalf("expected %q to define %q", fileName, "pull_request_title")
		}
		for key, value := range map[string]string{
			"success":            actual.Theme.SuccessHex,
			"success_background": actual.Theme.SuccessBackgroundHex,
			"failure":            actual.Theme.FailureHex,
			"failure_background": actual.Theme.FailureBackgroundHex,
			"pending":            actual.Theme.PendingHex,
			"pending_background": actual.Theme.PendingBackgroundHex,
			"muted":              actual.Theme.MutedHex,
		} {
			if value == "" {
				t.Fatalf("expected %q to define %q", fileName, key)
			}
		}
		if reflect.DeepEqual(actual.ResolvedTheme(), theme.DefaultPalette()) {
			t.Fatalf("expected %q to change the default palette", fileName)
		}
	}
}
