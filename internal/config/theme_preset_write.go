package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

func SaveThemePresetDefault(preset string) error {
	homeDirectory, actualErr := os.UserHomeDir()
	if actualErr != nil {
		return actualErr
	}

	return SaveThemePreset(DefaultPath(homeDirectory, os.Getenv("XDG_CONFIG_HOME")), preset)
}

func SaveThemePreset(configPath string, preset string) error {
	normalizedPreset := theme.NormalizePresetName(preset)
	if normalizedPreset == "" {
		return fmt.Errorf("unknown theme preset %q", preset)
	}

	contents, actualErr := os.ReadFile(configPath)
	if actualErr != nil && !os.IsNotExist(actualErr) {
		return actualErr
	}

	updatedContents := upsertThemePresetSection(string(contents), normalizedPreset)
	if actualErr := os.MkdirAll(filepath.Dir(configPath), 0o755); actualErr != nil {
		return actualErr
	}
	return os.WriteFile(configPath, []byte(updatedContents), 0o644)
}

func upsertThemePresetSection(contents string, preset string) string {
	normalizedContents := strings.ReplaceAll(contents, "\r\n", "\n")
	section := themePresetSection(preset)
	start, end, ok := themeSectionRange(normalizedContents)
	if !ok {
		trimmedContents := strings.TrimRight(normalizedContents, "\n")
		if trimmedContents == "" {
			return section
		}
		return trimmedContents + "\n\n" + section
	}

	prefix := normalizedContents[:start]
	suffix := normalizedContents[end:]
	return joinConfigSections(prefix, section, suffix)
}

func themePresetSection(preset string) string {
	return fmt.Sprintf("[theme]\npreset = %q\n", preset)
}

func themeSectionRange(contents string) (int, int, bool) {
	if contents == "" {
		return 0, 0, false
	}

	offset := 0
	start := -1
	for _, line := range strings.SplitAfter(contents, "\n") {
		trimmedLine := strings.TrimSpace(strings.TrimSuffix(line, "\n"))
		if start < 0 {
			if trimmedLine == "[theme]" {
				start = offset
			}
			offset += len(line)
			continue
		}
		if isConfigSectionHeader(trimmedLine) {
			return start, offset, true
		}
		offset += len(line)
	}
	if start >= 0 {
		return start, len(contents), true
	}
	return 0, 0, false
}

func isConfigSectionHeader(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return strings.HasPrefix(trimmedLine, "[")
}

func joinConfigSections(prefix string, section string, suffix string) string {
	trimmedPrefix := strings.TrimRight(prefix, "\n")
	trimmedSuffix := strings.TrimLeft(suffix, "\n")
	switch {
	case trimmedPrefix == "" && trimmedSuffix == "":
		return section
	case trimmedPrefix == "":
		return section + "\n" + trimmedSuffix
	case trimmedSuffix == "":
		return trimmedPrefix + "\n\n" + section
	default:
		return trimmedPrefix + "\n\n" + section + "\n" + trimmedSuffix
	}
}
