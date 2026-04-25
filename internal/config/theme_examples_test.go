package config

import (
	"path/filepath"
	"reflect"
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
		if reflect.DeepEqual(actual.ResolvedTheme(), theme.DefaultPalette()) {
			t.Fatalf("expected %q to change the default palette", fileName)
		}
	}
}
