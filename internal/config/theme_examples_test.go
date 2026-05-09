package config

import (
	"reflect"
	"testing"

	"github.com/l-lin/lazygh/internal/theme"
)

func TestConfig_ResolvedTheme_GivenBundledPresetNames_WhenResolving_ThenEachPresetProducesItsBundledPalette(t *testing.T) {
	expectedPresetNames := []string{
		"catppuccin-frappe",
		"catppuccin-latte",
		"catppuccin-macchiato",
		"catppuccin-mocha",
		"gruvbox-dark",
		"gruvbox-light",
		"kanagawa-dark",
		"kanagawa-light",
		"nord",
		"tokyonight-dark",
		"tokyonight-light",
	}

	for _, presetName := range expectedPresetNames {
		actual, ok := theme.PresetOverrides(presetName)
		if !ok {
			t.Fatalf("expected %q to be available as a theme preset", presetName)
		}
		if reflect.DeepEqual(actual, theme.Palette{}) {
			t.Fatalf("expected %q to define at least one theme override", presetName)
		}
		if actual.BackgroundHex == "" {
			t.Fatalf("expected %q to define %q", presetName, "background")
		}
		if actual.MarkdownHeadingBackgroundHex == "" {
			t.Fatalf("expected %q to define %q", presetName, "markdown_heading_background")
		}
		if actual.PullRequestReferenceHex == "" {
			t.Fatalf("expected %q to define %q", presetName, "pull_request_reference")
		}
		if actual.PullRequestTitleHex == "" {
			t.Fatalf("expected %q to define %q", presetName, "pull_request_title")
		}
		for key, value := range map[string]string{
			"success":            actual.SuccessHex,
			"success_background": actual.SuccessBackgroundHex,
			"failure":            actual.FailureHex,
			"failure_background": actual.FailureBackgroundHex,
			"pending":            actual.PendingHex,
			"pending_background": actual.PendingBackgroundHex,
			"muted":              actual.MutedHex,
		} {
			if value == "" {
				t.Fatalf("expected %q to define %q", presetName, key)
			}
		}

		subject := Config{ThemePreset: presetName}
		resolved := subject.ResolvedTheme()
		if reflect.DeepEqual(resolved, theme.DefaultPalette()) {
			t.Fatalf("expected %q to change the default palette", presetName)
		}
		if !reflect.DeepEqual(resolved, theme.ResolvePaletteWithPreset(presetName, theme.Palette{})) {
			t.Fatalf("expected resolved preset %q to match the bundled palette", presetName)
		}
	}
}
