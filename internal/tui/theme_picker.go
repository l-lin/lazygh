package tui

import (
	"errors"
	"strings"

	"github.com/l-lin/lazygh/internal/theme"
)

const (
	themePickerActionTitle             = "Change theme"
	themePickerTitle                   = "Select theme"
	themePresetStoreUnavailableMessage = "theme preset store is unavailable"
)

type themePresetStore interface {
	SaveThemePreset(string) error
}

type defaultThemePresetStore struct {
	save func(string) error
}

type themePickerState struct{}

func (store *defaultThemePresetStore) SaveThemePreset(preset string) error {
	if store == nil || store.save == nil {
		return errors.New("theme preset store is unavailable")
	}
	return store.save(preset)
}

func (program *Program) themePickerVisible() bool {
	return program.actionsPopupWidget.themePicker != nil
}

func (program *Program) changeThemeActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:        "change-theme",
		title:     themePickerActionTitle,
		icon:      actionsPopupChangeThemeIcon,
		requested: MsgOpenThemePickerRequested{},
	}
}

func (program *Program) currentThemePickerActions() []actionsPopupAction {
	if !program.themePickerVisible() {
		return nil
	}

	presets := theme.Presets()
	actions := make([]actionsPopupAction, 0, len(presets))
	for _, preset := range presets {
		actions = append(actions, program.themePickerAction(preset))
	}
	return actions
}

func (program *Program) themePickerAction(preset theme.Preset) actionsPopupAction {
	normalizedName := theme.NormalizePresetName(preset.Name)
	themeCapabilities := program.currentThemeCapabilitySnapshot()
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	switch {
	case !program.themePickerVisible():
	case !themeCapabilities.presetPersistenceAvailable:
		requested = actionsPopupErrorRequested(errors.New(themePresetStoreUnavailableMessage))
	case normalizedName == "":
	default:
		requested = MsgThemePresetSelected{NormalizedName: normalizedName, Label: strings.TrimSpace(preset.Label)}
	}
	return actionsPopupAction{
		id:        "theme-" + normalizedName,
		title:     strings.TrimSpace(preset.Label),
		requested: requested,
	}
}
