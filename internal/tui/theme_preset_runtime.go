package tui

type themeCapabilitySnapshot struct {
	presetPersistenceAvailable bool
}

func (program *Program) currentThemeCapabilitySnapshot() themeCapabilitySnapshot {
	if program == nil {
		return themeCapabilitySnapshot{}
	}
	return themeCapabilitySnapshot{presetPersistenceAvailable: program.themePresetStore != nil}
}

type themePresetRuntime struct {
	saveThemePreset func(string) error
}

func newThemePresetRuntime(program *Program) themePresetRuntime {
	if program == nil || program.themePresetStore == nil {
		return themePresetRuntime{}
	}
	return themePresetRuntime{saveThemePreset: program.themePresetStore.SaveThemePreset}
}
