package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l-lin/lazygh/internal/theme"
)

func TestSaveThemePreset_GivenMissingConfigFile_WhenSaving_ThenItCreatesTheThemeSection(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "lazygh", "config.toml")

	actualErr := SaveThemePreset(configPath, theme.DarkPresetName)

	then_noError(t, actualErr)
	actualContents, actualErr := os.ReadFile(configPath)
	then_noError(t, actualErr)
	if actual := string(actualContents); actual != "[theme]\npreset = \"dark\"\n" {
		t.Fatalf("expected config contents %q, actual %q", "[theme]\npreset = \"dark\"\n", actual)
	}
}

func TestSaveThemePreset_GivenExistingConfigWithoutAThemeSection_WhenSaving_ThenItAppendsTheThemeSection(t *testing.T) {
	configPath := given_configFile(t, `[keymaps.global]
quit = "q"
`)

	actualErr := SaveThemePreset(configPath, theme.DarkPresetName)

	then_noError(t, actualErr)
	actualContents, actualErr := os.ReadFile(configPath)
	then_noError(t, actualErr)
	expectedContents := `[keymaps.global]
quit = "q"

[theme]
preset = "dark"
`
	if actual := string(actualContents); actual != expectedContents {
		t.Fatalf("expected config contents %q, actual %q", expectedContents, actual)
	}
}

func TestSaveThemePreset_GivenExistingThemeSectionAndOtherSettings_WhenSaving_ThenItReplacesOnlyTheThemeSection(t *testing.T) {
	configPath := given_configFile(t, `[keymaps.global]
quit = "q"

[theme]
preset = "light"
active_border = "#123456"

[links]
open_command = ["open"]
`)

	actualErr := SaveThemePreset(configPath, "kanagawa-dark")

	then_noError(t, actualErr)
	actualContents, actualErr := os.ReadFile(configPath)
	then_noError(t, actualErr)
	expectedContents := `[keymaps.global]
quit = "q"

[theme]
preset = "kanagawa-dark"

[links]
open_command = ["open"]
`
	if actual := string(actualContents); actual != expectedContents {
		t.Fatalf("expected config contents %q, actual %q", expectedContents, actual)
	}
}

func TestSaveThemePreset_GivenAnInvalidPreset_WhenSaving_ThenItReturnsAnError(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")

	actualErr := SaveThemePreset(configPath, "solarized")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
}
